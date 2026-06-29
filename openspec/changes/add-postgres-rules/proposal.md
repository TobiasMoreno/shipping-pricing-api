## Why

El endpoint de quote ya funciona pero contra reglas/zonas en memoria con seed hardcodeado. Para ser un backend real necesitamos persistencia: reglas y zonas en PostgreSQL, editables vía API. Esto habilita los endpoints administrativos de reglas (lo que demuestra CRUD + validaciones + manejo de conflictos) y deja la base lista para la cache (que versiona e invalida sobre cambios de reglas).

## What Changes

- Se agregan **migraciones** versionadas (golang-migrate) para `shipping_zones`, `pricing_rules`, `promotions` y `shipping_quotes`, con un seed inicial equivalente al de memoria.
- Se agrega `internal/repositories/postgres`: pool de conexión pgx y repositorios que implementan los puertos existentes (`ZoneRepository`, `PricingRuleRepository`, `PromotionRepository`) **sin cambiar sus firmas de lectura**.
- Se extiende `PricingRuleRepository` con operaciones de administración (listar con filtros, obtener por id, crear, actualizar) y se agrega `PricingRulesService` que las orquesta con validaciones de negocio.
- Se agregan endpoints: `GET /shipping/rules` (filtros opcionales), `POST /shipping/rules` (crear, con validación y 409 ante regla activa duplicada), `PUT /shipping/rules/{id}` (actualizar, 404 si no existe).
- Se agrega un comando `cmd/migrate` (usa golang-migrate como librería) para correr/revertir migraciones sin depender del CLI; el `Makefile` cablea `migrate-up`/`migrate-down`/`seed` a él.
- `cmd/api/main.go` pasa a construir repos PostgreSQL desde `DATABASE_URL` e inyectarlos al `QuoteService` y al nuevo `PricingRulesService`. Se registra el check de PostgreSQL como dependencia **crítica** en `/ready`.
- Se persiste cada quote calculado en `shipping_quotes` de forma **best-effort** (un fallo de escritura se loguea pero no rompe la respuesta): nuevo puerto `QuoteRepository` y su implementación PostgreSQL, inyectado al `QuoteService`.
- Se agregan **tests de integración** (build tag `integration`) con Testcontainers levantando PostgreSQL real: migraciones, CRUD de reglas, persistencia de quote, y flujo de quote contra DB.
- Nuevas dependencias: `jackc/pgx/v5`, `golang-migrate/migrate/v4`, `testcontainers/testcontainers-go`.
- NO incluye: cache Redis (el `cached` sigue en false), proveedor externo, `GET /shipping/options`.

## Capabilities

### New Capabilities
- `pricing-rules-api`: endpoints administrativos de reglas — `GET /shipping/rules` con filtros, `POST /shipping/rules` con validaciones de negocio y detección de regla activa duplicada (409), `PUT /shipping/rules/{id}` con 404 ante regla inexistente; envelope de error consistente.

### Modified Capabilities
<!-- Ninguna a nivel de requirements. El cálculo del quote pasa a resolver reglas/zonas/promociones desde PostgreSQL, pero el contrato HTTP observable de `quote-api` no cambia (mismas requests/responses/status), así que no hay delta de spec; se documenta como impacto de implementación. -->



## Impact

- **Código nuevo**: `migrations/*.sql`, `cmd/migrate/main.go`, `internal/repositories/postgres/{db.go,zone_repository.go,pricing_rule_repository.go,promotion_repository.go,quote_repository.go}`, `internal/repositories/errors.go`, `internal/services/pricing_rules_service.go`, `internal/handlers/rules_handler.go`, `tests/integration/*`.
- **Modificado**: `internal/services/ports.go` (extensión de `PricingRuleRepository` con CRUD), `cmd/api/main.go` (repos PG + check de readiness), `internal/server/router.go` (rutas de reglas), `Makefile` (targets de migración apuntando a `cmd/migrate`).
- **Dependencias**: agrega pgx, golang-migrate y testcontainers.
- **Requiere PostgreSQL para arrancar** (dependencia crítica). Local: `docker compose up` + `make migrate-up`. Los tests unitarios siguen sin necesitar Docker; los de integración corren con build tag `integration`.
- Los repos en memoria (`internal/services/memory`) se conservan para tests unitarios del servicio; dejan de usarse en `main`.
