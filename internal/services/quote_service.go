package services

import (
	"context"
	"time"

	"github.com/TobiasMoreno/shipping-pricing-api/internal/domain"
)

// QuoteService orchestrates the quote use case: it resolves zones, selects the
// applicable pricing rule and promotion, and delegates the calculation to the
// pure domain engine.
type QuoteService struct {
	rules      PricingRuleRepository
	zones      ZoneRepository
	promotions PromotionRepository
}

// NewQuoteService wires the service with its repositories.
func NewQuoteService(rules PricingRuleRepository, zones ZoneRepository, promotions PromotionRepository) *QuoteService {
	return &QuoteService{rules: rules, zones: zones, promotions: promotions}
}

// Quote resolves the dependencies of the request and returns the calculated
// quote. Domain and repository errors are returned as-is so the transport layer
// can map them to status codes.
func (s *QuoteService) Quote(ctx context.Context, req domain.QuoteRequest, now time.Time) (domain.Quote, error) {
	origin, err := s.zones.GetByCode(ctx, req.OriginZoneCode)
	if err != nil {
		return domain.Quote{}, err
	}
	destination, err := s.zones.GetByCode(ctx, req.DestinationZoneCode)
	if err != nil {
		return domain.Quote{}, err
	}

	candidates, err := s.rules.FindActiveRules(ctx, req.ShippingType, req.OriginZoneCode, req.DestinationZoneCode)
	if err != nil {
		return domain.Quote{}, err
	}
	rule := selectRule(candidates, req)
	if rule == nil {
		return domain.Quote{}, domain.ErrNoApplicableRule
	}

	var promo *domain.Promotion
	if req.PromotionCode != "" {
		p, err := s.promotions.GetByCode(ctx, req.PromotionCode)
		if err != nil {
			return domain.Quote{}, err
		}
		promo = &p
	}

	quote, err := domain.Calculate(req, rule, origin, destination, promo, now)
	if err != nil {
		return domain.Quote{}, err
	}
	// ETA is a placeholder by shipping type until the logistics provider is
	// integrated in a later change.
	quote.EstimatedDeliveryDays = etaForShippingType(req.ShippingType)
	return quote, nil
}

// selectRule picks the most specific active rule matching the request: exact
// zone matches score higher than the `*` wildcard. Returns nil when none match.
func selectRule(candidates []domain.PricingRule, req domain.QuoteRequest) *domain.PricingRule {
	var best *domain.PricingRule
	bestScore := 0
	for i := range candidates {
		r := candidates[i]
		if !r.IsActive || r.ShippingType != req.ShippingType {
			continue
		}
		os := zoneScore(r.OriginZoneCode, req.OriginZoneCode)
		ds := zoneScore(r.DestinationZoneCode, req.DestinationZoneCode)
		if os == 0 || ds == 0 {
			continue // a zone does not match at all
		}
		if score := os + ds; score > bestScore {
			bestScore = score
			best = &candidates[i]
		}
	}
	return best
}

// zoneScore returns 2 for an exact match, 1 for a wildcard match and 0 for no
// match.
func zoneScore(ruleZone, requestZone string) int {
	switch ruleZone {
	case requestZone:
		return 2
	case "*":
		return 1
	default:
		return 0
	}
}

func etaForShippingType(t domain.ShippingType) int {
	switch t {
	case domain.ShippingExpress:
		return 1
	case domain.ShippingSameDay:
		return 0
	default:
		return 3
	}
}
