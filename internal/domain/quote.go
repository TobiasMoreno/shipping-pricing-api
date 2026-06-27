package domain

// Dimensions describes the package dimensions in centimeters.
type Dimensions struct {
	HeightCm float64
	WidthCm  float64
	LengthCm float64
}

// QuoteRequest is the input to a pricing calculation. Zones, rule and promotion
// are resolved by upper layers and passed to the engine separately; this struct
// only carries the caller-provided request data.
type QuoteRequest struct {
	OriginZipCode       string
	DestinationZipCode  string
	OriginZoneCode      string
	DestinationZoneCode string
	DistanceKm          float64
	WeightKg            float64
	Dimensions          *Dimensions
	ShippingType        ShippingType
	Priority            Priority
	PromotionCode       string
}

// PricingBreakdown is the itemized detail of a price calculation. All fields are
// in cents and reconcile exactly:
//
//	TotalBeforeDiscount = BasePrice + DistanceFee + WeightFee + ExpressFee + PriorityFee
//	Total               = TotalBeforeDiscount - Discount
type PricingBreakdown struct {
	BasePrice           Money
	DistanceFee         Money
	WeightFee           Money
	PriorityFee         Money
	ExpressFee          Money
	Discount            Money
	TotalBeforeDiscount Money
	Total               Money
}

// PricingDecisionTrace is an ordered, human-readable log of the decisions taken
// while calculating a quote.
type PricingDecisionTrace []string

// Add appends a decision message to the trace.
func (t *PricingDecisionTrace) Add(msg string) {
	*t = append(*t, msg)
}

// Quote is the result of a pricing calculation.
type Quote struct {
	Price                 Money
	Currency              string
	EstimatedDeliveryDays int
	ShippingType          ShippingType
	Available             bool
	Breakdown             PricingBreakdown
	DecisionTrace         PricingDecisionTrace
}
