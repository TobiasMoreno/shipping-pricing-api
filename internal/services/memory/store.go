// Package memory provides in-memory implementations of the service
// repositories, seeded with sample data. They let the quote endpoint work
// without PostgreSQL and are replaced by real repositories in a later change.
package memory

import (
	"context"
	"sync"

	"github.com/TobiasMoreno/shipping-pricing-api/internal/domain"
	"github.com/TobiasMoreno/shipping-pricing-api/internal/services"
)

// ZoneStore is an in-memory ZoneRepository.
type ZoneStore struct {
	mu    sync.RWMutex
	zones map[string]domain.Zone
}

// NewZoneStore creates a zone store from the given seed map.
func NewZoneStore(seed map[string]domain.Zone) *ZoneStore {
	return &ZoneStore{zones: seed}
}

// GetByCode returns the zone or services.ErrZoneNotFound.
func (s *ZoneStore) GetByCode(_ context.Context, code string) (domain.Zone, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	z, ok := s.zones[code]
	if !ok {
		return domain.Zone{}, services.ErrZoneNotFound
	}
	return z, nil
}

// RuleStore is an in-memory PricingRuleRepository.
type RuleStore struct {
	mu    sync.RWMutex
	rules []domain.PricingRule
}

// NewRuleStore creates a rule store from the given seed slice.
func NewRuleStore(seed []domain.PricingRule) *RuleStore {
	return &RuleStore{rules: seed}
}

// FindActiveRules returns active rules matching the shipping type and route,
// treating `*` as a zone wildcard.
func (s *RuleStore) FindActiveRules(_ context.Context, shippingType domain.ShippingType, origin, destination string) ([]domain.PricingRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.PricingRule
	for _, r := range s.rules {
		if !r.IsActive || r.ShippingType != shippingType {
			continue
		}
		if (r.OriginZoneCode == origin || r.OriginZoneCode == "*") &&
			(r.DestinationZoneCode == destination || r.DestinationZoneCode == "*") {
			out = append(out, r)
		}
	}
	return out, nil
}

// PromotionStore is an in-memory PromotionRepository.
type PromotionStore struct {
	mu     sync.RWMutex
	promos map[string]domain.Promotion
}

// NewPromotionStore creates a promotion store from the given seed map.
func NewPromotionStore(seed map[string]domain.Promotion) *PromotionStore {
	return &PromotionStore{promos: seed}
}

// GetByCode returns the promotion or services.ErrPromotionNotFound.
func (s *PromotionStore) GetByCode(_ context.Context, code string) (domain.Promotion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.promos[code]
	if !ok {
		return domain.Promotion{}, services.ErrPromotionNotFound
	}
	return p, nil
}

// Seed builds the three stores with sample data used for local runs and demos.
func Seed() (*ZoneStore, *RuleStore, *PromotionStore) {
	zones := NewZoneStore(map[string]domain.Zone{
		"CABA":            {Code: "CABA", Name: "Ciudad de Buenos Aires", Region: "Centro", IsActive: true},
		"CORDOBA_CAPITAL": {Code: "CORDOBA_CAPITAL", Name: "Córdoba Capital", Region: "Centro", IsActive: true},
		"PATAGONIA":       {Code: "PATAGONIA", Name: "Patagonia", Region: "Sur", IsActive: false},
	})

	rules := NewRuleStore([]domain.PricingRule{
		{
			ShippingType: domain.ShippingStandard, OriginZoneCode: "CORDOBA_CAPITAL", DestinationZoneCode: "CABA",
			BasePrice: 5000, PricePerKm: 4, PricePerKg: 1000,
			ExpressMultiplier: 1.5, PriorityMultiplier: 1.15,
			MaxWeightKg: 30, MinDistanceKm: 0, MaxDistanceKm: 1500, IsActive: true,
		},
		{
			ShippingType: domain.ShippingExpress, OriginZoneCode: "CORDOBA_CAPITAL", DestinationZoneCode: "CABA",
			BasePrice: 8000, PricePerKm: 6, PricePerKg: 1500,
			ExpressMultiplier: 1.5, PriorityMultiplier: 1.15,
			MaxWeightKg: 20, MinDistanceKm: 0, MaxDistanceKm: 1500, IsActive: true,
		},
		{
			// Wildcard fallback rule for standard shipments on any route.
			ShippingType: domain.ShippingStandard, OriginZoneCode: "*", DestinationZoneCode: "*",
			BasePrice: 9000, PricePerKm: 5, PricePerKg: 1200,
			ExpressMultiplier: 1.5, PriorityMultiplier: 1.2,
			MaxWeightKg: 40, MinDistanceKm: 0, MaxDistanceKm: 5000, IsActive: true,
		},
	})

	promos := NewPromotionStore(map[string]domain.Promotion{
		"SHIP10": {
			Code: "SHIP10", Description: "10% off shipping", DiscountType: domain.DiscountPercentage,
			DiscountValue: 10, MaxDiscount: 5000, IsActive: true,
		},
	})

	return zones, rules, promos
}
