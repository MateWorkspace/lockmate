# Composition

Wire concrete drivers and adapters into domain contracts, application usecases,
and presentation handlers. Composition is the only layer that imports concrete
implementations from infrastructure, application, and presentation together.
Keep business rules and request handling in their existing layers.

## Main composition

`main/launcher.go` exposes Launch() int, invoked by cmd/main/main.go. Load global
configuration and embedded seed definitions first, then initialize driver,
infrastructure, and application. Seed before presentation and cron startup. Use private structs holding each layer's dependencies.

- driver.go: slog JSON/text output, PostgreSQL pool, Gin engine + http.Server, and
  the gRPC listener + server.
- infrastructure.go: interface-typed repositories and utility adapters.
- application.go: interface-typed usecases and explicit constructor wiring.
- seeder.go: clone embedded user/client definitions and inject runtime credentials
  from config.Env before constructor wiring; never mutate global SeedData.
- presentation.go: HTTP handlers, middleware usecase dependencies, route setup,
  gRPC interceptor + service registration, and the cron scheduler lifecycle.
- launcher.go: signal handling and lifecycle; RunServers serves HTTP and gRPC
  concurrently and drains both; RunServer provides a testable HTTP boundary.

Follow concatenated package names, explicit import aliases, existing Id spelling,
and readable blank lines. No reflection, service locator, or dependency container.
Reuse one repository Pgxdt executor and the strict application Transactor; do not
replace its rejection of nested application transactions.

## Runtime

Use config.Env and config.App; never parse environment variables elsewhere.
Create only slog, PostgreSQL, HTTP, and gRPC dependencies currently implemented.
HTTP binds PortHttpServer; gRPC binds PortGrpcServer, both with typed shutdown
timeouts. Disable implicit trusted proxies and Gin's default request logger. The
gRPC server uses the shared interceptor chain (client auth, IP allowlist, bearer).

Startup failures return 1; normal signal shutdown returns 0. HTTP and gRPC drain
before PostgreSQL closes. Force-close servers when the shutdown timeout expires.
Use explicit infrastructure logger calls after infrastructure initialization;
bootstrap configuration and driver failures use slog directly. Do not dump
configuration or connection strings. Docker entrypoint applies migrations before
Launch; native Go startup requires migrations externally. Run the application
seeder when config.Env.SeederEnabled is true, and fail startup on seeding errors.

After seeding, `presentation.go` registers the HTTP routes and starts the cron
presentation through `presentation/cron/runner.Run`, which runs each job once
synchronously (failing launch on a required error) and then schedules it with the
cron engine. The signing-key job reconciles the active key, re-seals keys under
the current key-encryption key, and retires old keys on its own reconcile
interval (which must not exceed the rotation interval); the auth-flow job is
optional best-effort cleanup. Store the returned stop function on the
presentation and call it before PostgreSQL closes. Never parse environment
variables outside config.

Embedded seed definitions provision the control-space baseline. Its slug comes
from seederdata.SeedData.Space.Slug and is passed to management constructors even
when seeding is disabled. Google OAuth and handoff wiring are included; HTTP and
gRPC share the same usecases. Redis remains outside this composition. Future
additions must follow the same layer boundaries.

Run make swagger after annotation changes. Verify go test -race ./..., go vet
./..., and go build ./.... Tests for lifecycle and startup failures live under
tests/composition; PostgreSQL tests use isolated schemas and
LOCKMATE_TEST_DATABASE_URL.

Google is enabled explicitly through config.Env. HTTP defaults to 10701 behind
cloudflared; the public HTTPS origin config determines launch/callback URLs.
Wire AuthFlow random generation, PostgreSQL launch/attempt/handoff repositories,
Google OAuth, Google usecases, generic Handoff, signing-key reconciliation, and
misc expiration cleanup. Cron jobs run at startup and on their cron schedule with
bounded work/deadlines; stop and join them before closing PostgreSQL. Never log
OAuth settings or secrets.
