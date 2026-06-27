package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/TobiasMoreno/shipping-pricing-api/internal/domain"
	"github.com/TobiasMoreno/shipping-pricing-api/internal/services"
)

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	Details   []fieldError `json:"details,omitempty"`
	RequestID string       `json:"request_id"`
}

// writeError writes a consistent error envelope, populating request_id from the
// request context.
func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string, details []fieldError) {
	writeJSON(w, status, errorEnvelope{Error: errorBody{
		Code:      code,
		Message:   message,
		Details:   details,
		RequestID: middleware.GetReqID(r.Context()),
	}})
}

// writeServiceError maps a domain or repository error to the appropriate HTTP
// response. Business failures become 422; anything unexpected becomes 500.
func writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNoApplicableRule):
		writeError(w, r, http.StatusUnprocessableEntity, "no_applicable_rule", "no pricing rule applies to this request", nil)
	case errors.Is(err, domain.ErrWeightExceeded):
		writeError(w, r, http.StatusUnprocessableEntity, "weight_exceeded", err.Error(), nil)
	case errors.Is(err, domain.ErrDistanceOutOfRange):
		writeError(w, r, http.StatusUnprocessableEntity, "distance_out_of_range", err.Error(), nil)
	case errors.Is(err, domain.ErrZoneInactive), errors.Is(err, services.ErrZoneNotFound):
		writeError(w, r, http.StatusUnprocessableEntity, "zone_unavailable", err.Error(), nil)
	case errors.Is(err, domain.ErrPromotionInvalid), errors.Is(err, services.ErrPromotionNotFound):
		writeError(w, r, http.StatusUnprocessableEntity, "promotion_invalid", err.Error(), nil)
	default:
		writeError(w, r, http.StatusInternalServerError, "internal_error", "an unexpected error occurred", nil)
	}
}
