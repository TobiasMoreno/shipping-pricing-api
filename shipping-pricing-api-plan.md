# shipping-pricing-api

## Proyecto backend en Go orientado a portfolio e entrevistas técnicas

> API REST en Go que simula un servicio de cálculo de pricing y disponibilidad de envíos para un e-commerce. El objetivo es demostrar criterio backend real: diseño de servicios, lógica de negocio, cache, resiliencia, testing, observabilidad, Docker, CI/CD y buenas prácticas de ingeniería.

---

## 1. Definición clara del proyecto

### Nombre

`shipping-pricing-api`

### Problema que resuelve

En un marketplace o e-commerce, antes de confirmar una compra, el usuario necesita conocer:

- Qué opciones de envío están disponibles.
- Cuánto cuesta cada opción.
- Cuándo llegaría el paquete.
- Por qué una opción puede no estar disponible.
- Cómo afectan distancia, peso, zona, prioridad, reglas comerciales y promociones al precio final.

Este proyecto simula un **servicio backend de pricing y shipping** similar a una parte de un sistema real de e-commerce.

### Qué demuestra

El proyecto no busca ser un CRUD genérico. Busca demostrar señales fuertes de backend profesional:

- Diseño de una API REST clara.
- Lógica de negocio no trivial.
- Separación por capas.
- Uso realista de Redis como cache.
- Persistencia en PostgreSQL.
- Integración simulada con proveedor externo.
- Timeouts, retries y fallback.
- Logs estructurados y métricas.
- Health checks y readiness checks.
- Tests útiles por capa.
- Docker Compose para entorno local.
- CI con GitHub Actions.
- README profesional orientado a producto e ingeniería.

### Caso de uso principal

Un cliente envía una orden con origen, destino, peso, dimensiones, tipo de envío y datos contextuales. La API devuelve un quote con precio final, días estimados, desglose del cálculo, disponibilidad, trazabilidad de decisiones y si el resultado vino desde cache.

### Pitch corto para README / LinkedIn

`shipping-pricing-api` es una API REST desarrollada en Go que simula un servicio de pricing y disponibilidad de envíos para e-commerce. Incluye reglas de negocio, cache con Redis, PostgreSQL, integración simulada con proveedor logístico, observabilidad, testing y resiliencia ante fallas externas.

---

## 2. Alcance del MVP

El MVP debe ser realista, terminable y presentable. La clave no es cubrir todos los casos de logística del mundo real, sino construir una base sólida con buenas decisiones técnicas.

### Funcionalidades incluidas en el MVP

#### Core funcional

- Calcular precio estimado de envío.
- Consultar opciones de envío disponibles.
- Consultar reglas de pricing activas.
- Crear reglas de pricing.
- Actualizar reglas de pricing.
- Aplicar promociones simples.
- Detectar zonas no disponibles.
- Simular proveedor externo para disponibilidad y ETA.
- Devolver breakdown del cálculo.
- Devolver trazabilidad de decisiones.

#### Infraestructura local

- API en Go.
- PostgreSQL.
- Redis.
- Docker Compose.
- Makefile.
- Variables de entorno.
- Migraciones de base de datos.

#### Calidad técnica

- Tests unitarios de pricing.
- Tests unitarios de validaciones.
- Tests de handlers HTTP.
- Tests de servicios con mocks.
- Tests de integración con PostgreSQL y Redis.
- Logs estructurados.
- Request ID / Correlation ID.
- Métricas básicas.
- Health check.
- Readiness check.
- CI con GitHub Actions.
- OpenAPI / Swagger.

### Fuera del MVP

Esto se puede dejar como mejora futura:

- Autenticación completa con JWT.
- Panel administrativo.
- Sistema avanzado de promociones.
- Motor dinámico complejo de reglas.
- Kafka / eventos asincrónicos.
- Kubernetes.
- Multi-tenant real.
- Trazas distribuidas completas con OpenTelemetry.
- Rate limiting distribuido.
- Geocoding real.
- Integración real con un proveedor logístico.

### Criterio de éxito del MVP

El proyecto está listo para portfolio cuando permite correr:

```bash
make up
make migrate-up
make test
make test-integration
make swagger
```

Y luego probar:

```bash
POST /shipping/quote
GET /shipping/options
GET /shipping/rules
POST /shipping/rules
PUT /shipping/rules/{id}
GET /health
GET /ready
GET /metrics
```

---

## 3. Arquitectura propuesta

### Estilo arquitectónico

Arquitectura por capas, simple y mantenible:

```text
HTTP Handler
    ↓
Service / Use Case
    ↓
Domain Logic
    ↓
Repositories / Cache / External Clients
    ↓
PostgreSQL / Redis / Provider API
```

### Objetivo de la arquitectura

Separar responsabilidades para que:

- La lógica de negocio no dependa del framework HTTP.
- Los handlers solo traduzcan HTTP a casos de uso.
- Los servicios coordinen reglas, cache, persistencia y clientes externos.
- Los repositorios oculten detalles de SQL.
- El cache sea reemplazable o apagable sin romper el core.
- El proveedor externo sea testeable con mocks.
- Los errores sean explícitos y consistentes.

### Capas sugeridas

#### `cmd/api`

Responsabilidad:

- Punto de entrada de la aplicación.
- Cargar configuración.
- Inicializar dependencias.
- Crear router HTTP.
- Configurar middlewares.
- Iniciar servidor.
- Manejar graceful shutdown.

No debería contener lógica de negocio.

---

#### `internal/domain`

Responsabilidad:

- Entidades principales.
- Value objects.
- Errores de dominio.
- Reglas puras de pricing.
- Tipos de shipping.
- Estados de disponibilidad.

Ejemplos:

- `Quote`
- `QuoteRequest`
- `ShippingOption`
- `PricingRule`
- `Zone`
- `Promotion`
- `Money`
- `PricingBreakdown`
- `PricingDecisionTrace`

Esta capa debe poder testearse sin base de datos, Redis ni HTTP.

---

#### `internal/handlers`

Responsabilidad:

- Parsear request HTTP.
- Validar formato básico.
- Llamar a servicios.
- Mapear errores a status codes.
- Devolver JSON consistente.

No debe calcular precios ni acceder directamente a PostgreSQL o Redis.

---

#### `internal/services`

Responsabilidad:

- Orquestar casos de uso.
- Aplicar reglas de negocio.
- Consultar repositorios.
- Consultar cache.
- Consultar proveedor externo.
- Aplicar fallback.
- Registrar decisiones relevantes.

Ejemplos:

- `QuoteService`
- `ShippingOptionsService`
- `PricingRulesService`

---

#### `internal/repositories`

Responsabilidad:

- Acceso a PostgreSQL.
- Querys SQL.
- Persistencia de reglas, zonas, promociones y quotes calculados si se decide guardarlos.
- Mapear registros de DB a entidades de dominio.

