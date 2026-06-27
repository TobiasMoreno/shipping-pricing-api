package domain

// ShippingType enumerates the supported shipping speeds.
type ShippingType string

const (
	ShippingStandard ShippingType = "standard"
	ShippingExpress  ShippingType = "express"
	ShippingSameDay  ShippingType = "same_day"
)

// Priority enumerates the supported request priorities.
type Priority string

const (
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
)

// PricingRule holds the configurable parameters used to price a shipment for a
// given shipping type and route. Monetary fields are expressed in cents.
type PricingRule struct {
	ShippingType        ShippingType
	OriginZoneCode      string
	DestinationZoneCode string
	BasePrice           Money // cents
	PricePerKm          Money // cents per km
	PricePerKg          Money // cents per kg
	ExpressMultiplier   float64
	PriorityMultiplier  float64
	MaxWeightKg         float64
	MinDistanceKm       float64
	MaxDistanceKm       float64
	IsActive            bool
}
