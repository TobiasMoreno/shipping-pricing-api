package handlers

import "github.com/TobiasMoreno/shipping-pricing-api/internal/domain"

// fieldError describes a single validation failure for the error envelope.
type fieldError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

// validateQuoteRequest checks the request format and field-level constraints,
// returning all violations found (empty slice means valid).
func validateQuoteRequest(r quoteRequestDTO) []fieldError {
	var errs []fieldError

	if r.OriginZipCode == "" {
		errs = append(errs, fieldError{"origin_zip_code", "is required"})
	}
	if r.DestinationZipCode == "" {
		errs = append(errs, fieldError{"destination_zip_code", "is required"})
	}
	if r.OriginZoneCode == "" {
		errs = append(errs, fieldError{"origin_zone_code", "is required"})
	}
	if r.DestinationZoneCode == "" {
		errs = append(errs, fieldError{"destination_zone_code", "is required"})
	}
	if r.DistanceKm <= 0 {
		errs = append(errs, fieldError{"distance_km", "must be greater than 0"})
	}
	if r.WeightKg <= 0 {
		errs = append(errs, fieldError{"weight_kg", "must be greater than 0"})
	}
	if r.Dimensions != nil {
		if r.Dimensions.HeightCm <= 0 || r.Dimensions.WidthCm <= 0 || r.Dimensions.LengthCm <= 0 {
			errs = append(errs, fieldError{"dimensions", "all informed dimensions must be greater than 0"})
		}
	}
	if !validShippingType(r.ShippingType) {
		errs = append(errs, fieldError{"shipping_type", "must be standard, express or same_day"})
	}
	if r.Priority != "" && !validPriority(r.Priority) {
		errs = append(errs, fieldError{"priority", "must be normal or high"})
	}

	return errs
}

func validShippingType(s string) bool {
	switch domain.ShippingType(s) {
	case domain.ShippingStandard, domain.ShippingExpress, domain.ShippingSameDay:
		return true
	default:
		return false
	}
}

func validPriority(s string) bool {
	switch domain.Priority(s) {
	case domain.PriorityNormal, domain.PriorityHigh:
		return true
	default:
		return false
	}
}