Las interfaces deben vivir preferentemente cerca del consumidor. Por ejemplo, si `QuoteService` necesita reglas, puede definir una interfaz `PricingRuleRepository` en la capa de servicio o dominio de aplicación.

---

#### `internal/cache`

Responsabilidad:

- Implementar cache-aside con Redis.
- Serializar/deserializar valores.
- Construir cache keys.
- Manejar TTL.
- Medir hit/miss.
- Degradar de forma segura si Redis falla.

El servicio no debería conocer detalles internos de Redis.

---

#### `internal/clients`

Responsabilidad:

- Integraciones externas.
- Cliente HTTP del proveedor logístico simulado.
- Timeouts.
- Retries controlados.
- Mapeo de errores externos.
- Circuit breaker opcional en mejora futura.

En el MVP, el proveedor puede ser un endpoint interno simulado o un cliente que apunte a un mock server en tests.

---

#### `internal/config`

Responsabilidad:

- Leer variables de entorno.
- Validar configuración requerida.
- Exponer configuración tipada.

Ejemplos:

- `APP_ENV`
- `HTTP_PORT`
- `DATABASE_URL`
- `REDIS_ADDR`
- `REDIS_PASSWORD`
- `CACHE_ENABLED`
- `LOG_LEVEL`
- `PROVIDER_BASE_URL`
- `PROVIDER_TIMEOUT_MS`
- `PROVIDER_MAX_RETRIES`

---

#### `internal/middleware`

Responsabilidad:

- Request ID / Correlation ID.
- Logging HTTP.
- Recovery.
- Rate limiting simple.
- Métricas HTTP.
- Timeout por request si corresponde.

---

#### `internal/observability`

Responsabilidad:

- Logger estructurado.
- Métricas Prometheus.
- Helpers para registrar latencia, errores, cache hit/miss y llamadas externas.

---

### Dependencias recomendadas

| Necesidad | Librería sugerida | Justificación |
|---|---|---|
| Router HTTP | `github.com/go-chi/chi/v5` | Simple, idiomático, compatible con `net/http`, bueno para APIs REST y middlewares. |
| PostgreSQL | `github.com/jackc/pgx/v5` | Driver moderno, performante y muy usado para PostgreSQL en Go. |
| Redis | `github.com/redis/go-redis/v9` | Cliente oficial de Redis para Go, simple y mantenible. |
| Logs | `log/slog` | Logging estructurado en la standard library de Go. Reduce dependencias. |
| Métricas | `github.com/prometheus/client_golang` | Estándar común para exponer métricas en `/metrics`. |
| Migraciones | `github.com/golang-migrate/migrate/v4` | Herramienta simple para versionar cambios de DB. |
| Tests de integración | `github.com/testcontainers/testcontainers-go` | Permite levantar PostgreSQL/Redis reales en tests automatizados. |
| Mocks | `go.uber.org/mock/gomock` o mocks manuales | Para servicios chicos, mocks manuales alcanzan. Para crecer, gomock ayuda. |
| Validación | Manual o `go-playground/validator` | Para portfolio conviene mostrar validaciones explícitas; usar librería solo si simplifica. |
| OpenAPI | `docs/openapi.yaml` + Swagger UI | Mejor empezar con contrato explícito antes que llenar handlers de annotations. |

Decisión recomendada para MVP: usar pocas librerías y priorizar claridad.

---

## 4. Modelo de datos

### Entidades principales

#### `shipping_zones`

Representa zonas logísticas.

| Campo | Tipo | Descripción |
|---|---|---|
| `id` | UUID | Identificador de zona. |
| `code` | VARCHAR | Código único. Ejemplo: `CABA`, `CORDOBA_CAPITAL`, `PATAGONIA`. |
| `name` | VARCHAR | Nombre legible. |
| `region` | VARCHAR | Región general. |
| `is_active` | BOOLEAN | Si la zona está operativa. |
| `created_at` | TIMESTAMP | Fecha de creación. |
| `updated_at` | TIMESTAMP | Fecha de actualización. |

---

#### `pricing_rules`

Reglas configurables para calcular precios.

| Campo | Tipo | Descripción |
|---|---|---|
| `id` | UUID | Identificador de regla. |
| `shipping_type` | VARCHAR | `standard`, `express`, `same_day`. |
| `origin_zone_code` | VARCHAR | Zona origen. Puede ser `*`. |
| `destination_zone_code` | VARCHAR | Zona destino. Puede ser `*`. |
| `base_price_cents` | BIGINT | Precio base en centavos. |
| `price_per_km_cents` | BIGINT | Costo por km. |
| `price_per_kg_cents` | BIGINT | Costo por kg. |
| `express_multiplier` | NUMERIC | Multiplicador para express. |
| `priority_multiplier` | NUMERIC | Multiplicador por prioridad. |
| `max_weight_kg` | NUMERIC | Peso máximo permitido. |
| `min_distance_km` | NUMERIC | Distancia mínima aplicable. |
| `max_distance_km` | NUMERIC | Distancia máxima aplicable. |
| `is_active` | BOOLEAN | Si la regla está activa. |
| `created_at` | TIMESTAMP | Fecha de creación. |
| `updated_at` | TIMESTAMP | Fecha de actualización. |

---

#### `promotions`

Promociones simples aplicables a quotes.

| Campo | Tipo | Descripción |
|---|---|---|
| `id` | UUID | Identificador. |
| `code` | VARCHAR | Código de promo. Ejemplo: `FREE_SHIPPING_10`. |
| `description` | TEXT | Descripción. |
| `discount_type` | VARCHAR | `percentage` o `fixed_amount`. |
| `discount_value` | NUMERIC | Valor del descuento. |
| `max_discount_cents` | BIGINT | Tope de descuento. |
| `starts_at` | TIMESTAMP | Inicio. |
| `ends_at` | TIMESTAMP | Fin. |
| `is_active` | BOOLEAN | Si está activa. |
| `created_at` | TIMESTAMP | Fecha de creación. |
| `updated_at` | TIMESTAMP | Fecha de actualización. |

---

#### `provider_availability_cache` opcional

No es obligatorio si se usa Redis. En MVP, mejor no persistir esta tabla para evitar complejidad innecesaria.

---

#### `shipping_quotes` opcional

Puede usarse para auditoría o trazabilidad.

| Campo | Tipo | Descripción |
|---|---|---|
| `id` | UUID | Quote ID. |
| `request_hash` | VARCHAR | Hash de la request normalizada. |
| `origin_zip_code` | VARCHAR | Código postal origen. |
| `destination_zip_code` | VARCHAR | Código postal destino. |
| `origin_zone_code` | VARCHAR | Zona origen. |
| `destination_zone_code` | VARCHAR | Zona destino. |
| `shipping_type` | VARCHAR | Tipo de envío. |
| `final_price_cents` | BIGINT | Precio final. |
| `currency` | VARCHAR | Moneda. |
| `estimated_delivery_days` | INT | ETA. |
| `was_cached` | BOOLEAN | Si salió de cache. |
| `created_at` | TIMESTAMP | Fecha de creación. |

