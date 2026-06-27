## ADDED Requirements

### Requirement: Money is represented in integer cents

All monetary values SHALL be represented as a `Money` value object backed by a signed 64-bit integer count of cents. The engine MUST NOT use floating-point types for storing or returning monetary amounts. Any intermediate calculation that produces a fractional cent SHALL be rounded to the nearest whole cent using round-half-up, and the rounded result SHALL be used for all subsequent steps and in the breakdown.

#### Scenario: Fractional cent is rounded half-up
- **WHEN** a fee computation yields 5000.5 cents
- **THEN** the stored fee is 5001 cents

#### Scenario: Fee below half cent rounds down
- **WHEN** a fee computation yields 5000.4 cents
- **THEN** the stored fee is 5000 cents

### Requirement: Quote calculation produces price, breakdown and decision trace

Given a valid `QuoteRequest` and the applicable `PricingRule` (plus the resolved origin/destination `Zone`s and an optional `Promotion`), the engine SHALL return a `Quote` containing the final price in cents, the currency, the estimated delivery days, the shipping type, the availability flag, a `PricingBreakdown`, and a `PricingDecisionTrace`. The `PricingBreakdown` MUST contain `base_price`, `distance_fee`, `weight_fee`, `priority_fee`, `express_fee`, `discount`, `total_before_discount` and `total`, all in cents.

#### Scenario: Standard shipment without promotion
- **WHEN** a valid standard `QuoteRequest` is calculated against an active applicable rule with no promotion
- **THEN** the returned `Quote` has `total = base_price + distance_fee + weight_fee`
- **AND** `express_fee` and `priority_fee` are 0
- **AND** `discount` is 0
- **AND** the breakdown components sum exactly to `total`

#### Scenario: Breakdown always reconciles with total
- **WHEN** any quote is successfully calculated
- **THEN** `total_before_discount = base_price + distance_fee + weight_fee + express_fee + priority_fee`
- **AND** `total = total_before_discount - discount`

### Requirement: Base price is taken from the applicable rule

The engine SHALL use the rule's `base_price_cents` as the `base_price` component of the breakdown.

#### Scenario: Base price applied
- **WHEN** the applicable rule has `base_price_cents = 5000`
- **THEN** the breakdown `base_price` is 5000
- **AND** the decision trace records that the base price was applied

### Requirement: Distance fee is proportional to distance

The engine SHALL compute `distance_fee = round(distance_km * price_per_km_cents)` using the rule's `price_per_km_cents`.

#### Scenario: Distance fee computed
- **WHEN** `distance_km = 695` and the rule's `price_per_km_cents = 4`
- **THEN** the breakdown `distance_fee` is 2780
- **AND** the decision trace records the distance fee calculation

### Requirement: Weight fee is proportional to weight

The engine SHALL compute `weight_fee = round(weight_kg * price_per_kg_cents)` using the rule's `price_per_kg_cents`.

#### Scenario: Weight fee computed
- **WHEN** `weight_kg = 2.5` and the rule's `price_per_kg_cents = 1000`
- **THEN** the breakdown `weight_fee` is 2500
- **AND** the decision trace records the weight fee calculation

### Requirement: Express multiplier applies a surcharge for express shipments

When `shipping_type` is `express`, the engine SHALL apply the rule's `express_multiplier` to the sum of base, distance and weight fees and record the surcharge as `express_fee = round(subtotal * (express_multiplier - 1))`, where `subtotal = base_price + distance_fee + weight_fee`. For `standard` and `same_day` shipping types the `express_fee` SHALL be 0.

#### Scenario: Express surcharge applied
- **WHEN** `shipping_type = express`, the subtotal is 10000 cents and `express_multiplier = 1.5`
- **THEN** the breakdown `express_fee` is 5000
- **AND** the decision trace records the express surcharge

#### Scenario: No express surcharge for standard
- **WHEN** `shipping_type = standard`
- **THEN** the breakdown `express_fee` is 0

### Requirement: Priority multiplier applies a surcharge for high priority

When `priority` is `high`, the engine SHALL apply the rule's `priority_multiplier` to the subtotal of base, distance and weight fees and record the surcharge as `priority_fee = round(subtotal * (priority_multiplier - 1))`. When `priority` is `normal` the `priority_fee` SHALL be 0.

