// Package services orchestrates the application use cases, coordinating domain
// logic with repositories, cache and external clients through interfaces it
// defines on the consumer side.
package services

import (
	"context"
	"errors"

	"github.com/TobiasMoreno/shipping-pricing-api/internal/domain"
)

// Repository lookup errors. Repositories return these so the service and
// handlers can distinguish "not found" from infrastructure failures.
var (
	ErrZoneNotFound      = errors.New("zone not found")
	ErrPromotionNotFound = errors.New("promotion not found")
)

// PricingRuleRepository returns the active pricing rules that are candidates for
// a given shipping type and route (matching the exact zone or the `*` wildcard).
type PricingRuleRepository interface {
	FindActiveRules(ctx context.Context, shippingType domain.ShippingType, origin, destination string) ([]domain.PricingRule, error)
}

// ZoneRepository resolves a zone by its code, returning ErrZoneNotFound when it
// does not exist.
type ZoneRepository interface {
	GetByCode(ctx context.Context, code string) (domain.Zone, error)
}

// PromotionRepository resolves a promotion by its code, returning
// ErrPromotionNotFound when it does not exist.
type PromotionRepository interface {
	GetByCode(ctx context.Context, code string) (domain.Promotion, error)
}
