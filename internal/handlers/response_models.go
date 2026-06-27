package handlers

import "github.com/TobiasMoreno/shipping-pricing-api/internal/domain"

type breakdownDTO struct {
	BasePrice           int64 `json:"base_price"`
	DistanceFee         int64 `json:"distance_fee"`
	WeightFee           int64 `json:"weight_fee"`
	PriorityFee         int64 `json:"priority_fee"`
	ExpressFee          int64 `json:"express_fee"`
	Discount            int64 `json:"discount"`
	TotalBeforeDiscount int64 `json:"total_before_discount"`
	Total               int64 `json:"total"`
}

// quoteResponseDTO is the JSON body returned by POST /shipping/quote. Monetary
// values are integer cents.
type quoteResponseDTO struct {
	QuoteID               string       `json:"quote_id"`
	Price                 int64        `json:"price"`
	Currency              string       `json:"currency"`
	EstimatedDeliveryDays int          `json:"estimated_delivery_days"`
	ShippingType          string       `json:"shipping_type"`
	Available             bool         `json:"available"`
	Cached                bool         `json:"cached"`
	Breakdown             breakdownDTO `json:"breakdown"`
	DecisionTrace         []string     `json:"decision_trace"`
}

func newQuoteResponse(q domain.Quote, cached bool) quoteResponseDTO {
	return quoteResponseDTO{
		QuoteID:               newQuoteID(),
		Price:                 q.Price.Cents(),
		Currency:              q.Currency,
		EstimatedDeliveryDays: q.EstimatedDeliveryDays,
		ShippingType:          string(q.ShippingType),
		Available:             q.Available,
		Cached:                cached,
		Breakdown: breakdownDTO{
			BasePrice:           q.Breakdown.BasePrice.Cents(),
			DistanceFee:         q.Breakdown.DistanceFee.Cents(),
			WeightFee:           q.Breakdown.WeightFee.Cents(),
			PriorityFee:         q.Breakdown.PriorityFee.Cents(),
			ExpressFee:          q.Breakdown.ExpressFee.Cents(),
			Discount:            q.Breakdown.Discount.Cents(),
			TotalBeforeDiscount: q.Breakdown.TotalBeforeDiscount.Cents(),
			Total:               q.Breakdown.Total.Cents(),
		},
		DecisionTrace: []string(q.DecisionTrace),
	}
}