Recomendación: para el MVP, guardar quotes puede sumar valor porque muestra trazabilidad. No hace falta guardar todo el payload, pero sí datos clave.

---

### Modelo conceptual

```text
shipping_zones 1 ──── * pricing_rules
pricing_rules  * ──── * quotes calculados
promotions     * ──── * quotes calculados
provider       externo/simulado
redis          cache de quotes, reglas y disponibilidad
```

---

## 5. Diseño de endpoints

### Convenciones generales

#### Formato de error

```json
{
  "error": {
    "code": "invalid_request",
    "message": "Invalid request body",
    "details": [
      {
        "field": "weight_kg",
        "reason": "must be greater than 0"
      }
    ],
    "request_id": "req_01HT..."
  }
}
```

#### Status codes comunes

| Status | Uso |
|---|---|
| `200 OK` | Consulta exitosa. |
| `201 Created` | Recurso creado. |
| `400 Bad Request` | Request inválido. |
| `404 Not Found` | Recurso inexistente. |
| `409 Conflict` | Conflicto de regla duplicada o inconsistente. |
| `422 Unprocessable Entity` | Request válido en formato, pero inválido para negocio. |
| `429 Too Many Requests` | Rate limit. |
| `500 Internal Server Error` | Error interno no esperado. |
| `503 Service Unavailable` | Dependencia crítica no disponible. |

---

## `POST /shipping/quote`

Calcula un quote de envío.

### Request body

```json
{
  "origin_zip_code": "5000",
  "destination_zip_code": "1405",
  "origin_zone_code": "CORDOBA_CAPITAL",
  "destination_zone_code": "CABA",
  "distance_km": 695,
  "weight_kg": 2.5,
  "dimensions": {
    "height_cm": 20,
    "width_cm": 30,
    "length_cm": 40
  },
  "shipping_type": "standard",
  "priority": "normal",
  "promotion_code": "SHIP10"
}
```

### Validaciones

- `origin_zip_code` requerido.
- `destination_zip_code` requerido.
- `origin_zone_code` requerido.
- `destination_zone_code` requerido.
- `distance_km > 0`.
- `weight_kg > 0`.
- Dimensiones mayores a cero si se informan.
- `shipping_type` debe ser `standard`, `express` o `same_day`.
- `priority` debe ser `normal` o `high`.
- No permitir origen y destino vacíos o iguales si el caso de negocio no lo soporta.

### Response body

```json
{
  "quote_id": "9af54c7a-9a5e-4ff5-b95d-1d9b38dbf812",
  "price": 12500,
  "currency": "ARS",
  "estimated_delivery_days": 3,
  "shipping_type": "standard",
  "available": true,
  "cached": false,
  "breakdown": {
    "base_price": 5000,
    "distance_fee": 3000,
    "weight_fee": 2500,
    "priority_fee": 0,
    "express_fee": 0,
    "discount": 1000,
    "total_before_discount": 13500,
    "total": 12500
  },
  "decision_trace": [
    "Matched pricing rule for CORDOBA_CAPITAL -> CABA / standard",
    "Applied base price: 5000 ARS",
    "Applied distance fee: 695 km * configured price per km",
    "Applied weight fee: 2.5 kg * configured price per kg",
    "Applied promotion SHIP10 with max discount cap",
    "Provider availability confirmed"
  ]
}
```

### Status codes

| Status | Caso |
|---|---|
| `200 OK` | Quote calculado. |
| `400 Bad Request` | JSON inválido o campos faltantes. |
| `422 Unprocessable Entity` | Zona no disponible, peso excedido, regla no aplicable. |
| `500 Internal Server Error` | Error interno inesperado. |

### Casos borde

- Redis caído: calcular sin cache y loguear degradación.
- Proveedor externo caído: aplicar fallback si hay reglas locales suficientes.
- Zona destino inactiva: responder `422`.
- No existe regla aplicable: responder `422`.
- Promoción inexistente o expirada: no aplicar descuento y agregar decisión al trace, o devolver `422` si se decide que el código inválido debe fallar.

Recomendación: si el usuario envía un código de promoción inválido, devolver `422`. Si no envía promoción, seguir normal.

---

## `GET /shipping/options`

Devuelve opciones disponibles para una ruta y paquete.

### Query params

```text
origin_zone_code=CORDOBA_CAPITAL
destination_zone_code=CABA
weight_kg=2.5
distance_km=695
```

### Response body

```json
{
  "options": [
    {
      "shipping_type": "standard",
      "available": true,
      "estimated_delivery_days": 3,
      "reason": null
    },
    {
      "shipping_type": "express",
      "available": true,
      "estimated_delivery_days": 1,
      "reason": null
    },
    {
      "shipping_type": "same_day",
      "available": false,
      "estimated_delivery_days": null,
      "reason": "same_day is not available for this distance"
    }
  ],
  "cached": true
}
```

### Status codes

| Status | Caso |
|---|---|
| `200 OK` | Opciones calculadas. |
| `400 Bad Request` | Query params inválidos. |
| `422 Unprocessable Entity` | Zona no soportada. |

---

## `GET /shipping/rules`

Lista reglas de pricing.

### Query params opcionales

```text
shipping_type=standard
origin_zone_code=CORDOBA_CAPITAL
destination_zone_code=CABA
active=true
```

### Response body

```json
{
  "rules": [
    {
      "id": "b02d4d15-3ff2-40d1-aeb2-6cbad52440ed",
      "shipping_type": "standard",
      "origin_zone_code": "CORDOBA_CAPITAL",
      "destination_zone_code": "CABA",
      "base_price": 5000,
      "price_per_km": 4,
      "price_per_kg": 1000,
      "express_multiplier": 1.0,
      "priority_multiplier": 1.15,
      "max_weight_kg": 30,
      "min_distance_km": 0,
      "max_distance_km": 1000,
      "is_active": true
    }
  ],
  "cached": true
}
```

### Status codes

| Status | Caso |
|---|---|
| `200 OK` | Reglas encontradas. |
| `400 Bad Request` | Filtros inválidos. |

---

## `POST /shipping/rules`

Crea una regla de pricing.

### Request body

```json
{
  "shipping_type": "standard",
  "origin_zone_code": "CORDOBA_CAPITAL",
  "destination_zone_code": "CABA",
  "base_price": 5000,
  "price_per_km": 4,
  "price_per_kg": 1000,
  "express_multiplier": 1.0,
  "priority_multiplier": 1.15,
  "max_weight_kg": 30,
  "min_distance_km": 0,
  "max_distance_km": 1000,
  "is_active": true
}
```

