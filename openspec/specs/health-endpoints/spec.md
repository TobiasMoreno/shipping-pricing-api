# health-endpoints Specification

## Purpose
TBD - created by archiving change add-http-skeleton. Update Purpose after archive.
## Requirements
### Requirement: Liveness endpoint responds without checking dependencies

The service SHALL expose `GET /health` that returns HTTP 200 with a JSON body indicating the process is alive. This endpoint MUST NOT query PostgreSQL, Redis or any external dependency, so it reflects only whether the HTTP process is running.

#### Scenario: Health returns ok
- **WHEN** a client sends `GET /health`
- **THEN** the response status is 200
- **AND** the JSON body contains `"status": "ok"`

#### Scenario: Health ignores failing dependencies
- **WHEN** a downstream dependency (e.g. PostgreSQL) is unavailable
- **AND** a client sends `GET /health`
- **THEN** the response status is still 200

### Requirement: Readiness endpoint aggregates registered dependency checks

The service SHALL expose `GET /ready` that evaluates all registered dependency checks and returns their individual statuses in the JSON body under `dependencies`. Each dependency is registered as either critical or non-critical. When any critical dependency check fails, the endpoint SHALL return HTTP 503. When all critical dependencies pass, the endpoint SHALL return HTTP 200 even if some non-critical dependencies report `degraded`.

#### Scenario: All dependencies healthy
- **WHEN** every registered dependency check passes
- **AND** a client sends `GET /ready`
- **THEN** the response status is 200
- **AND** the body reports `"status": "ready"`

#### Scenario: Critical dependency down returns 503
- **WHEN** a dependency registered as critical fails its check
- **AND** a client sends `GET /ready`
- **THEN** the response status is 503
- **AND** the failing dependency is reported as not ok in the body

#### Scenario: Non-critical dependency degraded still ready
- **WHEN** a dependency registered as non-critical fails its check
- **AND** all critical dependencies pass
- **AND** a client sends `GET /ready`
- **THEN** the response status is 200
- **AND** the non-critical dependency is reported as `degraded` in the body

#### Scenario: No dependencies registered
- **WHEN** no dependency checks are registered
- **AND** a client sends `GET /ready`
- **THEN** the response status is 200

