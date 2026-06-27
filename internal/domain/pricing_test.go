package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/TobiasMoreno/shipping-pricing-api/internal/domain"
)

// evalTime is the fixed evaluation time used across tests so promotion validity
// is deterministic.
var evalTime = time.Date(2026, time.June, 27, 12, 0, 0, 0, time.UTC)

func baseRule() *domain.PricingRule {
	return &domain.PricingRule{
		ShippingType:        domain.ShippingStandard,
		OriginZoneCode:      "CORDOBA_CAPITAL",
		DestinationZoneCode: "CABA",
		BasePrice:           5000,
		PricePerKm:          4,
		PricePerKg:          1000,
		ExpressMultiplier:   1.5,
		PriorityMultiplier:  1.15,
		MaxWeightKg:         30,
		MinDistanceKm:       0,
		MaxDistanceKm:       1000,
		IsActive:            true,
	}
}

func baseReq() domain.QuoteRequest {
	return domain.QuoteRequest{
		OriginZoneCode:      "CORDOBA_CAPITAL",
		DestinationZoneCode: "CABA",
		DistanceKm:          695,
		WeightKg:            2.5,
		ShippingType:        domain.ShippingStandard,
		Priority:            domain.PriorityNormal,
	}
}

func activeZone(code string) domain.Zone {
	return domain.Zone{Code: code, IsActive: true}
}

// flatRule returns a rule whose subtotal equals exactly its base price (no
// distance/weight fees), useful for asserting surcharge and discount math.
func flatRule(base domain.Money) *domain.PricingRule {
	r := baseRule()
	r.BasePrice = base
	r.PricePerKm = 0
	r.PricePerKg = 0
	return r
}