### Validaciones

- Precios no negativos.
- Multiplicadores mayores o iguales a `1` cuando correspondan.
- `max_distance_km > min_distance_km`.
- `max_weight_kg > 0`.
- No permitir reglas activas duplicadas para la misma combinación exacta.
- Validar que las zonas existan.

### Response body

```json
{
  "id": "b02d4d15-3ff2-40d1-aeb2-6cbad52440ed",
  "message": "pricing rule created"
}
```

### Status codes

| Status | Caso |
|---|---|
| `201 Created` | Regla creada. |
| `400 Bad Request` | Request inválido. |
| `409 Conflict` | Regla duplicada o solapada. |
| `422 Unprocessable Entity` | Datos inconsistentes de negocio. |

### Impacto en cache

Al crear una regla:

- Invalidar cache de reglas activas.
- Invalidar quotes afectados si se puede detectar la combinación.
- En MVP, se puede invalidar por prefijo lógico o versionado de reglas.

---

## `PUT /shipping/rules/{id}`

Actualiza una regla existente.

### Request body

```json
{
  "base_price": 5500,
  "price_per_km": 5,
  "price_per_kg": 1100,
  "is_active": true
}
```

### Response body

```json
{
  "id": "b02d4d15-3ff2-40d1-aeb2-6cbad52440ed",
  "message": "pricing rule updated"
}
```

### Status codes

| Status | Caso |
|---|---|
| `200 OK` | Regla actualizada. |
| `400 Bad Request` | ID inválido o body inválido. |
| `404 Not Found` | Regla inexistente. |
| `409 Conflict` | Actualización genera solapamiento. |
| `422 Unprocessable Entity` | Datos inconsistentes. |

### Impacto en cache

Al actualizar una regla:

- Invalidar cache de reglas.
- Invalidar quotes asociados a esa ruta/tipo de envío.
- Incrementar `pricing_rules_version` si se usa versionado de cache keys.

---

## `GET /health`

Health check simple.

### Response body

```json
{
  "status": "ok"
}
```

Uso: saber si el proceso HTTP está vivo.

No debería depender de PostgreSQL ni Redis.

---

## `GET /ready`

Readiness check.

### Response body exitoso

```json
{
  "status": "ready",
  "dependencies": {
    "postgres": "ok",
    "redis": "degraded",
    "provider": "ok"
  }
}
```

### Criterio recomendado

- PostgreSQL caído: `503`, porque es dependencia crítica.
- Redis caído: puede ser `200` con estado `degraded`, porque el servicio puede calcular sin cache.
- Proveedor externo caído: depende. Para quote puede haber fallback, pero readiness puede reportarlo como `degraded`.

---

## `GET /metrics`

Endpoint Prometheus.

Métricas sugeridas:

- `http_requests_total`
- `http_request_duration_seconds`
- `http_errors_total`
- `shipping_quotes_total`
- `shipping_quote_duration_seconds`
- `shipping_cache_hits_total`
- `shipping_cache_misses_total`
- `shipping_cache_errors_total`
- `shipping_provider_requests_total`
- `shipping_provider_errors_total`
- `shipping_provider_timeout_total`
- `shipping_fallbacks_total`

---

## 6. Estrategia de cache

### Patrón elegido

Usar **cache-aside** con Redis.

Flujo:

```text
Request
  ↓
Buscar en Redis
  ↓ hit
Responder desde cache
  ↓ miss
Calcular / consultar DB / proveedor
  ↓
Guardar en Redis con TTL
  ↓
Responder
```

### Por qué cache-aside

Es simple, explícito y realista para este tipo de servicio. La aplicación controla cuándo lee, cuándo calcula y cuándo escribe en cache.

---

### Qué conviene cachear

#### 1. Quotes de envío

Cachear el resultado de `POST /shipping/quote` cuando la request sea equivalente.

Criterio:

- Mismo origen.
- Mismo destino.
- Misma zona origen.
- Misma zona destino.
- Misma distancia normalizada.
- Mismo peso normalizado.
- Mismas dimensiones si afectan el cálculo.
- Mismo tipo de envío.
- Misma prioridad.
- Misma promoción.
- Misma versión de reglas.

TTL recomendado: `5 a 15 minutos`.

Motivo: el precio puede cambiar por reglas, promociones o disponibilidad externa. No conviene cachearlo por demasiado tiempo.

---

#### 2. Reglas de pricing activas

Cachear reglas que cambian poco.

TTL recomendado: `30 a 60 minutos`.

Motivo: leer reglas activas en cada quote puede ser innecesario. Sin embargo, cuando se crea o actualiza una regla, se debe invalidar.

---

#### 3. Zonas de shipping

Cachear zonas activas.

TTL recomendado: `1 a 6 horas`.

Motivo: las zonas suelen cambiar menos que los quotes.

---

#### 4. Respuestas del proveedor externo

Cachear disponibilidad o ETA del proveedor.

TTL recomendado: `1 a 5 minutos`.

Motivo: disponibilidad logística puede cambiar más rápido. TTL corto reduce presión sobre el proveedor sin entregar datos demasiado viejos.

---

### Qué no conviene cachear

- Requests inválidos.
- Errores internos.
- Errores 5xx del proveedor, salvo que se use negative caching muy controlado.
- Datos sensibles.
- Quotes con reglas experimentales si se agregara feature flag.
- Respuestas con alto riesgo de inconsistencia comercial.

---

### Cache key design

Las keys deben ser:

- Determinísticas.
- Legibles parcialmente.
- Versionables.
- Libres de datos sensibles innecesarios.
- Resistentes a diferencias irrelevantes de formato.

#### Key para quote

```text
shipping:quote:v1:rules:{rules_version}:hash:{request_hash}
```

El `request_hash` sale de una request normalizada.

Ejemplo de normalización:

- `weight_kg` redondeado a 2 decimales.
- `distance_km` redondeado a 1 decimal o entero según decisión de negocio.
- Strings en uppercase.
- Campos ordenados antes de hashear.

#### Key para reglas

```text
shipping:rules:v1:active:{shipping_type}:{origin_zone}:{destination_zone}:rules:{rules_version}
```

#### Key para zonas

```text
shipping:zones:v1:active
```

#### Key para proveedor externo

```text
shipping:provider:v1:availability:{origin_zone}:{destination_zone}:{shipping_type}:{weight_bucket}
```

---

### Invalidación de cache

#### Para reglas

Cuando se crea o actualiza una regla:

- Invalidar cache de reglas activas.
- Incrementar `rules_version`.
- Las keys de quotes anteriores quedan naturalmente obsoletas porque incluyen `rules_version`.