#### Scenario: Priority surcharge applied
- **WHEN** `priority = high`, the subtotal is 10000 cents and `priority_multiplier = 1.15`
- **THEN** the breakdown `priority_fee` is 1500
- **AND** the decision trace records the priority surcharge

#### Scenario: No priority surcharge for normal priority
- **WHEN** `priority = normal`
- **THEN** the breakdown `priority_fee` is 0

### Requirement: Percentage promotion applies a capped discount

When a valid active `Promotion` of type `percentage` is supplied, the engine SHALL compute `discount = round(total_before_discount * discount_value / 100)` and SHALL cap it at the promotion's `max_discount_cents` when that cap is greater than zero. The discount SHALL never exceed `total_before_discount`. The discount reduces the total and is recorded in the decision trace.

#### Scenario: Percentage discount under the cap
- **WHEN** `total_before_discount = 13500`, the promotion is `percentage` with `discount_value = 10` and `max_discount_cents = 5000`
- **THEN** the breakdown `discount` is 1350
- **AND** `total` is 12150

#### Scenario: Percentage discount limited by cap
- **WHEN** `total_before_discount = 100000`, the promotion is `percentage` with `discount_value = 50` and `max_discount_cents = 5000`
- **THEN** the breakdown `discount` is 5000

### Requirement: Fixed-amount promotion applies a flat discount

When a valid active `Promotion` of type `fixed_amount` is supplied, the engine SHALL apply a discount equal to the promotion's value in cents, capped so it never exceeds `total_before_discount`.

#### Scenario: Fixed discount applied
- **WHEN** `total_before_discount = 13500` and the promotion is `fixed_amount` worth 1000 cents
- **THEN** the breakdown `discount` is 1000
- **AND** `total` is 12500

#### Scenario: Fixed discount cannot exceed total
- **WHEN** `total_before_discount = 800` and the promotion is `fixed_amount` worth 1000 cents
- **THEN** the breakdown `discount` is 800
- **AND** `total` is 0

### Requirement: Quote is rejected when weight exceeds the rule limit

The engine SHALL reject the request with a domain error when `weight_kg` is greater than the rule's `max_weight_kg`.

#### Scenario: Weight over the limit
- **WHEN** `weight_kg = 35` and the rule's `max_weight_kg = 30`
- **THEN** the engine returns a domain error indicating the weight limit was exceeded
- **AND** no quote is produced

### Requirement: Quote is rejected when distance is outside the rule range

The engine SHALL reject the request with a domain error when `distance_km` is below the rule's `min_distance_km` or above its `max_distance_km`.

#### Scenario: Distance above the maximum
- **WHEN** `distance_km = 1500` and the rule's `max_distance_km = 1000`
- **THEN** the engine returns a domain error indicating the distance is out of range
- **AND** no quote is produced

### Requirement: Quote is rejected when a zone is inactive

The engine SHALL reject the request with a domain error when the origin or destination `Zone` is inactive.

#### Scenario: Destination zone inactive
- **WHEN** the destination zone has `is_active = false`
- **THEN** the engine returns a domain error indicating the zone is not available
- **AND** no quote is produced

### Requirement: Quote is rejected when no applicable rule is provided

The engine SHALL return a domain error indicating no applicable pricing rule when it is asked to calculate without an applicable rule.

#### Scenario: No applicable rule
- **WHEN** the engine is invoked with no applicable pricing rule for the route and shipping type
- **THEN** the engine returns a domain error indicating no rule applies
- **AND** no quote is produced

### Requirement: Invalid or expired promotion is rejected as a domain error

When a promotion is supplied but is inactive, not yet started, or already expired, the engine SHALL return a domain error indicating the promotion is invalid. The engine SHALL NOT silently ignore an explicitly supplied invalid promotion code.

#### Scenario: Expired promotion supplied
- **WHEN** a promotion is supplied whose `ends_at` is before the evaluation time
- **THEN** the engine returns a domain error indicating the promotion is invalid
- **AND** no quote is produced

#### Scenario: No promotion supplied
- **WHEN** no promotion is supplied
- **THEN** the quote is calculated normally with `discount = 0`
