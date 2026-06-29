## 1. Dependencias

- [ ] 1.1 `go get github.com/jackc/pgx/v5`, `github.com/golang-migrate/migrate/v4`, `github.com/testcontainers/testcontainers-go` + `go mod tidy`

## 2. Migraciones y seed

- [ ] 2.1 `migrations/000001_create_shipping_zones.{up,down}.sql`
- [ ] 2.2 `migrations/000002_create_pricing_rules.{up,down}.sql` con índice único parcial `(shipping_type, origin_zone_code, destination_zone_code) WHERE is_active`
- [ ] 2.3 `migrations/000003_create_promotions.{up,down}.sql`
- [ ] 2.4 `migrations/000004_create_shipping_quotes.{up,down}.sql`
- [ ] 2.5 `migrations/000005_seed_data.{up,down}.sql`: zonas (CABA, CORDOBA_CAPITAL, PATAGONIA inactiva), reglas standard/express + wildcard, promo SHIP10
- [ ] 2.6 `cmd/migrate/main.go`: subcomandos `up`/`down` usando golang-migrate (file source + pgx) leyendo `DATABASE_URL`

## 3. Capa de persistencia (pgx)

- [ ] 3.1 `internal/repositories/postgres/db.go`: construir `pgxpool.Pool` desde `DATABASE_URL` + `Ping`
- [ ] 3.2 `internal/repositories/errors.go`: `ErrNotFound`, `ErrDuplicateRule` (+ helper para detectar `23505` de pgconn)
- [ ] 3.3 `internal/repositories/postgres/zone_repository.go`: `GetByCode` (implementa `ZoneRepository`)
- [ ] 3.4 `internal/repositories/postgres/promotion_repository.go`: `GetByCode` (implementa `PromotionRepository`)
- [ ] 3.5 `internal/repositories/postgres/pricing_rule_repository.go`: `FindActiveRules` + `List(filter)` + `GetByID` + `Create` + `Update`, mapeando duplicado activo a `ErrDuplicateRule`
- [ ] 3.6 `internal/repositories/postgres/quote_repository.go`: `Save(ctx, QuoteRecord)` (implementa `QuoteRepository`)

## 4. Puerto extendido y servicio de reglas

- [ ] 4.1 `internal/services/ports.go`: extender `PricingRuleRepository` con `List`/`GetByID`/`Create`/`Update`; definir `RuleRecord` (ID + domain.PricingRule + IsActive + timestamps) y `RuleFilter`
- [ ] 4.2 Actualizar el store en memoria para satisfacer el puerto extendido (mantiene tests unitarios del QuoteService compilando)
- [ ] 4.3 `internal/services/pricing_rules_service.go`: `List`, `Create` (valida precios≥0, multiplicadores≥1, max>min, peso>0, zonas existen, duplicado), `Update`
- [ ] 4.4 `internal/services`: definir puerto `QuoteRepository` + `QuoteRecord` + helper `requestHash`; agregar `WithPersistence(repo, logger)` al `QuoteService` y persistir best-effort tras calcular (fallo → log warn, no rompe respuesta)

## 5. Handlers de reglas

- [ ] 5.1 `internal/handlers/rules_handler.go`: DTOs + `GET /shipping/rules` (parsea filtros, 400 si inválidos), `POST` (201/400/409/422), `PUT /{id}` (200/400/404/409)
- [ ] 5.2 Extender el mapeo de errores: `repositories.ErrDuplicateRule` → 409, `repositories.ErrNotFound` → 404

## 6. Wiring

- [ ] 6.1 `cmd/api/main.go`: construir `pgxpool` + repos PG, inyectar a `QuoteService` (con `WithPersistence`) y `PricingRulesService`; registrar check crítico de PostgreSQL en `/ready`; cerrar el pool en shutdown
- [ ] 6.2 `internal/server/router.go`: montar rutas de reglas
- [ ] 6.3 `Makefile`: `migrate-up`/`migrate-down`/`seed` → `go run ./cmd/migrate ...`

## 7. Tests

- [ ] 7.1 `internal/services/pricing_rules_service_test.go`: validaciones (precio negativo, max≤min, zona inexistente, duplicado) con repos fake
- [ ] 7.2 `internal/handlers/rules_handler_test.go` (httptest + servicio stub): 201 crear, 400 inválido, 409 duplicado, 404 update inexistente, 400 id malformado, filtros de GET
- [ ] 7.3 `tests/integration/postgres_test.go` (`//go:build integration`, Testcontainers): migraciones corren; repo crea/lista/actualiza reglas; duplicado activo → error
- [ ] 7.4 `tests/integration/quote_flow_test.go` (`//go:build integration`): flujo de quote completo contra PostgreSQL real + verificar que el quote quedó persistido en `shipping_quotes`

## 8. Verificación y cierre

- [ ] 8.1 `go vet ./...` y `go test ./...` (unitarios) en verde
- [ ] 8.2 `docker compose up -d postgres`, `make migrate-up`, `go test -tags=integration ./tests/integration/...` en verde
- [ ] 8.3 Levantar la app contra Postgres y probar `POST /shipping/rules`, `GET /shipping/rules`, `PUT /shipping/rules/{id}` y un `POST /shipping/quote` que use la regla creada
- [ ] 8.4 `openspec validate add-postgres-rules --strict` sin errores
- [ ] 8.5 Confirmar con el usuario y archivar el change