Esta estrategia evita tener que borrar muchas keys por patrón.

#### Para zonas

Cuando se actualiza una zona:

- Invalidar `shipping:zones:v1:active`.
- Incrementar una versión de zonas si el cálculo depende mucho de ellas.

#### Para promociones

Si se cachean promociones:

- TTL corto o invalidación al cambiar promoción.
- Incluir promoción y versión en la key del quote.

---

### Qué pasa si Redis está caído

Redis no debe ser una dependencia crítica para calcular un quote.

Comportamiento esperado:

1. Intentar leer cache.
2. Si Redis falla, loguear `cache_error`.
3. Incrementar métrica `shipping_cache_errors_total`.
4. Continuar con cálculo normal usando DB/proveedor.
5. Intentar guardar en cache al final.
6. Si vuelve a fallar, no romper la respuesta.

La API debe responder correctamente aunque sea más lenta.

---

### Cómo evitar que la caída del cache rompa el servicio

- Timeouts cortos para Redis.
- No propagar errores de cache como errores HTTP.
- Encapsular Redis detrás de una interfaz.
- Devolver `cached: false` si no se pudo usar cache.
- Logs con nivel `warn`, no `error`, si el cálculo pudo completarse.
- Métricas específicas de cache errors.

---

### Cómo medir hit / miss

Métricas:

```text
shipping_cache_hits_total{operation="quote"}
shipping_cache_misses_total{operation="quote"}
shipping_cache_errors_total{operation="quote"}
```

Logs:

```json
{
  "level": "info",
  "message": "cache hit",
  "operation": "shipping_quote",
  "cache_key": "shipping:quote:v1:...",
  "request_id": "req_01HT..."
}
```

---

## 7. Estrategia de testing

El objetivo es que los tests demuestren criterio, no cantidad artificial.

### Pirámide de testing recomendada

```text
Muchos tests unitarios
Algunos tests de servicios con mocks
Algunos tests de handlers HTTP
Pocos tests de integración con PostgreSQL/Redis
```

---

### Tests unitarios de dominio

Ubicación sugerida:

```text
internal/domain/pricing_test.go
```

Qué testear:

- Costo base por tipo de envío.
- Fee por distancia.
- Fee por peso.
- Multiplicador por prioridad.
- Recargo express.
- Descuento por promoción porcentual.
- Descuento fijo.
- Tope máximo de descuento.
- Restricción por peso máximo.
- Restricción por distancia máxima.
- Zona no disponible.
- Redondeo de precios.
- Breakdown correcto.
- Decision trace correcto.

Ejemplos de casos:

| Caso | Resultado esperado |
|---|---|
| Standard sin promo | Precio = base + distancia + peso. |
| Express | Aplica multiplicador o fee express. |
| Promo 10% con tope | Descuento no supera el máximo. |
| Peso excedido | Error de dominio. |
| Zona inactiva | Error de dominio. |
| Sin regla aplicable | Error claro. |

---

### Tests unitarios de validaciones

Ubicación sugerida:

```text
internal/handlers/validators_test.go
```

Qué testear:

- `weight_kg <= 0`.
- `distance_km <= 0`.
- `shipping_type` inválido.
- `priority` inválida.
- CP vacío.
- Zona vacía.
- Dimensiones negativas.
- JSON mal formado.

---

### Tests de servicios con mocks

Ubicación sugerida:

```text
internal/services/quote_service_test.go
```

Mocks:

- `PricingRuleRepository`.
- `ZoneRepository`.
- `PromotionRepository`.
- `QuoteCache`.
- `LogisticsProviderClient`.

Qué testear:

- Cache hit: no consulta DB ni proveedor.
- Cache miss: consulta reglas, proveedor, calcula y guarda en cache.
- Redis caído en lectura: sigue calculando.
- Redis caído en escritura: responde OK igualmente.
- Proveedor caído: aplica fallback si hay reglas locales.
- Proveedor timeout: aplica fallback y registra métrica.
- No hay regla: devuelve error de negocio.
- Promoción expirada: devuelve error o no aplica según decisión.

---

### Tests de handlers HTTP

Ubicación sugerida:

```text
internal/handlers/shipping_handler_test.go
```

Qué testear:

- `POST /shipping/quote` responde `200` con request válido.
- `POST /shipping/quote` responde `400` con JSON inválido.
- `POST /shipping/quote` responde `422` con zona no disponible.
- El response incluye `request_id` en errores.
- `GET /shipping/options` valida query params.
- `GET /health` responde sin dependencias.
- `GET /ready` refleja dependencias.

Herramientas:

- `net/http/httptest`.
- Mocks simples de servicios.

---

### Tests de integración

Ubicación sugerida:

```text
tests/integration/postgres_test.go
tests/integration/redis_test.go
tests/integration/quote_flow_test.go
```

Qué testear:

- Migraciones corren correctamente.
- Repositorio crea y lista reglas.
- Repositorio actualiza reglas.
- Redis guarda y recupera quote cacheado.
- TTL funciona.
- Flujo completo quote con PostgreSQL y Redis reales.

Herramienta recomendada:

- Testcontainers para levantar PostgreSQL y Redis reales durante tests.

---

### Tests de resiliencia

Casos mínimos:

| Falla | Resultado esperado |
|---|---|
| Redis caído | Quote se calcula igual y `cached=false`. |
| Proveedor timeout | Fallback local si existe regla. |
| Proveedor devuelve 500 | Retry controlado y fallback. |
| PostgreSQL caído | `ready` devuelve `503`; quote falla si necesita reglas. |
| Request inválido | Error seguro sin stack trace. |
| Zona no disponible | `422` con error de negocio. |

---

## 8. Estrategia de observabilidad

### Objetivo

Poder responder preguntas típicas de producción:

- ¿Qué endpoint está fallando?
- ¿Qué requests son lentas?
- ¿Redis está ayudando o fallando?
- ¿El proveedor externo está generando timeouts?
- ¿Cuántos quotes se calculan por minuto?
- ¿Cuántas veces usamos fallback?
- ¿Qué decisión tomó el motor de pricing?

---

### Logs estructurados

Usar `slog` con formato JSON.

Campos mínimos:

- `timestamp`
- `level`
- `message`
- `request_id`
- `method`
- `path`
- `status`
- `latency_ms`
- `shipping_type`
- `origin_zone`
- `destination_zone`
- `cache_status`
- `provider_status`
- `error_code`

Ejemplo:

```json
{
  "time": "2026-06-27T19:00:00Z",
  "level": "INFO",
  "msg": "shipping quote calculated",
  "request_id": "req_01HT...",
  "shipping_type": "standard",
  "origin_zone": "CORDOBA_CAPITAL",
  "destination_zone": "CABA",
  "price": 12500,
  "cached": false,
  "provider_status": "ok",
  "latency_ms": 82
}
```

