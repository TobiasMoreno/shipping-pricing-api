package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TobiasMoreno/shipping-pricing-api/internal/domain"
	"github.com/TobiasMoreno/shipping-pricing-api/internal/services"
)

var now = time.Date(2026, time.June, 27, 12, 0, 0, 0, time.UTC)

// --- fakes ---

type fakeZones map[string]domain.Zone

func (f fakeZones) GetByCode(_ context.Context, code string) (domain.Zone, error) {
	z, ok := f[code]
	if !ok {
		return domain.Zone{}, services.ErrZoneNotFound
	}
	return z, nil
}

type fakeRules []domain.PricingRule

func (f fakeRules) FindActiveRules(_ context.Context, st domain.ShippingType, origin, dest string) ([]domain.PricingRule, error) {
	var out []domain.PricingRule
	for _, r := range f {
		if r.IsActive && r.ShippingType == st &&
			(r.OriginZoneCode == origin || r.OriginZoneCode == "*") &&
			(r.DestinationZoneCode == dest || r.DestinationZoneCode == "*") {
			out = append(out, r)
		}
	}
	return out, nil
}

type fakePromos map[string]domain.Promotion

func (f fakePromos) GetByCode(_ context.Context, code string) (domain.Promotion, error) {
	p, ok := f[code]
	if !ok {
		return domain.Promotion{}, services.ErrPromotionNotFound
	}
	return p, nil
}

// --- helpers ---

func activeZones() fakeZones {
	return fakeZones{
		"CORDOBA_CAPITAL": {Code: "CORDOBA_CAPITAL", IsActive: true},
		"CABA":            {Code: "CABA", IsActive: true},
	}
}

func baseReq() domain.QuoteRequest {
	return domain.QuoteRequest{
		OriginZipCode: "5000", DestinationZipCode: "1405",
		OriginZoneCode: "CORDOBA_CAPITAL", DestinationZoneCode: "CABA",
		DistanceKm: 695, WeightKg: 2.5,
		ShippingType: domain.ShippingStandard, Priority: domain.PriorityNormal,
	}
}

func exactRule() domain.PricingRule {
	return domain.PricingRule{
		ShippingType: domain.ShippingStandard, OriginZoneCode: "CORDOBA_CAPITAL", DestinationZoneCode: "CABA",
		BasePrice: 5000, PricePerKm: 4, PricePerKg: 1000,
		MaxWeightKg: 30, MinDistanceKm: 0, MaxDistanceKm: 1500, IsActive: true,
	}
}

func wildcardRule() domain.PricingRule {
	return domain.PricingRule{
		ShippingType: domain.ShippingStandard, OriginZoneCode: "*", DestinationZoneCode: "*",
		BasePrice: 9000, PricePerKm: 5, PricePerKg: 1200,
		MaxWeightKg: 40, MinDistanceKm: 0, MaxDistanceKm: 5000, IsActive: true,
	}
}

func newService(zones services.ZoneRepository, rules services.PricingRuleRepository, promos services.PromotionRepository) *services.QuoteService {
	return services.NewQuoteService(rules, zones, promos)
}

// --- tests ---

func TestQuote_Success(t *testing.T) {
	svc := newService(activeZones(), fakeRules{exactRule()}, fakePromos{})
	q, err := svc.Quote(context.Background(), baseReq(), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Price <= 0 {
		t.Errorf("price = %d, want > 0", q.Price.Cents())
	}
	if q.EstimatedDeliveryDays != 3 {
		t.Errorf("eta = %d, want 3 for standard", q.EstimatedDeliveryDays)
	}
}

func TestQuote_ExactRulePreferredOverWildcard(t *testing.T) {
	svc := newService(activeZones(), fakeRules{wildcardRule(), exactRule()}, fakePromos{})
	q, err := svc.Quote(context.Background(), baseReq(), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Breakdown.BasePrice != 5000 {
		t.Errorf("base price = %d, want 5000 (exact rule)", q.Breakdown.BasePrice.Cents())
	}
}

func TestQuote_WildcardFallback(t *testing.T) {
	svc := newService(activeZones(), fakeRules{wildcardRule()}, fakePromos{})
	q, err := svc.Quote(context.Background(), baseReq(), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Breakdown.BasePrice != 9000 {
		t.Errorf("base price = %d, want 9000 (wildcard rule)", q.Breakdown.BasePrice.Cents())
	}
}

func TestQuote_NoApplicableRule(t *testing.T) {
	svc := newService(activeZones(), fakeRules{}, fakePromos{})
	if _, err := svc.Quote(context.Background(), baseReq(), now); !errors.Is(err, domain.ErrNoApplicableRule) {
		t.Errorf("error = %v, want ErrNoApplicableRule", err)
	}
}

func TestQuote_InactiveZone(t *testing.T) {
	zones := activeZones()
	zones["CABA"] = domain.Zone{Code: "CABA", IsActive: false}
	svc := newService(zones, fakeRules{exactRule()}, fakePromos{})
	if _, err := svc.Quote(context.Background(), baseReq(), now); !errors.Is(err, domain.ErrZoneInactive) {
		t.Errorf("error = %v, want ErrZoneInactive", err)
	}
}

func TestQuote_ZoneNotFound(t *testing.T) {
	svc := newService(fakeZones{"CORDOBA_CAPITAL": {Code: "CORDOBA_CAPITAL", IsActive: true}}, fakeRules{exactRule()}, fakePromos{})
	if _, err := svc.Quote(context.Background(), baseReq(), now); !errors.Is(err, services.ErrZoneNotFound) {
		t.Errorf("error = %v, want ErrZoneNotFound", err)
	}
}

func TestQuote_ExpiredPromotion(t *testing.T) {
	promos := fakePromos{"OLD": {
		Code: "OLD", DiscountType: domain.DiscountPercentage, DiscountValue: 10,
		IsActive: true, EndsAt: now.Add(-time.Hour),
	}}
	req := baseReq()
	req.PromotionCode = "OLD"
	svc := newService(activeZones(), fakeRules{exactRule()}, promos)
	if _, err := svc.Quote(context.Background(), req, now); !errors.Is(err, domain.ErrPromotionInvalid) {
		t.Errorf("error = %v, want ErrPromotionInvalid", err)
	}
}

func TestQuote_PromotionNotFound(t *testing.T) {
	req := baseReq()
	req.PromotionCode = "DOESNOTEXIST"
	svc := newService(activeZones(), fakeRules{exactRule()}, fakePromos{})
	if _, err := svc.Quote(context.Background(), req, now); !errors.Is(err, services.ErrPromotionNotFound) {
		t.Errorf("error = %v, want ErrPromotionNotFound", err)
	}
}
