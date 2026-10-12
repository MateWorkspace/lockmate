# Lockmate

Lockmate is the shared authentication and authorization service for Mate projects.
This repository is the Go rewrite of the earlier Rust implementation in
`lockmate_old`, refined around explicit clean architecture boundaries. It manages
global user identities, project spaces, memberships, space-specific roles and
permissions, and the clients that consume the service.

Users may authenticate with classic credentials or Google. Clients authenticate
with their own slug and secret and determine the target space. User sessions use
access and refresh JWTs; application checks reload live identity and grants.
Google sign-in also registers new users and returns a single-use authentication
handoff to the client's configured backend callback.

## Read the scoped guides

Before changing code, read this guide and every applicable `AGENTS.md` along the
path to the files you will edit. Also read the guides for related layers when a
change crosses models, contracts, application, infrastructure, or presentation.
Find guides with `rg --files -g AGENTS.md`. Read neighboring implementations to
preserve their layout, naming, signatures, and handling patterns.

The current scoped guides are:

- Domain: [contracts](internal/domain/contracts/AGENTS.md),
  [models](internal/domain/models/AGENTS.md), and
  [usecases](internal/domain/usecases/AGENTS.md).
- Application: [auth](internal/application/auth/AGENTS.md),
  [profile](internal/application/profile/AGENTS.md),
  [management](internal/application/management/AGENTS.md),
  [misc](internal/application/misc/AGENTS.md),
  [shared](internal/application/shared/AGENTS.md), and
  [permissions](internal/application/permissions/AGENTS.md).
- Infrastructure: [repository](internal/infrastructure/repository/AGENTS.md) and
  [utility](internal/infrastructure/utility/AGENTS.md).
- Delivery and wiring: [HTTP presentation](internal/presentation/http/AGENTS.md)
  and [composition](internal/composition/AGENTS.md).

Treat scoped guides as refinements of this overview. Keep them accurate when
responsibilities, interfaces, or behavior change. Follow explicit user instructions
when they override repository guidance. Establish new patterns deliberately and
align all affected layers rather than introducing isolated conventions.

## Repository layout and dependencies

- `cmd/main/`: executable entrypoint delegating startup to composition.
- `internal/config/`: global application information and environment loading.
- `internal/domain/models/`: entities, environment/trace values, logger models,
  and classified errors.
- `internal/domain/contracts/`: repository and utility interfaces consumed by
  application; argument structs follow interface declarations.
- `internal/domain/usecases/`: application-facing interfaces and request/result
  types, grouped into `auth`, `profile`, `management`, `misc`, and `shared`.
- `internal/application/`: business orchestration implementing usecases through
  contracts. Permission constants live in `permissions`; focused helpers in
  `shared`.
- `internal/infrastructure/`: concrete repository and utility adapters. PostgreSQL
  repositories separate query building from execution and error handling.
- `internal/presentation/http/`: Gin handlers, DTOs, response mapping, routes,
  middleware, and transport helpers.
- `internal/composition/main/`: explicit driver, infrastructure, application,
  presentation, and lifecycle wiring.
- `database/migrations/`: ordered SQL up/down migrations.
- `docs/`: ERD, generated Swagger, and the Google integration guide.
- `tests/`: behavior and integration tests organized by layer and capability.

Domain defines the vocabulary and interfaces. Application depends on those
interfaces, infrastructure implements contracts, and presentation calls usecase
interfaces. Composition connects concrete implementations. Keep driver, framework,
and provider SDK types outside domain contracts; presentation must not import
repository/utility contracts or concrete application/infrastructure packages.

Use PostgreSQL through pgx and Squirrel, without an ORM. Transactions use the
shared `mate-golang/pgxdt` module: application owns the top-level transaction;
repository helpers reuse its context. Application transaction nesting is rejected.
The module currently has a local replacement in `go.mod`; preserve the sibling
`mate-golang` checkout for local builds.

## Code style and behavior

Use idiomatic Go with the established explicit structure: private implementation
structs, constructor dependency injection, interface return types, concatenated
package names, and explicit import aliases. Keep `Id`, `UserId`, and `SpaceId`
spelling. Use argument structs for long signatures, one argument per line for
long constructor calls, and blank lines between logical steps. Format with gofmt.

Pass `context.Context` first for contextual operations and retain cancellation,
deadlines, tracing, and the transaction context. Value-oriented utilities retain
their existing signatures. Preserve space ownership and immutable identifiers.
Use pointer updates and explicit null flags consistently; never accept conflicting
replacement and clearing requests.

Protected application functions declare their tag and check PermissionGuard
immediately afterward, before identity reads or transactions. Permission slugs
come from the centralized catalog. Authentication establishes client/user identity
through credentials and live records rather than role-name shortcuts.

Return classified domain errors and retain useful native causes. Application
writes direct, operation-specific error logs at failure points, warnings for
best-effort handling, and success logs after committed creates, deletes, and
important updates. Do not replace these with deferred or shared logging wrappers.
Logging uses the logger contract backed by slog. Never log credentials, tokens,
flow secrets, request bodies, or complete configuration/results containing secrets.
Presentation maps errors to safe responses and maps explicit response DTOs rather
than serializing persistence entities or native errors.

## Runtime and verification

Load settings through global `config.Env` and `LoadEnv()`. An optional local `.env`
is supported without overriding process variables. Document settings in
`.env.example`; keep local secrets out of Git.

HTTP defaults to port `10701`. Development uses
`https://lockmate-dev.dhonan.my.id` through cloudflared. Business API calls require
`X-Client-Slug`, `X-Client-Secret`, and a nonzero UUID `X-Client-Exec-Id`; protected
user operations also require a bearer token. Dynamic CORS checks persisted client
origins. Google browser launch/callback instead validate saved state and an
attempt-specific browser-binding cookie. Client `auth_redirect_uri` is independent
of CORS origins. See [Google setup and integration](docs/google-auth.md).

Google is optional and disabled until configured. Migrations and initial
client/space/role/permission records are provisioned externally; startup does not
create them. Expiring auth records are cleaned periodically, while expiration is
also enforced during consumption. gRPC presentation, Redis integration, seed
tooling, hosted login UI, OTP/recovery, and logout/revocation remain future work.

For Go changes, run from this repository:

```sh
go test -race ./...
go vet ./...
go build ./...
```

PostgreSQL integration tests require `LOCKMATE_TEST_DATABASE_URL` and use isolated
schemas. Without it, database tests skip; distinguish that from verified database
behavior. Keep tests under the corresponding `tests/` layer. Regenerate Swagger
with `make swagger` after annotation or wire-model changes; never hand-edit
`docs/swagger` outputs. Documentation-only changes need content, link, and
whitespace checks rather than a full Go test run.
