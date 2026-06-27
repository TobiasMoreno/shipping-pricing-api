## Why

Tenemos el motor de pricing puro y un servidor HTTP que corre, pero ningún endpoint de negocio. `POST /shipping/quote` es el caso de uso principal del producto: es lo que demuestra valor en una demo y en una entrevista. Construirlo ahora con reglas en memoria (detrás de interfaces) permite probar el flujo completo HTTP → servicio → dominio rápido, sin la complejidad de Postgres/Redis, y deja los puertos listos para enchufar persistencia y cache en los changes siguientes.

## What Changes

- Se agrega `POST /shipping/quote`: valida el request, orquesta la resolución de zonas/regla/promoción, invoca `domain.Calculate` y devuelve el quote con precio, ETA, breakdown, decision trace y `cached: false`.
- Se introduce la capa de servicios (`internal/services`): `QuoteService` que coordina el caso de uso, y `ports.go` con las interfaces que consume (`PricingRuleRepository`, `ZoneRepository`, `PromotionRepository`) definidas del lado del consumidor.
- Se agregan implementaciones **en memoria** de esos puertos con seed data (zonas, reglas, una promoción), reemplazables por Postgres en `add-postgres-rules` sin tocar el servicio.
- Selección de regla aplicable: dado el request, se elige la regla activa más específica (match exacto de zona por sobre wildcard `*`) que aplique al tipo de envío; si ninguna aplica → error de negocio.
- Se agrega validación de request (`internal/handlers/validators.go`) y DTOs de request/response.
- Se agrega un **envelope de error consistente** (`code`, `message`, `details`, `request_id`) y el mapeo de errores de dominio a status HTTP (400 formato, 422 negocio, 500 inesperado).
- Se incorpora `chi/middleware.RequestID` para poblar `request_id` en respuestas y errores (la observabilidad completa llega en `add-observability`).
- NO incluye: persistencia real, cache, proveedor externo, `GET /shipping/options`, endpoints de reglas (CRUD de reglas llega en `add-postgres-rules`).

## Capabilities

### New Capabilities
- `quote-api`: el endpoint `POST /shipping/quote` — forma del request, reglas de validación (400), respuesta exitosa con breakdown/decision-trace/flag de cache (200), selección de la regla aplicable, y mapeo de errores de negocio a 422 con envelope consistente.

### Modified Capabilities
<!-- Ninguna; se reutiliza pricing-engine sin cambiar sus requirements. -->

## Impact

- **Código nuevo**: `internal/services/{quote_service.go,ports.go}`, `internal/services/memory/` (repos en memoria + seed), `internal/handlers/{shipping_handler.go,request_models.go,response_models.go,validators.go,errors.go}`.
- **Modificado**: `internal/server/router.go` (monta la ruta y `middleware.RequestID`), `cmd/api/main.go` (construye repos en memoria + `QuoteService` y los inyecta).
- **Sin dependencias nuevas** (se usa chi ya presente + stdlib).
- Establece los puertos (`PricingRuleRepository`, `ZoneRepository`, `PromotionRepository`) que `add-postgres-rules` implementará contra PostgreSQL y `add-redis-cache` envolverá con cache-aside.
