# Composition

Wire concrete drivers and adapters into domain contracts, application usecases,
and presentation handlers. Composition is the only layer that imports concrete
implementations from infrastructure, application, and presentation together.
Keep business rules and request handling in their existing layers.

## Main composition

`main/launcher.go` exposes Launch() int, invoked by cmd/main/main.go. Load global
configuration and embedded seed definitions first, then initialize driver,
infrastructure, and application. Seed before presentation and maintenance startup. Use private structs holding each layer's dependencies.

- driver.go: slog JSON/text output, PostgreSQL pool, Gin engine, and http.Server.
- infrastructure.go: interface-typed repositories and utility adapters.
- application.go: interface-typed usecases and explicit constructor wiring.
- seeder.go: clone embedded user/client definitions and inject runtime credentials
  from config.Env before constructor wiring; never mutate global SeedData.
- presentation.go: handlers, middleware usecase dependencies, and route setup.
- launcher.go: signal handling and lifecycle; RunServer provides a testable HTTP
  serving/draining boundary without prematurely closing application resources.

Follow concatenated package names, explicit import aliases, existing Id spelling,
and readable blank lines. No reflection, service locator, or dependency container.
Reuse one repository Pgxdt executor and the strict application Transactor; do not
replace its rejection of nested application transactions.

## Runtime

Use config.Env and config.App; never parse environment variables elsewhere.
Create only slog, PostgreSQL, and HTTP dependencies currently implemented. HTTP
binds the configured PortHttpServer with typed read-header/read/write/idle/shutdown
timeouts. Disable implicit trusted proxies and Gin's default request logger.

Startup failures return 1; normal signal shutdown returns 0. HTTP drains before
PostgreSQL closes. Force-close HTTP connections when the shutdown timeout expires.
Use explicit infrastructure logger calls after infrastructure initialization;
bootstrap configuration and driver failures use slog directly. Do not dump
configuration or connection strings. Docker entrypoint applies migrations before
Launch; native Go startup requires migrations externally. Run the application
seeder when config.Env.SeederEnabled is true, and fail startup on seeding errors.

Embedded seed definitions provision the control-space baseline. Its slug comes
from seederdata.SeedData.Space.Slug and is passed to management constructors even
when seeding is disabled. Google OAuth and handoff wiring are included; gRPC and
Redis remain outside this composition. Future additions must follow the same layer boundaries.

Run make swagger after annotation changes. Verify go test -race ./..., go vet
./..., and go build ./.... Tests for lifecycle and startup failures live under
tests/composition; PostgreSQL tests use isolated schemas and
LOCKMATE_TEST_DATABASE_URL.

Google is enabled explicitly through config.Env. HTTP defaults to 10701 behind
cloudflared; the public HTTPS origin config determines launch/callback URLs.
Wire AuthFlow random generation, PostgreSQL launch/attempt/handoff repositories,
Google OAuth, Google usecases, generic Handoff, and misc expiration cleanup.
Cleanup runs at startup and on a timer with bounded work/deadlines; cancel and
join maintenance before closing PostgreSQL. Never log OAuth settings or secrets.