---

### Request ID / Correlation ID

Middleware:

- Si llega header `X-Request-ID`, reutilizarlo.
- Si no llega, generar uno.
- Devolverlo en response header.
- Inyectarlo en `context.Context`.
- Usarlo en logs, errores y llamadas externas.

Header recomendado:

```text
X-Request-ID: req_01HT...
```

---

### Métricas

#### HTTP

- Cantidad de requests por endpoint/status.
- Latencia por endpoint.
- Errores por endpoint.

#### Cache

- Hits.
- Misses.
- Errores.
- Latencia de Redis si se quiere sumar.

#### Proveedor externo

- Requests.
- Timeouts.
- Errores 5xx.
- Retries.
- Fallbacks.

#### Negocio

- Quotes calculados.
- Quotes cacheados.
- Quotes rechazados por zona.
- Quotes rechazados por regla no aplicable.
- Distribución por shipping type.

---

### Health vs Readiness

#### `/health`

Indica si el proceso está vivo.

No consulta dependencias.

#### `/ready`

Indica si la aplicación está lista para recibir tráfico.

Consulta:

- PostgreSQL.
- Redis.
- Proveedor externo si aplica.

Criterio:

- PostgreSQL es crítico.
- Redis puede ser degradado.
- Proveedor puede ser degradado si hay fallback.

---

### Cómo ayuda a debuggear producción

Ejemplo de problema: usuarios reportan que el envío express está caro.

Con esta observabilidad se puede revisar:

1. Logs por `shipping_type=express`.
2. Breakdown del quote.
3. Regla aplicada.
4. Si hubo promoción o no.
5. Si el resultado vino de cache.
6. Versión de reglas usada en la cache key.
7. Métricas de cambios de latencia o errores.

Ejemplo de problema: la API está lenta.

Se puede revisar:

1. Histograma de latencia por endpoint.
2. Cache hit ratio.
3. Timeouts del proveedor.
4. Errores de Redis.
5. Logs con `latency_ms` alto y `request_id`.

---

## 9. Estructura de carpetas

