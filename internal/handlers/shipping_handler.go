package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/TobiasMoreno/shipping-pricing-api/internal/domain"
)

// QuoteService is the use case the handler depends on (defined here, on the
// consumer side, so it can be mocked in tests).
type QuoteService interface {
	Quote(ctx context.Context, req domain.QuoteRequest, now time.Time) (domain.Quote, error)
}

// ShippingHandler exposes the shipping HTTP endpoints.
type ShippingHandler struct {
	quotes QuoteService
	now    func() time.Time
}

// NewShippingHandler builds the handler with its service dependency.
func NewShippingHandler(quotes QuoteService) *ShippingHandler {
	return &ShippingHandler{quotes: quotes, now: time.Now}
}

// Quote handles POST /shipping/quote.
func (h *ShippingHandler) Quote(w http.ResponseWriter, r *http.Request) {
	var req quoteRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "invalid request body", nil)
		return
	}

	if errs := validateQuoteRequest(req); len(errs) > 0 {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "request validation failed", errs)
		return
	}

	quote, err := h.quotes.Quote(r.Context(), req.toDomain(), h.now())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, newQuoteResponse(quote, false))
}

// newQuoteID returns a random opaque identifier for a quote.
func newQuoteID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return "q_" + hex.EncodeToString(b)
}
