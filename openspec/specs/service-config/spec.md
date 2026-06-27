# service-config Specification

## Purpose
TBD - created by archiving change add-http-skeleton. Update Purpose after archive.
## Requirements
### Requirement: Configuration is loaded from environment variables

The application SHALL read its configuration from environment variables into a typed configuration struct at startup. At minimum it SHALL support `APP_ENV`, `HTTP_PORT`, `LOG_LEVEL`, `DATABASE_URL`, `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`, `CACHE_ENABLED`, and the cache TTL and provider settings used by later changes.

#### Scenario: Values read from the environment
- **WHEN** `HTTP_PORT=9000` is set in the environment and configuration is loaded
- **THEN** the typed configuration exposes the HTTP port as 9000

### Requirement: Sensible defaults for optional values

The configuration loader SHALL apply documented default values for optional settings when their environment variables are absent, so the service can run locally with minimal setup.

#### Scenario: Default HTTP port applied
- **WHEN** `HTTP_PORT` is not set
- **THEN** the configuration uses the default HTTP port 8080

#### Scenario: Cache enabled by default
- **WHEN** `CACHE_ENABLED` is not set
- **THEN** the configuration reports cache as enabled

### Requirement: Required configuration is validated at startup

The configuration loader SHALL validate that required settings are present and well-formed, and SHALL return an error (causing the process to fail fast at startup) when a required value is missing or invalid.

#### Scenario: Missing required value fails fast
- **WHEN** a required configuration value (e.g. `DATABASE_URL`) is missing or empty
- **THEN** configuration loading returns an error
- **AND** the process does not start serving traffic

#### Scenario: Invalid numeric value rejected
- **WHEN** `HTTP_PORT` is set to a non-numeric value
- **THEN** configuration loading returns an error