```text
shipping-pricing-api/
├── cmd/
│   └── api/
│       └── main.go
│
├── internal/
│   ├── config/
│   │   └── config.go
│   │
│   ├── domain/
│   │   ├── quote.go
│   │   ├── pricing_rule.go
│   │   ├── shipping_option.go
│   │   ├── promotion.go
│   │   ├── zone.go
│   │   ├── money.go
│   │   └── errors.go
│   │
│   ├── handlers/
│   │   ├── shipping_handler.go
│   │   ├── rules_handler.go
│   │   ├── health_handler.go
│   │   ├── metrics_handler.go
│   │   ├── request_models.go
│   │   ├── response_models.go
│   │   └── validators.go
│   │
│   ├── services/
│   │   ├── quote_service.go
│   │   ├── shipping_options_service.go
│   │   ├── pricing_rules_service.go
│   │   └── ports.go
│   │
│   ├── repositories/
│   │   ├── postgres/
│   │   │   ├── db.go
│   │   │   ├── pricing_rule_repository.go
│   │   │   ├── zone_repository.go
│   │   │   ├── promotion_repository.go
│   │   │   └── quote_repository.go
│   │   └── errors.go
│   │
│   ├── cache/
│   │   ├── redis_client.go
│   │   ├── quote_cache.go
│   │   ├── rules_cache.go
│   │   ├── provider_cache.go
│   │   ├── keys.go
│   │   └── noop_cache.go
│   │
│   ├── clients/
│   │   └── logistics/
│   │       ├── client.go
│   │       ├── models.go
│   │       └── errors.go
│   │
│   ├── middleware/
│   │   ├── request_id.go
│   │   ├── logging.go
│   │   ├── recovery.go
│   │   ├── rate_limit.go
│   │   └── metrics.go
│   │
│   ├── observability/
│   │   ├── logger.go
│   │   └── metrics.go
│   │
│   └── server/
│       ├── router.go
│       └── http_server.go
│
├── migrations/
│   ├── 000001_create_shipping_zones.up.sql
│   ├── 000001_create_shipping_zones.down.sql
│   ├── 000002_create_pricing_rules.up.sql
│   ├── 000002_create_pricing_rules.down.sql
│   ├── 000003_create_promotions.up.sql
│   ├── 000003_create_promotions.down.sql
│   ├── 000004_create_shipping_quotes.up.sql
│   └── 000004_create_shipping_quotes.down.sql
│
├── docs/
│   ├── openapi.yaml
│   ├── architecture.md
│   └── decisions/
│       ├── 0001-use-cache-aside.md
│       ├── 0002-use-postgresql.md
│       └── 0003-provider-fallback.md
│
├── tests/
│   └── integration/
│       ├── postgres_test.go
│       ├── redis_test.go
│       └── quote_flow_test.go
│
├── deployments/
│   └── docker/
│       └── Dockerfile
│
├── .github/
│   └── workflows/
│       └── ci.yml
│
├── .env.example
├── .gitignore
├── docker-compose.yml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

---

## 10. Roadmap por etapas

### Etapa 1 — Base del proyecto

Objetivo: tener una API mínima corriendo.

Entregables:

- Crear módulo Go.
- Estructura de carpetas.
- Config por env vars.
- Router.
- `/health`.
- `/ready` básico.
- Dockerfile.
- Docker Compose con API, PostgreSQL y Redis.
- Makefile inicial.

Resultado visible:

```bash
make up
curl localhost:8080/health
```

---

### Etapa 2 — Dominio y pricing engine

Objetivo: implementar la lógica pura.

Entregables:

- Entidades de dominio.
- Reglas de pricing.
- Cálculo de breakdown.
- Decision trace.
- Errores de dominio.
- Tests unitarios de pricing.

Resultado visible:

- Tests pasando sin DB ni Redis.
- Pricing entendible y defendible en entrevista.

---

### Etapa 3 — PostgreSQL y reglas

Objetivo: persistir reglas reales.

Entregables:

- Migraciones.
- Seed data.
- Repositorios PostgreSQL.
- `GET /shipping/rules`.
- `POST /shipping/rules`.
- `PUT /shipping/rules/{id}`.
- Tests de repositorio.

Resultado visible:

- Reglas editables vía API.
- Base de datos real en Docker.

---

### Etapa 4 — Quote endpoint

Objetivo: endpoint principal funcionando.

Entregables:

- `POST /shipping/quote`.
- Validaciones.
- Servicio de quote.
- Uso de reglas desde DB.
- Persistencia opcional de quote.
- Tests de handler y servicio.

Resultado visible:

- Endpoint devuelve precio, ETA, breakdown y decision trace.

---

### Etapa 5 — Redis cache

Objetivo: agregar cache-aside realista.

Entregables:

- Cliente Redis.
- Quote cache.
- Rules cache.
- Cache key design.
- TTLs.
- Degradación si Redis falla.
- Métricas hit/miss.
- Tests de cache.

Resultado visible:

- Primera request calcula.
- Segunda request equivalente responde desde cache.
- Si Redis se apaga, el quote sigue funcionando.

---

### Etapa 6 — Proveedor externo simulado

Objetivo: mostrar integración y resiliencia.

Entregables:

- Cliente HTTP.
- Timeout.
- Retry controlado.
- Fallback.
- Cache de respuesta del proveedor.
- Logs y métricas del proveedor.
- Tests de proveedor caído y timeout.

Resultado visible:

- El sistema no se rompe ante fallas externas.

---

### Etapa 7 — Observabilidad

Objetivo: hacerlo presentable como servicio real.

Entregables:

- Logs JSON con `slog`.
- Request ID.
- Latencia por request.
- `/metrics`.
- Métricas de negocio/cache/proveedor.
- README explicando debugging.

Resultado visible:

- Se puede inspeccionar comportamiento del sistema.

---

### Etapa 8 — OpenAPI, CI y README

Objetivo: cerrar como proyecto de portfolio.

Entregables:

- `docs/openapi.yaml`.
- Swagger UI o instrucciones para visualizar OpenAPI.
- GitHub Actions.
- README profesional.
- Diagrama simple.
- Ejemplos curl.
- ADRs de decisiones técnicas.

Resultado visible:

- Proyecto listo para GitHub, LinkedIn, CV y entrevista.

---

## 11. Qué debería implementar primero

Orden recomendado:

### 1. Dominio antes que infraestructura

Primero implementar la lógica de pricing como funciones/servicios puros.

Motivo:

- Es el corazón del proyecto.
- Es fácil de testear.
- Evita arrancar por Docker/DB y perder foco.
- Demuestra pensamiento de negocio.

Primera meta concreta:

```text
Dado un QuoteRequest y una PricingRule aplicable,
devolver un Quote con price, breakdown y decision_trace.
```

---

### 2. API mínima con `/health`

Después crear el esqueleto HTTP.

Motivo:

- Permite validar estructura.
- Permite correr el servicio.
- Da sensación de avance rápido.

---

### 3. `POST /shipping/quote` sin DB ni Redis

Primera versión con reglas en memoria.

Motivo:

- Permite probar el caso de uso principal rápido.
- Reduce complejidad inicial.
- Después reemplazás reglas en memoria por repositorio PostgreSQL.

---

### 4. PostgreSQL

Agregar persistencia de reglas.

Motivo:

- Convierte el proyecto en backend real.
- Permite endpoints administrativos de reglas.

---

### 5. Redis

Agregar cache cuando el cálculo ya funciona.

Motivo:

- Evita cachear una lógica que todavía cambia.
- Permite comparar con/sin cache.

---

### 6. Proveedor externo simulado

Agregar integración y resiliencia.

Motivo:

- Muestra manejo de sistemas distribuidos.
- Suma mucho para entrevistas.

---

### 7. Observabilidad y CI

Cerrar con calidad profesional.

Motivo:

- Hace que el proyecto parezca preparado para producción.
- Mejora mucho la presentación en GitHub.

---

## 12. Decisiones técnicas que podés explicar en una entrevista

### 1. Por qué Go

Go es una buena elección para servicios backend por su simplicidad, performance, concurrencia nativa, binarios simples de desplegar y ecosistema sólido para APIs HTTP.

Cómo explicarlo:

> Elegí Go porque quería construir un servicio backend simple, performante y fácil de desplegar. Para una API de pricing/shipping, Go permite manejar HTTP, timeouts, context propagation y concurrencia de forma clara sin agregar demasiada complejidad.

---

### 2. Por qué arquitectura por capas

Cómo explicarlo:

> Separé handlers, servicios, dominio, repositorios, cache y clientes externos para que cada capa tenga una responsabilidad clara. Esto me permite testear la lógica de pricing sin HTTP ni base de datos, mockear dependencias externas y mantener el código más fácil de cambiar.

---

### 3. Por qué interfaces

Cómo explicarlo:

> Uso interfaces en los límites del servicio: repositorios, cache y proveedor externo. No las uso para todo, solo donde necesito desacoplar dependencias, facilitar tests o permitir reemplazar implementaciones.

---

### 4. Por qué cache-aside

Cómo explicarlo:

> Elegí cache-aside porque es simple y explícito. Primero intento leer Redis; si hay miss, calculo el quote con reglas y proveedor, y luego guardo el resultado con TTL. Si Redis falla, el servicio sigue funcionando porque el cache no es una dependencia crítica.

---

### 5. Qué cachear y qué no

Cómo explicarlo:

> Cacheo quotes repetidos, reglas activas, zonas y disponibilidad del proveedor con TTLs distintos. No cacheo errores internos ni requests inválidos. Los quotes tienen TTL corto porque pueden cambiar por promociones, reglas o disponibilidad.

---

### 6. Cómo resolvés invalidación de cache

Cómo explicarlo:

> Para evitar borrar muchas keys, uso versionado en las cache keys. Cuando cambia una regla, incremento la versión de reglas. Entonces los quotes anteriores quedan obsoletos automáticamente porque su key incluye una versión vieja.

---

### 7. Qué pasa si Redis se cae

Cómo explicarlo:

> Redis está tratado como dependencia no crítica. Si falla una lectura o escritura de cache, registro el error, incremento una métrica y continúo con el cálculo usando DB y proveedor. El impacto es más latencia, no caída funcional del endpoint.

---

### 8. Cómo manejás fallas del proveedor externo

Cómo explicarlo:

> El cliente del proveedor usa timeout por request, retries acotados para errores transitorios y fallback local cuando tiene sentido. Por ejemplo, si el proveedor falla pero tengo reglas locales suficientes, devuelvo un quote marcado como calculado con fallback y lo registro en logs/métricas.

---

### 9. Por qué context propagation

Cómo explicarlo:

> Propago `context.Context` desde el handler hasta repositorios, Redis y cliente externo para respetar cancelaciones, deadlines y request ID. Esto evita trabajo innecesario si el cliente corta la conexión o si se supera el timeout.

---

### 10. Por qué logs estructurados

Cómo explicarlo:

> Uso logs estructurados para poder filtrar por request ID, endpoint, shipping type, zona, estado de cache o proveedor. Esto hace mucho más fácil debuggear problemas de producción que logs de texto libres.

---

### 11. Por qué métricas

Cómo explicarlo:

> Las métricas permiten ver comportamiento agregado: latencia, errores, cache hit ratio, timeouts del proveedor y uso de fallback. Los logs ayudan a investigar un caso puntual; las métricas ayudan a detectar tendencias.

---

### 12. Por qué tests por capa

Cómo explicarlo:

> Testeo la lógica de pricing con unit tests porque ahí está el negocio. Testeo servicios con mocks para cubrir cache, proveedor y repositorios. Testeo handlers para validar contratos HTTP. Y uso integración con PostgreSQL/Redis para asegurar que las dependencias reales funcionan.

---

### 13. Cómo evitás sobreingeniería

Cómo explicarlo:

> No agregué Kafka, Kubernetes ni microservicios innecesarios. El proyecto está pensado como un servicio modular pero monolítico, con buenas prácticas de backend real. La complejidad está donde aporta: pricing, cache, resiliencia, testing y observabilidad.

---

### 14. Cómo se relaciona con Mercado Libre

Cómo explicarlo:

> El proyecto está inspirado en problemas típicos de marketplace: cálculo de costos, disponibilidad logística, reglas comerciales, performance, cache, integración con proveedores y resiliencia. No replica sistemas internos de Mercado Libre, pero demuestra habilidades transferibles para backend en e-commerce a escala.

---

## README profesional sugerido

El README final del repositorio debería incluir estas secciones:

```text
# shipping-pricing-api

