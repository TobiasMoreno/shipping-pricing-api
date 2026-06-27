package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/TobiasMoreno/shipping-pricing-api/internal/domain"
	"github.com/TobiasMoreno/shipping-pricing-api/internal/handlers"
	"github.com/TobiasMoreno/shipping-pricing-api/internal/server"
)

// stubQuoteService returns a canned quote or error.
type stubQuoteService struct {
	quote domain.Quote
	err   error
}

func (s stubQuoteService) Quote(context.Context, domain.QuoteRequest, time.Time) (domain.Quote, error) {
	return s.quote, s.err
}

func sampleQuote(t *testing.T) domain.Quote {
	t.Helper()
	rule := &domain.PricingRule{
		ShippingType: domain.ShippingStandard, OriginZoneCode: "CORDOBA_CAPITAL", DestinationZoneCode: "CABA",
		BasePrice: 5000, PricePerKm: 4, PricePerKg: 1000,
		MaxWeightKg: 30, MinDistanceKm: 0, MaxDistanceKm: 1500, IsActive: true,
	}
	req := domain.QuoteRequest{
		OriginZoneCode: "CORDOBA_CAPITAL", DestinationZoneCode: "CABA",
		DistanceKm: 695, WeightKg: 2.5, ShippingType: domain.ShippingStandard, Priority: domain.PriorityNormal,
	}
	q, err := domain.Calculate(req, rule, domain.Zone{Code: "CORDOBA_CAPITAL", IsActive: true}, domain.Zone{Code: "CABA", IsActive: true}, nil, time.Now())
	if err != nil {
		t.Fatalf("building sample quote: %v", err)
	}
	return q
}

func router(svc handlers.QuoteService) http.Handler {
	return server.NewRouter(handlers.NewHealthRegistry(), handlers.NewShippingHandler(svc))
}

const validBody = `{
  "origin_zip_code": "5000",
  "destination_zip_code": "1405",
  "origin_zone_code": "CORDOBA_CAPITAL",
  "destination_zone_code": "CABA",
  "distance_km": 695,
  "weight_kg": 2.5,
  "shipping_type": "standard",
  "priority": "normal"
}`

func do(t *testing.T, h http.Handler, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/shipping/quote", strings.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestQuoteHandler_Success(t *testing.T) {
	h := router(stubQuoteService{quote: sampleQuote(t)})
	rec := do(t, h, validBody, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Price         int64    `json:"price"`
		Cached        bool     `json:"cached"`
		DecisionTrace []string `json:"decision_trace"`
		Breakdown     struct {
			BasePrice           int64 `json:"base_price"`
			DistanceFee         int64 `json:"distance_fee"`
			WeightFee           int64 `json:"weight_fee"`
			PriorityFee         int64 `json:"priority_fee"`
			ExpressFee          int64 `json:"express_fee"`
			Discount            int64 `json:"discount"`
			TotalBeforeDiscount int64 `json:"total_before_discount"`
			Total               int64 `json:"total"`
		} `json:"breakdown"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp.Cached {
		t.Error("cached = true, want false")
	}
	if len(resp.DecisionTrace) == 0 {
		t.Error("decision_trace is empty")
	}
	b := resp.Breakdown
	if b.BasePrice+b.DistanceFee+b.WeightFee+b.ExpressFee+b.PriorityFee != b.TotalBeforeDiscount {
		t.Error("breakdown components do not reconcile to total_before_discount")
	}
	if b.Total != b.TotalBeforeDiscount-b.Discount || b.Total != resp.Price {
		t.Error("total does not reconcile with price/discount")
	}
}

func TestQuoteHandler_MalformedJSON(t *testing.T) {
	h := router(stubQuoteService{quote: sampleQuote(t)})
	rec := do(t, h, "not json", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if code := errorCode(t, rec); code != "invalid_request" {
		t.Errorf("code = %q, want invalid_request", code)
	}
}

func TestQuoteHandler_NonPositiveWeight(t *testing.T) {
	h := router(stubQuoteService{quote: sampleQuote(t)})
	body := strings.Replace(validBody, `"weight_kg": 2.5`, `"weight_kg": 0`, 1)
	rec := do(t, h, body, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "weight_kg") {
		t.Errorf("details do not reference weight_kg: %s", rec.Body.String())
	}
}

func TestQuoteHandler_InvalidShippingType(t *testing.T) {
	h := router(stubQuoteService{quote: sampleQuote(t)})
	body := strings.Replace(validBody, `"shipping_type": "standard"`, `"shipping_type": "drone"`, 1)
	rec := do(t, h, body, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestQuoteHandler_BusinessErrorsReturn422(t *testing.T) {
	cases := map[string]error{
		"no rule":       domain.ErrNoApplicableRule,
		"zone inactive": domain.ErrZoneInactive,
		"promo invalid": domain.ErrPromotionInvalid,
	}
	for name, svcErr := range cases {
		t.Run(name, func(t *testing.T) {
			h := router(stubQuoteService{err: svcErr})
			rec := do(t, h, validBody, nil)
			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422", rec.Code)
			}
		})
	}
}

func TestQuoteHandler_ErrorEnvelopeHasRequestID(t *testing.T) {
	h := router(stubQuoteService{quote: sampleQuote(t)})
	rec := do(t, h, "not json", nil)

	var env struct {
		Error struct {
			Code      string `json:"code"`
			Message   string `json:"message"`
			RequestID string `json:"request_id"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error.Code == "" || env.Error.Message == "" || env.Error.RequestID == "" {
		t.Errorf("envelope missing fields: %+v", env.Error)
	}
}

func TestQuoteHandler_ReusesIncomingRequestID(t *testing.T) {
	h := router(stubQuoteService{quote: sampleQuote(t)})
	const id = "req_test_123"
	rec := do(t, h, validBody, map[string]string{"X-Request-ID": id})

	if got := rec.Header().Get("X-Request-ID"); got != id {
		t.Errorf("X-Request-ID header = %q, want %q", got, id)
	}
}

func errorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var env struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	return env.Error.Code
}
