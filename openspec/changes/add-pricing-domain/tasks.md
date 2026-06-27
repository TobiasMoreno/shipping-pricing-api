## 1. Module setup

- [x] 1.1 Correr `go mod init github.com/TobiasMoreno/shipping-pricing-api` (Go 1.25)
- [x] 1.2 Agregar `.gitignore` para Go (binarios, `*.test`, cobertura) si no existe

## 2. Value objects y entidades de dominio

- [x] 2.1 Implementar `internal/domain/money.go`: `type Money int64` con `Add`, `Sub`, helper de multiplicación por factor con round-half-up a centavos, y `Cents()`
- [x] 2.2 Implementar `internal/domain/zone.go`: `Zone` (`Code`, `Name`, `Region`, `IsActive`)
- [x] 2.3 Implementar `internal/domain/promotion.go`: `Promotion` (`Code`, `DiscountType` {percentage|fixed_amount}, `DiscountValue`, `MaxDiscount Money`, `StartsAt`, `EndsAt`, `IsActive`) y método `IsValidAt(now time.Time) bool`
- [x] 2.4 Implementar `internal/domain/pricing_rule.go`: `PricingRule` (`ShippingType`, zonas, `BasePrice Money`, `PricePerKm Money`, `PricePerKg Money`, `ExpressMultiplier`, `PriorityMultiplier`, `MaxWeightKg`, `MinDistanceKm`, `MaxDistanceKm`, `IsActive`)
- [x] 2.5 Implementar `internal/domain/quote.go`: `QuoteRequest`, `Quote`, `PricingBreakdown` (todos los campos en `Money`/cents) y `PricingDecisionTrace` ([]string con helper de append)
- [x] 2.6 Definir constantes/tipos para `ShippingType` (standard|express|same_day) y `Priority` (normal|high)

## 3. Errores de dominio

- [x] 3.1 Implementar `internal/domain/errors.go` con centinelas: `ErrNoApplicableRule`, `ErrWeightExceeded`, `ErrDistanceOutOfRange`, `ErrZoneInactive`, `ErrPromotionInvalid` (comparables con `errors.Is`)

## 4. Motor de pricing

- [x] 4.1 Implementar `internal/domain/pricing.go` con `Calculate(req QuoteRequest, rule *PricingRule, origin, destination Zone, promo *Promotion, now time.Time) (Quote, error)` (rule por puntero: nil = sin regla aplicable)
- [x] 4.2 Validaciones de negocio previas al cálculo: zona origen/destino activa, peso ≤ `MaxWeightKg`, distancia dentro de `[MinDistanceKm, MaxDistanceKm]`, devolviendo el error de dominio correspondiente
- [x] 4.3 Cálculo de componentes base: `base_price`, `distance_fee = round(distance_km * PricePerKm)`, `weight_fee = round(weight_kg * PricePerKg)`
- [x] 4.4 Surcharges: `express_fee` (solo si `shipping_type == express`) y `priority_fee` (solo si `priority == high`) sobre el subtotal, con round-half-up
- [x] 4.5 Aplicación de promoción: validar con `IsValidAt(now)` (error si inválida/expirada cuando se envía), descuento percentage con tope `MaxDiscount`, descuento fixed_amount, nunca mayor que `total_before_discount`
- [x] 4.6 Ensamblar `PricingBreakdown` (verificando que reconcilie: `total_before_discount = base+distance+weight+express+priority`, `total = total_before_discount - discount`) y el `Quote` final
- [x] 4.7 Registrar cada paso relevante en el `PricingDecisionTrace`

## 5. Tests unitarios

- [x] 5.1 `internal/domain/money_test.go`: redondeo half-up (5000.5→5001, 5000.4→5000), suma/resta, no-negatividad donde aplique
- [x] 5.2 `internal/domain/pricing_test.go` — happy paths: standard sin promo (total = base+distance+weight), distance_fee y weight_fee con los ejemplos de la spec, breakdown que suma exacto
- [x] 5.3 `pricing_test.go` — surcharges: express_fee con multiplier 1.5, priority_fee con 1.15, sin surcharge para standard/normal
- [x] 5.4 `pricing_test.go` — promociones: percentage bajo el tope, percentage limitado por tope, fixed_amount, fixed_amount que no supera el total (total=0), sin promoción (discount=0)
- [x] 5.5 `pricing_test.go` — errores de dominio: peso excedido, distancia fuera de rango, zona inactiva, sin regla aplicable, promoción expirada (assert con `errors.Is`)
- [x] 5.6 Correr `go test ./...` y `go vet ./...` en verde (16 tests OK, vet limpio, cobertura 92.1%)

## 6. Cierre del change

- [x] 6.1 `openspec validate add-pricing-domain --strict` sin errores
- [ ] 6.2 Confirmar con el usuario y archivar el change con `/opsx:archive` (consolida la spec en `openspec/specs/pricing-engine/`)
