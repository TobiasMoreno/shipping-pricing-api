## ADDED Requirements

### Requirement: Quote endpoint calculates a shipment price

The service SHALL expose `POST /shipping/quote` that accepts a JSON request describing the shipment (origin/destination zip and zone codes, distance, weight, optional dimensions, shipping type, priority, optional promotion code), resolves the applicable rule, zones and promotion, and returns HTTP 200 with the calculated quote. The response SHALL include the final price in cents, currency, estimated delivery days, shipping type, availability, a `cached` flag, the pricing breakdown and the decision trace.

#### Scenario: Successful standard quote
- **WHEN** a client POSTs a valid standard quote request for an active route with a matching active rule
- **THEN** the response status is 200
- **AND** the body contains a positive `price`, a `breakdown` whose components reconcile to the total, and a non-empty `decision_trace`
- **AND** the body reports `cached: false`

### Requirement: Quote requests are validated before processing

The endpoint SHALL reject malformed or invalid requests with HTTP 400 before any pricing logic runs. Validation MUST cover: malformed JSON; missing required origin/destination zip and zone codes; `distance_km` not greater than 0; `weight_kg` not greater than 0; any informed dimension not greater than 0; `shipping_type` outside {standard, express, same_day}; and `priority` outside {normal, high}.

#### Scenario: Malformed JSON body
- **WHEN** a client POSTs a body that is not valid JSON
- **THEN** the response status is 400
- **AND** the error envelope has code `invalid_request`

#### Scenario: Non-positive weight
- **WHEN** a client POSTs a request with `weight_kg` of 0
- **THEN** the response status is 400
- **AND** the error details reference the `weight_kg` field

#### Scenario: Invalid shipping type
- **WHEN** a client POSTs a request with `shipping_type` set to `drone`
- **THEN** the response status is 400

### Requirement: Business rule failures return 422

When a request is well-formed but cannot be priced for a business reason, the endpoint SHALL return HTTP 422 with an error envelope describing the reason. This MUST cover: no applicable pricing rule, weight exceeding the rule limit, distance outside the rule range, an inactive origin or destination zone, and an explicitly supplied invalid or expired promotion code.

#### Scenario: No applicable rule
- **WHEN** a client POSTs a valid request for a route/shipping-type combination with no active rule
- **THEN** the response status is 422

#### Scenario: Inactive destination zone
- **WHEN** a client POSTs a valid request whose destination zone is inactive
- **THEN** the response status is 422

#### Scenario: Invalid promotion code
- **WHEN** a client POSTs a valid request with a promotion code that does not exist or has expired
- **THEN** the response status is 422

### Requirement: The most specific active rule is selected

When resolving the applicable rule, the service SHALL consider only active rules matching the requested shipping type and route, treating `*` as a zone wildcard, and SHALL prefer a rule with exact zone matches over a rule that matches via wildcard.

#### Scenario: Exact rule preferred over wildcard
- **WHEN** both an exact-zone active rule and a wildcard active rule match the request
- **THEN** the exact-zone rule is used to price the quote

#### Scenario: Wildcard rule used as fallback
- **WHEN** only a wildcard active rule matches the request
- **THEN** the wildcard rule is used to price the quote

### Requirement: Errors use a consistent envelope with a request id

All error responses SHALL use a consistent JSON envelope with an `error` object containing `code`, `message`, an optional `details` array of field/reason pairs, and a `request_id`. Every response SHALL carry an `X-Request-ID` header, reusing an incoming `X-Request-ID` when present and generating one otherwise.

#### Scenario: Error envelope shape
- **WHEN** any request fails validation or business rules
- **THEN** the body has an `error` object with `code`, `message` and `request_id`
- **AND** the `request_id` matches the `X-Request-ID` response header

#### Scenario: Incoming request id is reused
- **WHEN** a client sends a request with an `X-Request-ID` header
- **THEN** the response `X-Request-ID` header equals the value sent by the client
