package domain

import (
	"fmt"
	"time"
)

// DefaultCurrency is the currency used for quotes in the MVP.
const DefaultCurrency = "ARS"

// Calculate prices a shipment deterministically from a request and the resolved
// applicable rule, zones and optional promotion. It performs no I/O and has no
// side effects. The evaluation time (now) is injected so promotion validity is
// testable without reading the global clock.
//
// A nil rule means no applicable pricing rule was found and yields
// ErrNoApplicableRule. A nil promo means no promotion is applied.
func Calculate(req QuoteRequest, rule *PricingRule, origin, destination Zone, promo *Promotion, now time.Time) (Quote, error) {
	if rule == nil {
		return Quote{}, ErrNoApplicableRule
	}

	// Business restrictions, evaluated before any pricing.
	if !origin.IsActive {
		return Quote{}, fmt.Errorf("origin zone %q: %w", origin.Code, ErrZoneInactive)
	}
	if !destination.IsActive {
		return Quote{}, fmt.Errorf("destination zone %q: %w", destination.Code, ErrZoneInactive)
	}
	if rule.MaxWeightKg > 0 && req.WeightKg > rule.MaxWeightKg {
		return Quote{}, fmt.Errorf("weight %.2f kg exceeds max %.2f kg: %w", req.WeightKg, rule.MaxWeightKg, ErrWeightExceeded)
	}
	if req.DistanceKm < rule.MinDistanceKm || (rule.MaxDistanceKm > 0 && req.DistanceKm > rule.MaxDistanceKm) {
		return Quote{}, fmt.Errorf("distance %.2f km outside [%.2f, %.2f]: %w", req.DistanceKm, rule.MinDistanceKm, rule.MaxDistanceKm, ErrDistanceOutOfRange)
	}

	trace := PricingDecisionTrace{}
	trace.Add(fmt.Sprintf("Matched pricing rule for %s -> %s / %s", rule.OriginZoneCode, rule.DestinationZoneCode, rule.ShippingType))

	basePrice := rule.BasePrice
	trace.Add(fmt.Sprintf("Applied base price: %d cents", basePrice.Cents()))

	distanceFee := rule.PricePerKm.MulFloat(req.DistanceKm)
	trace.Add(fmt.Sprintf("Applied distance fee: %.2f km * %d cents/km = %d cents", req.DistanceKm, rule.PricePerKm.Cents(), distanceFee.Cents()))

	weightFee := rule.PricePerKg.MulFloat(req.WeightKg)
	trace.Add(fmt.Sprintf("Applied weight fee: %.2f kg * %d cents/kg = %d cents", req.WeightKg, rule.PricePerKg.Cents(), weightFee.Cents()))

	subtotal := basePrice.Add(distanceFee).Add(weightFee)

	var expressFee Money
	if req.ShippingType == ShippingExpress && rule.ExpressMultiplier > 1 {
		expressFee = subtotal.MulFloat(rule.ExpressMultiplier - 1)
		trace.Add(fmt.Sprintf("Applied express surcharge: %d * (%.2f - 1) = %d cents", subtotal.Cents(), rule.ExpressMultiplier, expressFee.Cents()))
	}

	var priorityFee Money
	if req.Priority == PriorityHigh && rule.PriorityMultiplier > 1 {
		priorityFee = subtotal.MulFloat(rule.PriorityMultiplier - 1)
		trace.Add(fmt.Sprintf("Applied priority surcharge: %d * (%.2f - 1) = %d cents", subtotal.Cents(), rule.PriorityMultiplier, priorityFee.Cents()))
	}

	totalBeforeDiscount := subtotal.Add(expressFee).Add(priorityFee)

	discount, err := computeDiscount(promo, totalBeforeDiscount, now, &trace)
	if err != nil {
		return Quote{}, err
	}

	total := totalBeforeDiscount.Sub(discount)
	if total < 0 {
		total = 0
	}

	return Quote{
		Price:                 total,
		Currency:              DefaultCurrency,
		EstimatedDeliveryDays: 0, // resolved by the logistics provider in a later change
		ShippingType:          req.ShippingType,
		Available:             true,
		Breakdown: PricingBreakdown{
			BasePrice:           basePrice,
			DistanceFee:         distanceFee,
			WeightFee:           weightFee,
			PriorityFee:         priorityFee,
			ExpressFee:          expressFee,
			Discount:            discount,
			TotalBeforeDiscount: totalBeforeDiscount,
			Total:               total,
		},
		DecisionTrace: trace,
	}, nil
}

// computeDiscount validates the optional promotion and returns the discount to
// apply, capped so it never exceeds the total before discount. An explicitly
// supplied invalid or expired promotion is a domain error rather than a no-op.
func computeDiscount(promo *Promotion, totalBeforeDiscount Money, now time.Time, trace *PricingDecisionTrace) (Money, error) {
	if promo == nil {
		return 0, nil
	}
	if !promo.IsValidAt(now) {
		return 0, fmt.Errorf("promotion %q: %w", promo.Code, ErrPromotionInvalid)
	}

	var discount Money
	switch promo.DiscountType {
	case DiscountPercentage:
		discount = totalBeforeDiscount.MulFloat(promo.DiscountValue / 100)
	case DiscountFixedAmount:
		discount = CentsFromFloat(promo.DiscountValue)
	default:
		return 0, fmt.Errorf("promotion %q has unknown discount type %q: %w", promo.Code, promo.DiscountType, ErrPromotionInvalid)
	}

	if promo.MaxDiscount > 0 && discount > promo.MaxDiscount {
		discount = promo.MaxDiscount
	}
	if discount > totalBeforeDiscount {
		discount = totalBeforeDiscount
	}

	trace.Add(fmt.Sprintf("Applied promotion %q: discount %d cents", promo.Code, discount.Cents()))
	return discount, nil
}