func TestCalculate_StandardWithoutPromotion(t *testing.T) {
	q, err := domain.Calculate(baseReq(), baseRule(), activeZone("CORDOBA_CAPITAL"), activeZone("CABA"), nil, evalTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b := q.Breakdown
	if b.BasePrice != 5000 {
		t.Errorf("BasePrice = %d, want 5000", b.BasePrice.Cents())
	}
	if b.DistanceFee != 2780 {
		t.Errorf("DistanceFee = %d, want 2780", b.DistanceFee.Cents())
	}
	if b.WeightFee != 2500 {
		t.Errorf("WeightFee = %d, want 2500", b.WeightFee.Cents())
	}
	if b.ExpressFee != 0 || b.PriorityFee != 0 || b.Discount != 0 {
		t.Errorf("expected no surcharges/discount, got express=%d priority=%d discount=%d",
			b.ExpressFee.Cents(), b.PriorityFee.Cents(), b.Discount.Cents())
	}
	if b.Total != 10280 {
		t.Errorf("Total = %d, want 10280", b.Total.Cents())
	}
	if q.Price != b.Total {
		t.Errorf("Quote.Price = %d, want %d", q.Price.Cents(), b.Total.Cents())
	}
	if q.Currency != domain.DefaultCurrency {
		t.Errorf("Currency = %q, want %q", q.Currency, domain.DefaultCurrency)
	}
	assertReconciles(t, b)
	if len(q.DecisionTrace) == 0 {
		t.Error("expected a non-empty decision trace")
	}
}

func TestCalculate_ExpressSurcharge(t *testing.T) {
	req := baseReq()
	req.ShippingType = domain.ShippingExpress
	rule := flatRule(10000)
	rule.ExpressMultiplier = 1.5

	q, err := domain.Calculate(req, rule, activeZone("CORDOBA_CAPITAL"), activeZone("CABA"), nil, evalTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Breakdown.ExpressFee != 5000 {
		t.Errorf("ExpressFee = %d, want 5000", q.Breakdown.ExpressFee.Cents())
	}
	if q.Breakdown.PriorityFee != 0 {
		t.Errorf("PriorityFee = %d, want 0", q.Breakdown.PriorityFee.Cents())
	}
	if q.Breakdown.Total != 15000 {
		t.Errorf("Total = %d, want 15000", q.Breakdown.Total.Cents())
	}
	assertReconciles(t, q.Breakdown)
}

func TestCalculate_NoExpressSurchargeForStandard(t *testing.T) {
	q, err := domain.Calculate(baseReq(), baseRule(), activeZone("CORDOBA_CAPITAL"), activeZone("CABA"), nil, evalTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Breakdown.ExpressFee != 0 {
		t.Errorf("ExpressFee = %d, want 0 for standard", q.Breakdown.ExpressFee.Cents())
	}
}

func TestCalculate_PrioritySurcharge(t *testing.T) {
	req := baseReq()
	req.Priority = domain.PriorityHigh
	rule := flatRule(10000)
	rule.PriorityMultiplier = 1.15

	q, err := domain.Calculate(req, rule, activeZone("CORDOBA_CAPITAL"), activeZone("CABA"), nil, evalTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Breakdown.PriorityFee != 1500 {
		t.Errorf("PriorityFee = %d, want 1500", q.Breakdown.PriorityFee.Cents())
	}
	assertReconciles(t, q.Breakdown)
}

func TestCalculate_NoPrioritySurchargeForNormal(t *testing.T) {
	q, _ := domain.Calculate(baseReq(), baseRule(), activeZone("CORDOBA_CAPITAL"), activeZone("CABA"), nil, evalTime)
	if q.Breakdown.PriorityFee != 0 {
		t.Errorf("PriorityFee = %d, want 0 for normal priority", q.Breakdown.PriorityFee.Cents())
	}
}

func TestCalculate_PercentagePromotionUnderCap(t *testing.T) {
	promo := &domain.Promotion{
		Code:          "SHIP10",
		DiscountType:  domain.DiscountPercentage,
		DiscountValue: 10,
		MaxDiscount:   5000,
		IsActive:      true,
	}
	q, err := domain.Calculate(baseReq(), flatRule(13500), activeZone("CORDOBA_CAPITAL"), activeZone("CABA"), promo, evalTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Breakdown.Discount != 1350 {
		t.Errorf("Discount = %d, want 1350", q.Breakdown.Discount.Cents())
	}
	if q.Breakdown.Total != 12150 {
		t.Errorf("Total = %d, want 12150", q.Breakdown.Total.Cents())
	}
	assertReconciles(t, q.Breakdown)
}

func TestCalculate_PercentagePromotionLimitedByCap(t *testing.T) {
	promo := &domain.Promotion{
		Code:          "BIG50",
		DiscountType:  domain.DiscountPercentage,
		DiscountValue: 50,
		MaxDiscount:   5000,
		IsActive:      true,
	}
	q, err := domain.Calculate(baseReq(), flatRule(100000), activeZone("CORDOBA_CAPITAL"), activeZone("CABA"), promo, evalTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Breakdown.Discount != 5000 {
		t.Errorf("Discount = %d, want 5000 (capped)", q.Breakdown.Discount.Cents())
	}
}

func TestCalculate_FixedAmountPromotion(t *testing.T) {
	promo := &domain.Promotion{
		Code:          "FLAT1000",
		DiscountType:  domain.DiscountFixedAmount,
		DiscountValue: 1000,
		IsActive:      true,
	}
	q, err := domain.Calculate(baseReq(), flatRule(13500), activeZone("CORDOBA_CAPITAL"), activeZone("CABA"), promo, evalTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Breakdown.Discount != 1000 {
		t.Errorf("Discount = %d, want 1000", q.Breakdown.Discount.Cents())
	}
	if q.Breakdown.Total != 12500 {
		t.Errorf("Total = %d, want 12500", q.Breakdown.Total.Cents())
	}
}

func TestCalculate_FixedAmountCannotExceedTotal(t *testing.T) {
	promo := &domain.Promotion{
		Code:          "FLAT1000",
		DiscountType:  domain.DiscountFixedAmount,
		DiscountValue: 1000,
		IsActive:      true,
	}
	q, err := domain.Calculate(baseReq(), flatRule(800), activeZone("CORDOBA_CAPITAL"), activeZone("CABA"), promo, evalTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Breakdown.Discount != 800 {
		t.Errorf("Discount = %d, want 800 (clamped to total)", q.Breakdown.Discount.Cents())
	}
	if q.Breakdown.Total != 0 {
		t.Errorf("Total = %d, want 0", q.Breakdown.Total.Cents())
	}
}

func TestCalculate_DomainErrors(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*domain.QuoteRequest, **domain.PricingRule, *domain.Zone, *domain.Zone, **domain.Promotion)
		wantErr error
	}{
		{
			name: "weight exceeded",
			mutate: func(req *domain.QuoteRequest, _ **domain.PricingRule, _, _ *domain.Zone, _ **domain.Promotion) {
				req.WeightKg = 35
			},
			wantErr: domain.ErrWeightExceeded,
		},
		{
			name: "distance out of range",
			mutate: func(req *domain.QuoteRequest, _ **domain.PricingRule, _, _ *domain.Zone, _ **domain.Promotion) {
				req.DistanceKm = 1500
			},
			wantErr: domain.ErrDistanceOutOfRange,
		},
		{
			name: "destination zone inactive",
			mutate: func(_ *domain.QuoteRequest, _ **domain.PricingRule, _, dst *domain.Zone, _ **domain.Promotion) {
				dst.IsActive = false
			},
			wantErr: domain.ErrZoneInactive,
		},
		{
			name: "no applicable rule",
			mutate: func(_ *domain.QuoteRequest, rule **domain.PricingRule, _, _ *domain.Zone, _ **domain.Promotion) {
				*rule = nil
			},
			wantErr: domain.ErrNoApplicableRule,
		},
		{
			name: "expired promotion",
			mutate: func(_ *domain.QuoteRequest, _ **domain.PricingRule, _, _ *domain.Zone, promo **domain.Promotion) {
				*promo = &domain.Promotion{
					Code:          "OLD",
					DiscountType:  domain.DiscountPercentage,
					DiscountValue: 10,
					IsActive:      true,
					EndsAt:        evalTime.Add(-time.Hour),
				}
			},
			wantErr: domain.ErrPromotionInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := baseReq()
			rule := baseRule()
			origin := activeZone("CORDOBA_CAPITAL")
			dst := activeZone("CABA")
			var promo *domain.Promotion
			tt.mutate(&req, &rule, &origin, &dst, &promo)

			_, err := domain.Calculate(req, rule, origin, dst, promo, evalTime)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want errors.Is(_, %v)", err, tt.wantErr)
			}
		})
	}
}

func assertReconciles(t *testing.T, b domain.PricingBreakdown) {
	t.Helper()
	wantTBD := b.BasePrice + b.DistanceFee + b.WeightFee + b.ExpressFee + b.PriorityFee
	if b.TotalBeforeDiscount != wantTBD {
		t.Errorf("TotalBeforeDiscount = %d, want %d (sum of components)", b.TotalBeforeDiscount.Cents(), wantTBD.Cents())
	}
	if b.Total != b.TotalBeforeDiscount-b.Discount {
		t.Errorf("Total = %d, want %d (TBD - discount)", b.Total.Cents(), (b.TotalBeforeDiscount - b.Discount).Cents())
	}
}
