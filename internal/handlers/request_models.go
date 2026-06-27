package handlers

import "github.com/TobiasMoreno/shipping-pricing-api/internal/domain"

type dimensionsDTO struct {
	HeightCm float64 `json:"height_cm"`
	WidthCm  float64 `json:"width_cm"`
	LengthCm float64 `json:"length_cm"`
}

// quoteRequestDTO is the JSON body of POST /shipping/quote.
type quoteRequestDTO struct {
	OriginZipCode       string         `json:"origin_zip_code"`
	DestinationZipCode  string         `json:"destination_zip_code"`
	OriginZoneCode      string         `json:"origin_zone_code"`
	DestinationZoneCode string         `json:"destination_zone_code"`
	DistanceKm          float64        `json:"distance_km"`
	WeightKg            float64        `json:"weight_kg"`
	Dimensions          *dimensionsDTO `json:"dimensions,omitempty"`
	ShippingType        string         `json:"shipping_type"`
	Priority            string         `json:"priority"`
	PromotionCode       string         `json:"promotion_code,omitempty"`
}

// toDomain maps the DTO to the domain request. Priority defaults to normal when
// omitted.
func (r quoteRequestDTO) toDomain() domain.QuoteRequest {
	var dims *domain.Dimensions
	if r.Dimensions != nil {
		dims = &domain.Dimensions{
			HeightCm: r.Dimensions.HeightCm,
			WidthCm:  r.Dimensions.WidthCm,
			LengthCm: r.Dimensions.LengthCm,
		}
	}
	priority := domain.Priority(r.Priority)
	if priority == "" {
		priority = domain.PriorityNormal
	}
	return domain.QuoteRequest{
		OriginZipCode:       r.OriginZipCode,
		DestinationZipCode:  r.DestinationZipCode,
		OriginZoneCode:      r.OriginZoneCode,
		DestinationZoneCode: r.DestinationZoneCode,
		DistanceKm:          r.DistanceKm,
		WeightKg:            r.WeightKg,
		Dimensions:          dims,
		ShippingType:        domain.ShippingType(r.ShippingType),
		Priority:            priority,
		PromotionCode:       r.PromotionCode,
	}
}
