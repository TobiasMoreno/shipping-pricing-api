package domain

import "time"

// DiscountType enumerates the supported promotion discount kinds.
type DiscountType string

const (
	// DiscountPercentage applies a percentage of the total before discount.
	DiscountPercentage DiscountType = "percentage"
	// DiscountFixedAmount applies a flat amount expressed in cents.
	DiscountFixedAmount DiscountType = "fixed_amount"
)

// Promotion represents a simple discount applicable to a quote.
//
// DiscountValue is interpreted according to DiscountType: for percentage it is
// the percentage points (e.g. 10 means 10%); for fixed_amount it is the amount
// of cents to subtract.
type Promotion struct {
	Code          string
	Description   string
	DiscountType  DiscountType
	DiscountValue float64
	MaxDiscount   Money // cap in cents; zero means no cap
	StartsAt      time.Time
	EndsAt        time.Time
	IsActive      bool
}

// IsValidAt reports whether the promotion is active and within its validity
// window at the given evaluation time. Zero StartsAt/EndsAt values are treated
// as open-ended bounds.
func (p Promotion) IsValidAt(now time.Time) bool {
	if !p.IsActive {
		return false
	}
	if !p.StartsAt.IsZero() && now.Before(p.StartsAt) {
		return false
	}
	if !p.EndsAt.IsZero() && now.After(p.EndsAt) {
		return false
	}
	return true
}