## Overview
## Problem
## Why I built this
## Features
## Tech Stack
## Architecture
## Component Diagram
## Business Rules
## API Endpoints
## Example Requests
## Example Responses
## Cache Strategy
## Observability
## Resilience
## Testing Strategy
## Local Setup
## Environment Variables
## Makefile Commands
## CI Pipeline
## Technical Decisions
## Future Improvements
## How this project maps to backend engineering skills
```

---

## Diagrama simple de componentes

```text
                 ┌──────────────────────┐
                 │      API Client       │
                 └──────────┬───────────┘
                            │
                            ▼
                 ┌──────────────────────┐
                 │   Go REST API         │
                 │ chi + net/http        │
                 └──────────┬───────────┘
                            │
          ┌─────────────────┼─────────────────┐
          ▼                 ▼                 ▼
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│ Quote Service    │ │ Rules Service    │ │ Options Service  │
└────────┬─────────┘ └────────┬─────────┘ └────────┬─────────┘
         │                    │                    │
         ▼                    ▼                    ▼
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│ Pricing Engine   │ │ PostgreSQL       │ │ Logistics Client │
│ Domain Logic     │ │ Rules/Zones      │ │ External Mock    │
└────────┬─────────┘ └──────────────────┘ └──────────────────┘
         │
         ▼
┌──────────────────┐
│ Redis Cache      │
│ quote/rules/ETA  │
└──────────────────┘
```

---

## Variables de entorno sugeridas

```text
APP_ENV=local
HTTP_PORT=8080
LOG_LEVEL=debug
DATABASE_URL=postgres://postgres:postgres@postgres:5432/shipping?sslmode=disable
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_DB=0
CACHE_ENABLED=true
QUOTE_CACHE_TTL_SECONDS=600
RULES_CACHE_TTL_SECONDS=3600
PROVIDER_BASE_URL=http://localhost:9090
PROVIDER_TIMEOUT_MS=800
PROVIDER_MAX_RETRIES=2
RATE_LIMIT_REQUESTS_PER_MINUTE=120
```

---

## Makefile sugerido

```text
make run                 # correr API local
make up                  # levantar docker compose
make down                # apagar docker compose
make test                # tests unitarios
make test-integration    # tests con PostgreSQL/Redis
make lint                # lint
make migrate-up          # correr migraciones
make migrate-down        # rollback
make seed                # cargar datos iniciales
make swagger             # validar/generar documentación OpenAPI
```

---

## CI con GitHub Actions

Pipeline mínimo:

```text
on: pull_request / push

jobs:
  test:
    - checkout
    - setup-go
    - go mod download
    - go test ./...
    - go vet ./...
    - docker compose config
```

Pipeline mejorado:

```text
jobs:
  unit-tests
  integration-tests
  lint
  openapi-validation
  docker-build
```

---

## Próximas mejoras después del MVP

- Autenticación con API key para endpoints administrativos.
- Rate limiting por IP o API key.
- Circuit breaker para proveedor externo.
- OpenTelemetry tracing.
- Dashboard Grafana.
- Feature flags para reglas nuevas.
- Auditoría completa de cambios de reglas.
- Endpoint para comparar todas las opciones con precio.
- Soporte para volumen dimensional.
- Soporte para múltiples monedas.
- Simulación de picos de tráfico.
- Load testing con k6.
- Deploy en Fly.io, Render, Railway o AWS.

---

## Señales fuertes para recruiter/interviewer

Este proyecto comunica:

| Señal | Dónde se ve |
|---|---|
| Backend real | Pricing, reglas, proveedor, errores, DB. |
| Go | API, context, interfaces, tests, graceful shutdown. |
| APIs REST | Endpoints claros, status codes, OpenAPI. |
| Diseño de servicios | Capas, puertos, repositorios, clientes. |
| Cache | Redis, TTL, keys, invalidación, hit/miss. |
| Testing | Unit, handlers, services, integración. |
| Observabilidad | Logs, request ID, métricas, health/ready. |
| Resiliencia | Timeouts, retries, fallback, degradación. |
| Docker | Compose con API, DB, Redis. |
| CI/CD | GitHub Actions. |
| Producto | Quote con breakdown y decision trace. |
| Escalabilidad | Cache, métricas, separación de responsabilidades. |

---

## Implementación recomendada inmediata

Primera tarea concreta:

```text
Crear el dominio de pricing y sus tests unitarios.
```

Primeros archivos a implementar cuando empiece el código:

```text
internal/domain/quote.go
internal/domain/pricing_rule.go
internal/domain/money.go
internal/domain/errors.go
internal/domain/pricing_test.go
```

Primer test que debería existir:

```text
Given a valid standard shipping quote request
And an active pricing rule
When the quote is calculated
Then the response includes final price, breakdown and decision trace
```

Con esto el proyecto arranca desde la lógica de negocio, que es lo que más diferencia a este backend de un CRUD genérico.

---

## Referencias técnicas oficiales consultadas

- Go `log/slog`: https://pkg.go.dev/log/slog
- Go blog sobre structured logging con `slog`: https://go.dev/blog/slog
- chi router: https://github.com/go-chi/chi
- pgx PostgreSQL driver: https://github.com/jackc/pgx
- go-redis: https://github.com/redis/go-redis
- Redis Go guide: https://redis.io/docs/latest/develop/clients/go/
- Prometheus client_golang: https://github.com/prometheus/client_golang
- Prometheus Go instrumentation guide: https://prometheus.io/docs/guides/go-application/
- golang-migrate: https://github.com/golang-migrate/migrate
- Testcontainers for Go: https://golang.testcontainers.org/
