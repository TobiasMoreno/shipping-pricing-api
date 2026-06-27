package handlers

import "testing"

// validators_test.go is in package handlers (not handlers_test) so it can test
// the unexported validator directly.

func validDTO() quoteRequestDTO {
	return quoteRequestDTO{
		OriginZipCode: "5000", DestinationZipCode: "1405",
		OriginZoneCode: "CORDOBA_CAPITAL", DestinationZoneCode: "CABA",
		DistanceKm: 695, WeightKg: 2.5,
		ShippingType: "standard", Priority: "normal",
	}
}

func TestValidateQuoteRequest_Valid(t *testing.T) {
	if errs := validateQuoteRequest(validDTO()); len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateQuoteRequest_Violations(t *testing.T) {
	tests := []struct {
		name  string
		mut   func(*quoteRequestDTO)
		field string
	}{
		{"missing origin zip", func(d *quoteRequestDTO) { d.OriginZipCode = "" }, "origin_zip_code"},
		{"missing destination zone", func(d *quoteRequestDTO) { d.DestinationZoneCode = "" }, "destination_zone_code"},
		{"zero distance", func(d *quoteRequestDTO) { d.DistanceKm = 0 }, "distance_km"},
		{"negative weight", func(d *quoteRequestDTO) { d.WeightKg = -1 }, "weight_kg"},
		{"invalid shipping type", func(d *quoteRequestDTO) { d.ShippingType = "drone" }, "shipping_type"},
		{"invalid priority", func(d *quoteRequestDTO) { d.Priority = "urgent" }, "priority"},
		{"non-positive dimension", func(d *quoteRequestDTO) {
			d.Dimensions = &dimensionsDTO{HeightCm: 10, WidthCm: 0, LengthCm: 5}
		}, "dimensions"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := validDTO()
			tt.mut(&d)
			errs := validateQuoteRequest(d)
			if !hasField(errs, tt.field) {
				t.Errorf("expected a violation on %q, got %v", tt.field, errs)
			}
		})
	}
}

func TestValidateQuoteRequest_PriorityOptional(t *testing.T) {
	d := validDTO()
	d.Priority = ""
	if errs := validateQuoteRequest(d); len(errs) != 0 {
		t.Errorf("empty priority should be valid, got %v", errs)
	}
}

func hasField(errs []fieldError, field string) bool {
	for _, e := range errs {
		if e.Field == field {
			return true
		}
	}
	return false
}
