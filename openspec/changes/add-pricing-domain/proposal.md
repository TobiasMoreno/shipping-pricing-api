## Why

El corazón de `shipping-pricing-api` es la lógica de cálculo de precios de envío, no el CRUD. Empezar por el dominio puro (sin HTTP, DB ni Redis) permite testear el negocio de forma aislada, fijar contratos de cálculo claros y construir el resto de las capas sobre una base estable y defendible. Es el primer entregable del roadmap y el que más diferencia al proyecto de un CRUD genérico.

## What Changes

- Se introduce el paquete de dominio (`internal/domain`) con las entidades y value objects centrales: `QuoteRequest`, `Quote`, `PricingRule`, `Zone`, `Promotion`, `Money`, `PricingBreakdown`, `PricingDecisionTrace`.
- Se implementa el **motor de pricing puro**: una función/servicio de dominio que, dado un `QuoteRequest` y la regla aplicable (más zonas y promoción opcional), produce un `Quote` con precio final, breakdown detallado y decision trace.
- Se define el modelo de dinero en **centavos `int64`** (`Money`) — sin floats en montos.
- Se definen los **errores de dominio** explícitos (regla no aplicable, peso excedido, distancia fuera de rango, zona inactiva, promoción inválida/expirada).
- Se agregan **tests unitarios** de pricing que cubren cada componente del cálculo y los casos borde de negocio.
- NO incluye: handlers HTTP, persistencia, cache, cliente de proveedor. Esas capas consumirán este dominio en changes posteriores.

## Capabilities

### New Capabilities
- `pricing-engine`: cálculo determinístico del precio de un envío a partir de un request y una regla de pricing aplicable. Cubre precio base por tipo de envío, fee por distancia, fee por peso, multiplicadores (express/prioridad), aplicación de promociones (porcentual y monto fijo) con tope, restricciones de negocio (peso máximo, rango de distancia, zona inactiva), redondeo en centavos, breakdown que suma exacto al total y decision trace de las decisiones tomadas.

### Modified Capabilities
<!-- Ninguna: no existen specs previas. -->

## Impact

- **Código nuevo**: `internal/domain/` (`quote.go`, `pricing_rule.go`, `zone.go`, `promotion.go`, `money.go`, `errors.go`, `pricing.go`) y `internal/domain/pricing_test.go`.
- **Módulo Go**: inicializa `go.mod` (module path a confirmar) con Go 1.25. Sin dependencias externas en esta etapa (solo stdlib + testing).
- **Sin impacto** en infraestructura, APIs HTTP ni datos persistidos todavía.
- Establece los contratos (tipos y semántica de cálculo) que consumirán los changes `add-quote-endpoint`, `add-postgres-rules` y `add-redis-cache`.
