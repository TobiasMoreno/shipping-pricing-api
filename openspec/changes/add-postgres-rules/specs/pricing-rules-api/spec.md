## ADDED Requirements

### Requirement: List pricing rules with optional filters

The service SHALL expose `GET /shipping/rules` that returns the pricing rules, optionally filtered by `shipping_type`, `origin_zone_code`, `destination_zone_code` and `active`. The response SHALL include each rule's id and configurable fields, and a `cached` flag.

#### Scenario: List all rules
- **WHEN** a client sends `GET /shipping/rules` with no filters
- **THEN** the response status is 200
- **AND** the body contains a `rules` array

#### Scenario: Filter by shipping type
- **WHEN** a client sends `GET /shipping/rules?shipping_type=express`
- **THEN** the response status is 200
- **AND** every returned rule has shipping type `express`

#### Scenario: Invalid filter value
- **WHEN** a client sends `GET /shipping/rules?shipping_type=drone`
- **THEN** the response status is 400

### Requirement: Create a pricing rule

The service SHALL expose `POST /shipping/rules` that creates a pricing rule and returns 201 with the new id. The request SHALL be validated: prices non-negative; multipliers greater than or equal to 1 where they apply; `max_distance_km` greater than `min_distance_km`; `max_weight_kg` greater than 0; and the referenced origin and destination zones must exist (unless the wildcard `*` is used). Validation failures return 400; references to non-existent zones return 422.

#### Scenario: Successful creation
- **WHEN** a client POSTs a valid pricing rule for existing zones
- **THEN** the response status is 201
- **AND** the body contains the new rule id

#### Scenario: Negative price rejected
- **WHEN** a client POSTs a rule with a negative `base_price`
- **THEN** the response status is 400

#### Scenario: Distance range inverted
- **WHEN** a client POSTs a rule whose `max_distance_km` is not greater than its `min_distance_km`
- **THEN** the response status is 400

#### Scenario: Unknown zone referenced
- **WHEN** a client POSTs a rule referencing an origin zone code that does not exist
- **THEN** the response status is 422

### Requirement: Reject duplicate active rules

The service SHALL NOT allow two active rules for the same exact combination of shipping type, origin zone and destination zone. An attempt to create such a duplicate SHALL return 409.

#### Scenario: Duplicate active combination
- **WHEN** an active rule already exists for (standard, CORDOBA_CAPITAL, CABA)
- **AND** a client POSTs another active rule for the same combination
- **THEN** the response status is 409

### Requirement: Update a pricing rule

The service SHALL expose `PUT /shipping/rules/{id}` that updates an existing rule's mutable fields and returns 200. A non-existent id SHALL return 404. The same validation rules as creation apply, and an update that produces a duplicate active combination SHALL return 409.

#### Scenario: Successful update
- **WHEN** a client PUTs valid changes for an existing rule id
- **THEN** the response status is 200

#### Scenario: Update missing rule
- **WHEN** a client PUTs changes for an id that does not exist
- **THEN** the response status is 404

#### Scenario: Invalid id format
- **WHEN** a client PUTs to `/shipping/rules/{id}` with a malformed id
- **THEN** the response status is 400
