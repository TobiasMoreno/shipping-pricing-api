package domain

import "math"

// Money represents a monetary amount as an integer number of cents. The domain
// never uses floating-point types to store monetary values; any fractional cent
// produced by an intermediate computation is rounded half-up to a whole cent.
type Money int64

// Cents returns the raw integer cent value.
func (m Money) Cents() int64 { return int64(m) }

// Add returns the sum of two amounts.
func (m Money) Add(other Money) Money { return m + other }

// Sub returns the difference of two amounts.
func (m Money) Sub(other Money) Money { return m - other }

// MulFloat multiplies the amount by a floating-point factor and rounds the
// result half-up to the nearest cent. It is used for proportional fees
// (distance, weight) and multiplier surcharges (express, priority).
func (m Money) MulFloat(factor float64) Money {
	return Money(math.Round(float64(m) * factor))
}

// CentsFromFloat converts a floating-point amount of cents into Money, rounding
// half-up to the nearest whole cent.
func CentsFromFloat(cents float64) Money {
	return Money(math.Round(cents))
}
