## 1. Puertos y servicio

- [x] 1.1 `internal/services/ports.go`: interfaces `PricingRuleRepository`, `ZoneRepository`, `PromotionRepository` (definidas del lado del consumidor) + errores `ErrZoneNotFound`, `ErrPromotionNotFound`
- [x] 1.2 `internal/services/quote_service.go`: `QuoteService` con dependencias inyectadas y método `Quote(ctx, domain.QuoteRequest, now) (domain.Quote, error)`
- [x] 1.3 Resolución de zonas (origin/destination) vía `ZoneRepository`; zona no encontrada → error de negocio
- [x] 1.4 Selección de regla aplicable: pedir reglas activas candidatas y elegir la más específica (exacta > wildcard `*`); ninguna → `domain.ErrNoApplicableRule`
- [x] 1.5 Resolución de promoción si llega `promotion_code` (inexistente/expirada → `domain.ErrPromotionInvalid`); invocar `domain.Calculate`

## 2. Repositorios en memoria

- [x] 2.1 `internal/services/memory/store.go`: stores thread-safe (RWMutex) que implementan los tres puertos
- [x] 2.2 Seed inicial: zonas (CABA, CORDOBA_CAPITAL activas; una inactiva p/ tests), reglas standard/express CORDOBA_CAPITAL→CABA, promoción `SHIP10`

## 3. Handler HTTP y modelos

- [x] 3.1 `internal/handlers/request_models.go`: DTO del request de quote (con dimensiones opcionales) + mapeo a `domain.QuoteRequest`
- [x] 3.2 `internal/handlers/response_models.go`: DTOs de response (quote, breakdown, decision_trace, cached) desde `domain.Quote`
- [x] 3.3 `internal/handlers/validators.go`: validación de formato/negocio mínima (campos requeridos, distance/weight > 0, dimensiones > 0 si vienen, shipping_type y priority válidos)
- [x] 3.4 `internal/handlers/errors.go`: envelope de error (`code`, `message`, `details`, `request_id`) + helper `writeError` y mapeo de errores de dominio → status (400/422/500) usando `errors.Is`
- [x] 3.5 `internal/handlers/shipping_handler.go`: `ShippingHandler` con `QuoteService`; handler `POST /shipping/quote` (parse → validar → servicio → response/echo de errores con request_id)

## 4. Wiring

- [x] 4.1 `internal/server/router.go`: agregar `middleware.RequestID` y montar `POST /shipping/quote`
- [x] 4.2 `cmd/api/main.go`: construir stores en memoria + `QuoteService` + `ShippingHandler` e inyectarlos al router

## 5. Tests

- [x] 5.1 `internal/services/quote_service_test.go`: con repos fake/mock — éxito standard, regla más específica preferida, wildcard fallback, sin regla → error, zona inactiva → error, promo expirada → error
- [x] 5.2 `internal/handlers/shipping_handler_test.go` con `httptest`: 200 quote válido (breakdown reconcilia, cached=false), 400 JSON inválido, 400 weight_kg=0, 400 shipping_type inválido, 422 zona inactiva, 422 sin regla, 422 promo inválida
- [x] 5.3 `shipping_handler_test.go`: envelope de error tiene `code`/`message`/`request_id`; `X-Request-ID` entrante se refleja en la respuesta
- [x] 5.4 `internal/handlers/validators_test.go`: casos de validación unitarios

## 6. Verificación y cierre

- [x] 6.1 `go vet ./...` y `go test ./...` en verde
- [x] 6.2 Levantar `go run ./cmd/api` y probar `POST /shipping/quote` con un request válido (200 + breakdown) y uno inválido (400/422)
- [x] 6.3 `openspec validate add-quote-endpoint --strict` sin errores
- [x] 6.4 Confirmar con el usuario y archivar el change
