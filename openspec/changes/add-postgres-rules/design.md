## Context

El quote endpoint funciona contra repos en memoria detrás de los puertos `ZoneRepository`/`PricingRuleRepository`/`PromotionRepository`. Este change implementa esos puertos contra PostgreSQL y agrega la administración de reglas. Como los puertos ya existen, el `QuoteService` y su handler no se tocan para lectura; solo cambia el wiring en `main` y se extiende el puerto de reglas con operaciones de escritura.

## Goals / Non-Goals

**Goals:**
- Persistir zonas, reglas y promociones en PostgreSQL con migraciones versionadas y seed inicial.
- Implementar los puertos existentes con pgx, sin cambiar las firmas de lectura.
- CRUD de reglas (`GET`/`POST`/`PUT`) con validaciones, 409 por duplicado activo y 404 en update inexistente.
- Tests de integración con PostgreSQL real (Testcontainers), aislados del `go test ./...` por build tag.

**Non-Goals:**
- Cache Redis ni versionado de reglas para cache (llega en `add-redis-cache`; este change deja la base que aquella usará).
- Proveedor externo, `GET /shipping/options`.
- Borrado de reglas (`DELETE`) — fuera del MVP; se usa `is_active=false`.

## Decisions

**1. Driver: pgx v5 con pool (`pgxpool`).**
Driver moderno y performante. Se usa `pgxpool.Pool` inyectado a los repos. Las queries son SQL explícito (sin ORM) para mostrar control y mantener simple el mapeo registro→entidad de dominio.

**2. Migraciones con golang-migrate como librería + `cmd/migrate`.**
En lugar de depender del CLI `migrate` (no instalado), se usa golang-migrate como librería desde un binario `cmd/migrate` con subcomandos `up`/`down`. El `Makefile` cablea `migrate-up`/`migrate-down`/`seed` a `go run ./cmd/migrate ...`. Multiplataforma y sin instalar herramientas extra. Los archivos viven en `migrations/` con el formato `NNNNNN_name.up.sql` / `.down.sql`.

**3. Money en DB: columnas `*_cents BIGINT`.**
Consistente con el dominio (centavos `int64`). Los multiplicadores y límites (peso/distancia) son `NUMERIC`. El mapeo repo↔dominio convierte `BIGINT`→`domain.Money`.

**4. Extensión del puerto de reglas.**
`PricingRuleRepository` gana `List(ctx, filter) ([]RuleRecord, error)`, `GetByID`, `Create`, `Update`. Se introduce un tipo de regla con `ID` (UUID) para la administración — el dominio `PricingRule` no tiene ID (es valor de cálculo), así que la capa de repos/servicio usa un `RuleRecord` que envuelve `domain.PricingRule` + metadata (`ID`, `IsActive`, timestamps). Las lecturas para quote siguen devolviendo `[]domain.PricingRule` vía `FindActiveRules`.

**5. Detección de duplicado activo: índice único parcial.**
`CREATE UNIQUE INDEX ... ON pricing_rules (shipping_type, origin_zone_code, destination_zone_code) WHERE is_active`. El insert/update viola el índice ante duplicado activo; el repo traduce el error de constraint (`pgconn.PgError` código `23505`) a `repositories.ErrDuplicateRule`, que el handler mapea a 409. Defensa en DB, no solo en código.

**6. Validación de existencia de zonas.**
El `PricingRulesService` valida contra `ZoneRepository` que las zonas referenciadas existan (salvo `*`). Zona inexistente → error de negocio (422). El resto de validaciones (precios, multiplicadores, rangos) son de formato → 400.

**7. PostgreSQL como dependencia crítica de readiness.**
`main` registra en el `HealthRegistry` un check crítico que hace `pool.Ping(ctx)`. Si PostgreSQL está caído, `/ready` devuelve 503 (consistente con el plan). `/health` sigue sin tocar dependencias.

**8. Tests de integración con Testcontainers, build-tagged.**
Archivos `//go:build integration` en `tests/integration/`. Levantan `postgres:16`, corren migraciones, y prueban repos + flujo de quote. `make test` corre solo unitarios; `make test-integration` agrega el tag. Así el CI puede separar jobs y el dev no necesita Docker para el ciclo rápido.

**9. Persistencia de quotes best-effort.**
Nuevo puerto `QuoteRepository` con `Save(ctx, QuoteRecord) error`. El `QuoteService` lo invoca después de calcular, de forma **no bloqueante para el resultado**: si la escritura falla, se loguea (`warn`) y se devuelve igual el quote. Se inyecta opcionalmente (vía `WithPersistence(repo, logger)`); si no hay repo, no se persiste (mantiene los tests unitarios del servicio sin DB). Se guarda un `request_hash` (sha256 de la request normalizada — origen/destino/zonas/distancia/peso/tipo/prioridad/promo) que además será reutilizable por la cache en `add-redis-cache`. `was_cached=false` por ahora.

## Risks / Trade-offs

- **[Testcontainers en Windows: lento/dependiente de Docker]** → Primer arranque baja la imagen. Mitigación: build tag separado; documentar que requiere Docker corriendo; los unitarios no se ven afectados.
- **[App ahora requiere DB para arrancar]** → `go run ./cmd/api` solo falla rápido si no hay DB. Mitigación: flujo claro (`docker compose up` + `make migrate-up`); el check de readiness lo refleja como 503.
- **[Divergencia seed memoria vs SQL]** → Se mantiene el seed SQL como fuente para runs reales; el seed en memoria queda solo para tests unitarios del servicio. Documentado.
- **[Mapeo manual SQL↔dominio verboso]** → Aceptable; es explícito y demuestra control. Si creciera, se evaluaría sqlc.

## Open Questions

- ~~`shipping_quotes`~~: RESUELTO → se crea la tabla **y** se persiste cada quote de forma best-effort (decisión #9).
