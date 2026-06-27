## Context

Este es el primer change del proyecto: no hay código Go todavía, solo el plan y la estructura de OpenSpec. La capa de dominio debe ser pura (sin HTTP, DB, Redis ni cliente de proveedor) para poder testearse en aislamiento y servir de contrato a las capas superiores. Las reglas, zonas y promociones llegan al motor ya resueltas (en changes posteriores las traerán repositorios y cache); este change no decide cómo se obtienen, solo cómo se calcula con ellas.

## Goals / Non-Goals

**Goals:**
- Definir las entidades y value objects del dominio con tipos correctos (centavos `int64`, sin floats en montos).
- Implementar un motor de pricing determinístico, sin estado y sin efectos colaterales.
- Producir un breakdown que reconcilie exactamente con el total y un decision trace legible.
- Cubrir con tests unitarios cada componente del cálculo y cada caso borde de negocio.

**Non-Goals:**
- Selección de la regla aplicable a partir de un conjunto (matching por zona/tipo/rango) — eso vive en el servicio de quote (`add-quote-endpoint`). Este change asume que recibe la regla ya elegida.
- Persistencia, serialización JSON de la API, cache, validación de input HTTP.
- Cálculo de ETA / disponibilidad del proveedor (va en `add-logistics-provider`). El campo `available`/`estimated_delivery_days` del `Quote` se completa aguas arriba; el dominio expone el tipo pero no consulta proveedor.

## Decisions

**1. `Money` como `int64` de centavos.**
Se modela `type Money int64` (cents) con helpers (`Add`, `Sub`, `MulFloat` con redondeo half-up, `Cents()`). Alternativa descartada: `float64` o `decimal` de tercero. Centavos enteros eliminan errores de redondeo de punto flotante y mantienen el dominio sin dependencias. Cuando se multiplica por un factor (distancia, peso, multiplicadores), se hace en `float64` solo para el cómputo intermedio y se redondea inmediatamente a `int64` con round-half-up.

**2. Multiplicadores reportados como fee absoluto.**
La regla guarda multiplicadores (`express_multiplier`, `priority_multiplier`) pero el breakdown los expone como montos (`express_fee`, `priority_fee`). El surcharge se calcula sobre el subtotal `base + distance + weight`: `fee = round(subtotal * (multiplier - 1))`. Así el breakdown suma exacto al total y es legible. Alternativa descartada: exponer el multiplicador crudo (menos legible, no suma).

**3. Express vs same_day.**
`express_multiplier` aplica solo cuando `shipping_type == express`. `same_day` se calcula con los componentes base (su disponibilidad y ETA se resuelven con el proveedor en un change posterior). Esto evita inventar semántica de pricing para same_day antes de tiempo.

**4. Orden de cálculo fijo y trazado.**
base → distance_fee → weight_fee → express_fee → priority_fee → total_before_discount → discount → total. Cada paso agrega una línea al `PricingDecisionTrace` (slice de strings). El orden es determinístico para que los tests y el trace sean estables.

**5. Reloj inyectable para validar promociones.**
La validez de una promoción (`starts_at`/`ends_at`/`is_active`) depende del tiempo actual. Para mantener el dominio puro y testeable, el momento de evaluación se pasa como parámetro (`now time.Time`) en lugar de llamar a `time.Now()` dentro del motor. Alternativa descartada: leer el reloj global (tests no determinísticos).

**6. Errores de dominio tipados.**
Errores centinela en `internal/domain/errors.go` (`ErrNoApplicableRule`, `ErrWeightExceeded`, `ErrDistanceOutOfRange`, `ErrZoneInactive`, `ErrPromotionInvalid`), comparables con `errors.Is`. Las capas superiores los mapearán a status HTTP (422, etc.) en `add-quote-endpoint`. El dominio no conoce códigos HTTP.

**7. Firma del motor.**
`func Calculate(req QuoteRequest, rule PricingRule, origin, destination Zone, promo *Promotion, now time.Time) (Quote, error)`. Promoción opcional vía puntero nil. Función libre en el paquete `domain` (sin estado), no un método de struct.

## Risks / Trade-offs

- **[Acoplamiento del orden de redondeo]** → Redondear en cada paso vs al final puede dar diferencias de 1 centavo. Decisión: redondear cada fee al calcularlo (los fees son montos reportados en el breakdown y deben ser enteros). Documentado en los scenarios de la spec con ejemplos concretos.
- **[`available`/`estimated_delivery_days` sin fuente en este change]** → El `Quote` los expone pero el dominio no los calcula. Mitigación: en este change se setean a valores por defecto (`available=true`, ETA desde un campo placeholder de la regla o 0) y se documenta que su fuente real llega con el proveedor. No se testea su valor "real" todavía.
- **[Module path de Go aún sin definir]** → Bloquea `go mod init`. Mitigación: resolver en Open Questions antes de `apply`.

## Open Questions

- ~~**Module path**~~: RESUELTO → `github.com/TobiasMoreno/shipping-pricing-api`.
- **same_day**: ¿requiere su propio multiplicador en la regla a futuro, o se gobierna 100% por disponibilidad del proveedor? Se decide en `add-logistics-provider`.
