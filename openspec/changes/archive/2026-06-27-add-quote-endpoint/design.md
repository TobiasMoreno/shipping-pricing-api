## Context

El dominio (`pricing-engine`) y el runtime HTTP ya existen. Este change conecta ambos con la primera ruta de negocio y la capa de servicios, usando repositorios en memoria detrás de interfaces. El objetivo es validar el flujo completo y dejar los puertos definidos para que `add-postgres-rules` (persistencia) y `add-redis-cache` (cache) los reemplacen/envuelvan sin tocar el servicio ni el handler.

## Goals / Non-Goals

**Goals:**
- `POST /shipping/quote` end-to-end: handler valida → servicio orquesta → dominio calcula → response JSON.
- Puertos (`PricingRuleRepository`, `ZoneRepository`, `PromotionRepository`) definidos del lado del consumidor (paquete services).
- Mapeo limpio de errores de dominio a status HTTP y envelope de error consistente.
- Selección de regla aplicable testeable de forma aislada.

**Non-Goals:**
- Persistencia real (repos en memoria con seed; Postgres llega en `add-postgres-rules`).
- Cache (`cached` siempre false por ahora; Redis llega en `add-redis-cache`).
- Proveedor externo / ETA real (`estimated_delivery_days` se deriva de la regla o queda en 0; el proveedor llega en `add-logistics-provider`).
- `GET /shipping/options`, CRUD de reglas, logging estructurado avanzado.

## Decisions

**1. Puertos del lado del consumidor.**
Las interfaces viven en `internal/services/ports.go`, no en el paquete de repositorios. El servicio define lo que necesita; las implementaciones (memoria ahora, Postgres después) dependen de esa abstracción. Esto sigue la guía del plan y permite mockear en tests de servicio.

```go
type PricingRuleRepository interface {
    FindActiveRules(ctx context.Context, shippingType domain.ShippingType, origin, destination string) ([]domain.PricingRule, error)
}
type ZoneRepository interface {
    GetByCode(ctx context.Context, code string) (domain.Zone, error)
}
type PromotionRepository interface {
    GetByCode(ctx context.Context, code string) (domain.Promotion, error)
}
```

**2. Selección de regla en el servicio, no en el repo.**
El repo devuelve las reglas activas candidatas (match de tipo + zona, incluyendo wildcard `*`); el servicio elige la más específica (match exacto de ambas zonas > una exacta > wildcard). Mantener el matching en el servicio lo hace testeable sin DB y deja el repo como simple acceso a datos.

**3. Mapeo de errores dominio → HTTP en el handler.**
El servicio devuelve errores de dominio (`errors.Is`) sin conocer HTTP. El handler los traduce: `ErrNoApplicableRule`, `ErrWeightExceeded`, `ErrDistanceOutOfRange`, `ErrZoneInactive`, `ErrPromotionInvalid` → 422; errores de zona/promo "no encontrada" del repo → 422; validación de formato → 400; cualquier otro → 500. Un helper `writeError(w, r, status, code, message, details)` arma el envelope.

**4. Envelope de error y request id.**
Se usa `chi/middleware.RequestID` (setea contexto + header `X-Request-ID`, reusando el entrante). El handler obtiene el id con `middleware.GetReqID(ctx)` y lo incluye en el envelope. La observabilidad completa (logging estructurado correlacionado) es de `add-observability`; acá solo poblar el id.

**5. Resolución de zonas y promoción.**
El servicio resuelve `origin_zone_code` y `destination_zone_code` vía `ZoneRepository`; si el repo no las encuentra, es error de negocio (422). Si llega `promotion_code`, lo resuelve vía `PromotionRepository`; código inexistente o expirado → `ErrPromotionInvalid` (422), consistente con la decisión del plan de fallar ante promo inválida explícita. Sin `promotion_code` → sin descuento.

**6. Repos en memoria con seed.**
`internal/services/memory/` con stores thread-safe (map + RWMutex) y un seed inicial (zonas CABA/CORDOBA_CAPITAL activas y una inactiva para tests; reglas standard/express; promo `SHIP10`). Se construyen en `main.go` e inyectan al servicio.

## Risks / Trade-offs

- **[`estimated_delivery_days` sin fuente real]** → Se deriva de un campo simple por tipo de envío o queda en 0 hasta `add-logistics-provider`. Mitigación: documentado; no se testea su valor "real".
- **[Seed en memoria diverge del seed SQL futuro]** → Cuando llegue Postgres habrá que alinear seeds. Mitigación: mantener el seed en un único lugar por implementación y cubrir el flujo con tests de servicio/handler que no dependan de valores mágicos del seed salvo lo necesario.
- **[Matching de reglas simplista]** → "Más específica" se limita a exactitud de zonas; no contempla prioridad configurable. Aceptable para MVP; ampliable después.

## Open Questions

- ¿`estimated_delivery_days` en este change: 0, o un default por shipping type (standard=3, express=1, same_day=0)? Propuesta: default por tipo de envío para que la demo se vea completa, marcándolo como placeholder hasta el proveedor.
