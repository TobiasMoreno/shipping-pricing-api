## Context

El proyecto tiene el dominio de pricing listo y testeado pero no es ejecutable. Este change agrega la capa de runtime HTTP mínima y la infraestructura local, sin todavía conectar Postgres/Redis ni exponer endpoints de negocio. Es el primer código que importa una dependencia externa (chi). El registry de readiness debe diseñarse pensando en que changes futuros (Postgres, Redis, proveedor) le agregarán checks sin reescribirlo.

## Goals / Non-Goals

**Goals:**
- Proceso ejecutable: `go run ./cmd/api` levanta un servidor HTTP en el puerto configurado.
- Config tipada desde env con defaults y validación fail-fast.
- `/health` y `/ready` con un mecanismo de checks de dependencias extensible.
- Graceful shutdown ante SIGINT/SIGTERM con timeout.
- `docker compose up` levanta API + Postgres + Redis.

**Non-Goals:**
- Conexión real a Postgres o Redis (este change no abre pools ni clientes; solo deja el registry listo).
- Endpoints `/shipping/*`, métricas, logging estructurado avanzado (slog se introduce en `add-observability`; acá un logger básico alcanza).
- Autenticación, rate limiting.

## Decisions

**1. Router: go-chi/chi/v5.**
Idiomático, compatible con `net/http`, buen soporte de middlewares y subrouters. Alternativas descartadas: `net/http` puro con `http.ServeMux` (menos ergonómico para middlewares/grupos), o frameworks pesados como gin/echo (más dependencias de las necesarias para el objetivo).

**2. Config con stdlib (os.Getenv) + helpers tipados.**
Un paquete `config` con `Load() (Config, error)` que lee env vars, parsea con helpers (`getString`, `getInt`, `getBool` con default) y valida requeridos. Sin dependencia de viper/envconfig: para el MVP, código explícito es más legible y demuestra criterio. `.env` se carga vía Docker Compose / shell, no con una lib de dotenv en runtime.

**3. Readiness via registry de checks.**
`type Check func(ctx context.Context) error` y un `HealthRegistry` que guarda checks con nombre y un flag `critical`. `/ready` ejecuta todos con un timeout corto por check, arma `dependencies: {name: ok|degraded}` y decide el status: 503 si falla algún `critical`, 200 si no. Alternativa descartada: chequear dependencias concretas hardcodeadas en el handler (no extensible, acopla el handler a cada dependencia).

**4. Graceful shutdown.**
`main` escucha `signal.NotifyContext` (SIGINT/SIGTERM); al cancelarse, llama `server.Shutdown(ctx)` con un timeout (p. ej. 10s) para drenar requests en curso. El `http.Server` se configura con `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout` para no quedar expuesto a conexiones lentas.

**5. Dockerfile multi-stage.**
Stage de build con `golang:1.25` (compila binario estático), stage final `gcr.io/distroless/static` o `alpine` con el binario. Imagen chica y sin toolchain. Compose monta API + `postgres:16` + `redis:7` con healthchecks y `depends_on`.

**6. Logger básico por ahora.**
Un `*slog.Logger` con handler de texto/JSON simple inicializado en `main` y pasado a los componentes. El cableado completo de observabilidad (request-id, métricas) es del change `add-observability`; acá solo evitamos `fmt.Println`.

## Risks / Trade-offs

- **[`/ready` sin dependencias reales en este change]** → Podría parecer trivial. Mitigación: el valor está en el registry extensible y sus tests (incluyendo el caso "sin checks registrados → 200"); los checks reales se enchufan en changes posteriores sin tocar el handler.
- **[Make no instalado en el entorno del autor]** → Los targets del Makefile no corren localmente hasta instalar make. Mitigación: `choco install make -y` (acción del usuario, en terminal elevada); mientras tanto los comandos equivalentes (`go run`, `docker compose up`) funcionan directo. El Makefile igual aporta valor para CI (runners Linux) y para el README.
- **[Timeouts del server]** → Valores fijos iniciales; se podrán parametrizar por env si hiciera falta. No es crítico para el MVP.

## Open Questions

- ¿Imagen base final del Dockerfile: `distroless/static` (más segura, sin shell) o `alpine` (debuggeable con shell)? Se decide al escribir el Dockerfile; default propuesto: `distroless/static`.
