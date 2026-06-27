package domain

import "errors"

// Domain errors returned by the pricing engine. They are sentinel values meant
// to be compared with errors.Is; upper layers map them to transport-specific
// status codes (the domain itself knows nothing about HTTP).
var (
	ErrNoApplicableRule   = errors.New("no applicable pricing rule")
	ErrWeightExceeded     = errors.New("weight exceeds the maximum allowed by the pricing rule")
	ErrDistanceOutOfRange = errors.New("distance is outside the pricing rule range")
	ErrZoneInactive       = errors.New("shipping zone is not available")
	ErrPromotionInvalid   = errors.New("promotion is invalid or expired")
)
