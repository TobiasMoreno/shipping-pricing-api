## Why

El dominio de pricing ya existe y está testeado, pero no hay forma de correr el servicio. Necesitamos el esqueleto HTTP: un proceso que cargue configuración, exponga un router, responda health/readiness, apague de forma ordenada y se pueda levantar junto a Postgres y Redis con Docker Compose. Esto convierte el proyecto en algo ejecutable y demostrable (`make up` → `curl /health`) y establece el punto de entrada sobre el que se montarán los endpoints de negocio.

## What Changes

- Se agrega `cmd/api/main.go`: punto de entrada que carga config, arma el router, inicia el servidor HTTP y maneja graceful shutdown.
- Se agrega `internal/config`: lectura y validación tipada de variables de entorno (fail-fast ante config requerida ausente).
- Se agrega `internal/server`: router (chi) y servidor HTTP con timeouts y shutdown ordenado por señal del sistema.
- Se agrega `internal/handlers/health_handler.go`: endpoints `GET /health` (liveness, sin dependencias) y `GET /ready` (readiness, agrega checks de dependencias registrables).
- Se introduce un **registry de readiness checks** extensible: en este change no hay dependencias reales registradas (Postgres/Redis/proveedor se suman en changes posteriores), pero queda el mecanismo que las agregará.
- Se agrega infraestructura local: `docker-compose.yml` (API + PostgreSQL + Redis), `deployments/docker/Dockerfile` (multi-stage), `.env.example` y `Makefile` con targets iniciales.
- Primera dependencia externa: `github.com/go-chi/chi/v5`.
- NO incluye: endpoints de negocio (`/shipping/*`), conexión real a Postgres/Redis, cache, métricas, logging estructurado (llegan en changes siguientes).

## Capabilities

### New Capabilities
- `service-config`: carga de configuración desde variables de entorno con tipos explícitos, valores por defecto razonables y validación que falla al arranque si falta config requerida.
- `health-endpoints`: endpoints de liveness (`/health`) y readiness (`/ready`). `/health` responde siempre sin tocar dependencias; `/ready` agrega el estado de las dependencias registradas, devolviendo 503 si alguna crítica falla y 200 (con estado `degraded` por dependencia) en caso contrario.

### Modified Capabilities
<!-- Ninguna. -->

## Impact

- **Código nuevo**: `cmd/api/main.go`, `internal/config/config.go`, `internal/server/router.go`, `internal/server/http_server.go`, `internal/handlers/health_handler.go`.
- **Infra**: `docker-compose.yml`, `deployments/docker/Dockerfile`, `.env.example`, `Makefile`.
- **Dependencias**: agrega `go-chi/chi/v5` a `go.mod`.
- **Tooling local**: el Makefile requiere `make` (pendiente de instalar en el entorno Windows del autor vía `choco install make -y` en terminal elevada). No bloquea la implementación del código, solo la ejecución de los targets.
- **Sin impacto** en el dominio existente (`internal/domain` no se toca). El registry de readiness es el punto de extensión que consumirán `add-postgres-rules`, `add-redis-cache` y `add-logistics-provider`.
