## 1. Dependencias

- [x] 1.1 `go get github.com/go-chi/chi/v5` y `go mod tidy`

## 2. Configuración

- [x] 2.1 Implementar `internal/config/config.go`: struct `Config` tipada con campos para APP_ENV, HTTP_PORT, LOG_LEVEL, DATABASE_URL, REDIS_ADDR, REDIS_PASSWORD, REDIS_DB, CACHE_ENABLED, QUOTE_CACHE_TTL_SECONDS, RULES_CACHE_TTL_SECONDS, PROVIDER_BASE_URL, PROVIDER_TIMEOUT_MS, PROVIDER_MAX_RETRIES, RATE_LIMIT_REQUESTS_PER_MINUTE
- [x] 2.2 Helpers `getString`/`getInt`/`getBool` con default y `Load() (Config, error)` que parsea y valida
- [x] 2.3 Validación fail-fast: requeridos ausentes/inválidos devuelven error (DATABASE_URL, HTTP_PORT numérico)
- [x] 2.4 `internal/config/config_test.go`: defaults aplicados, valor leído de env, requerido faltante → error, HTTP_PORT no numérico → error

## 3. Health registry y handlers

- [x] 3.1 Implementar `internal/handlers/health_handler.go`: `HealthRegistry` con `Register(name string, critical bool, check Check)` y `Check func(ctx) error`
- [x] 3.2 Handler `GET /health` → 200 `{"status":"ok"}` sin tocar dependencias
- [x] 3.3 Handler `GET /ready` → ejecuta checks con timeout por check, arma `dependencies`, 503 si falla algún crítico, 200 (con `degraded` en no críticos) en caso contrario; sin checks registrados → 200
- [x] 3.4 `internal/handlers/health_handler_test.go` con `httptest`: /health ok, /ready sin checks → 200, crítico falla → 503, no crítico falla → 200 degraded

## 4. Servidor HTTP

- [x] 4.1 Implementar `internal/server/router.go`: construye `chi.Router`, monta `/health` y `/ready`, deja lugar para middlewares y rutas futuras
- [x] 4.2 Implementar `internal/server/http_server.go`: `http.Server` con ReadHeaderTimeout/ReadTimeout/WriteTimeout/IdleTimeout y método de arranque + `Shutdown(ctx)`

## 5. Entry point y graceful shutdown

- [x] 5.1 Implementar `cmd/api/main.go`: cargar config (fail-fast), crear logger `slog` básico, construir router y server, arrancar
- [x] 5.2 Graceful shutdown con `signal.NotifyContext` (SIGINT/SIGTERM) y `Shutdown` con timeout

## 6. Infraestructura local

- [x] 6.1 `.env.example` con todas las variables del plan
- [x] 6.2 `deployments/docker/Dockerfile` multi-stage (build con golang:1.25, runtime distroless/static)
- [x] 6.3 `docker-compose.yml`: servicios api, postgres:16, redis:7, con healthchecks, env y `depends_on`
- [x] 6.4 `Makefile` con targets: run, up, down, test, migrate-up, migrate-down, lint (los que aplican a esta etapa funcionan; los de migración quedan como placeholders hasta `add-postgres-rules`)
- [x] 6.5 `.dockerignore`

## 7. Verificación

- [x] 7.1 `go vet ./...` y `go test ./...` en verde
- [x] 7.2 `go run ./cmd/api` levanta y `curl localhost:8080/health` responde `{"status":"ok"}` (o equivalente con PowerShell `Invoke-RestMethod`)
- [x] 7.3 `docker compose config` valida el compose
- [x] 7.4 `openspec validate add-http-skeleton --strict` sin errores
- [x] 7.5 Confirmar con el usuario y archivar el change
