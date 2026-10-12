# HTTP presentation

Expose Lockmate's existing application usecases through Gin. Follow the HTTP
organization of nusapala-things/backend, adapted to Gin's `*gin.Context` handler
signature. Presentation parses transport shapes and maps responses; application
owns permissions, value validation, transactions, and operation logging.

## Layout and style

- `handler/auth`, `handler/profile`, and `handler/management`: private handler
  structs with `NewHandler` constructors receiving domain usecase interfaces.
  Keep dependencies in handler.go and implementations in capability/entity files.
- `handler/version`: expose the app information supplied by composition.
- `request`: explicit snake_case body DTOs, preserving pointers and set-null flags.
- `response`: safe wire models and explicit entity/join mappers. Never serialize
  a persistence entity or diagnostic error directly. Normalize empty JSON values
  to `{}` and empty lists to `[]`.
- `route`: route handler interfaces followed by explicit registrations. route.go
  owns grouping, middleware ordering, documentation, and JSON 404/405 responses.
- `middleware`: recovery, tracing, dynamic CORS, client authentication, and bearer
  authentication. Middleware calls domain usecase interfaces only.
- `utils`: JSON binding, raw-shape parsing, pagination, private context keys, and
  the centralized error mapping table. Helpers do not log.

Use concatenated package names and explicit domain/presentation import aliases.
Keep `Id` spelling and int64 entity IDs. Handlers use `<Resource><Action>` names,
explicit usecase argument structs, direct error handling with Error(c, err), and
blank lines between parsing, application calls, and response shaping. Avoid
reflection, generic CRUD handlers, repository dependencies, and transport-side
permission slugs.

## Request pipeline

Recovery -> trace injection -> global CORS -> business client authentication and
client-specific origin check -> bearer authentication on profile/management ->
handler -> application. Preserve the request's context, cancellation, and deadline.

Every business route requires one X-Client-Slug, X-Client-Secret, and
X-Client-Exec-Id header. Execution IDs must be nonzero UUIDs; they are tracing
metadata, not authenticated identity. Client authentication produces the trusted
ClientIdentity. Bearer authentication uses Session.AuthenticateAccessToken so
permissions and live membership are reloaded. Never build trusted identity or
claims from DTOs, query parameters, unsigned JWTs, or role names.

Swagger, version, and CORS preflight do not require client credentials. Tracing
still generates an event ID; absent/invalid diagnostic metadata stays zero until
business middleware rejects an invalid execution header. Secrets are preserved
exactly, never trimmed, dumped, logged, or included in error responses.

## Dynamic CORS

Global origin lookup happens before authentication because browser preflight
cannot carry client credentials. Actual requests with Origin also check the
trusted client's own origin list. Missing Origin permits normal server-to-server
requests; it does not bypass authentication. Origin lookup errors stop processing.

Use only exact canonical persisted origins. Do not add wildcard, same-host,
localhost, or hardcoded development exemptions. CORS cookies/private-network
access are disabled. Preflight returns 204 without invoking authentication or
business handlers; max-age is zero and allowlists are not cached. Preserve Vary
headers. CORS is not client authentication.

## DTOs, errors, and documentation

Bind one JSON object with application/json, up to 1 MiB. Parse required fields,
int64/UUID/boolean shapes, and query values. Application validates business values
and ownership. Unknown body fields cannot become audit actors, hashes, or scope
because handlers map only declared fields. Explicit set-null flags clear nullable
fields; omission leaves pointer updates unchanged.

Pagination defaults to page 1/limit 10, caps limit at 100, and normalizes
nonpositive page/limit values like the reference. Preserve all usecase filters.
Creates return 201, reads/session refresh/sign-in/secret rotation return 200,
and ordinary mutations return 204. Tokens and one-time secrets use no-store.

Classify every domain sentinel with errors.Is. Extract safe Message with errors.As;
never expose Error() or Source. Unknown/internal errors return a fixed sanitized
500. Recover panics without Gin's credential-containing request dump. Do not add
presentation request logging.

Each handler needs accurate Swaggo annotations, brace-form path parameters,
example tags, and AND-combined client security requirements (plus BearerAuth on
protected operations). `make swagger` runs the pinned generator and updates
`docs/swagger/{docs.go,swagger.json,swagger.yaml}`; never hand-edit those outputs.
Swagger is served at /api/docs/index.html, with base path /api and business
operations under /api/v1. Keep resource wildcard names consistent throughout the
Gin tree: space_id, client_id, space_member_id, and origin_id, for example.

Tests live under tests/presentation/http. Cover route/usecase completeness,
trusted context and DTO mapping, error sanitization, Swagger/runtime parity,
preflight, origin revocation, and real PostgreSQL authentication. Logout/revocation, forgotten-password recovery, hosted login UI, and object storage
remain deferred; existing Google identity monitoring is available.

Google start and generic handoff exchange require the normal client headers.
Google launch/callback are browser navigation routes without those headers or
bearer authentication; saved state and the browser-binding cookie authorize the
callback. Use attempt-specific __Host cookies, Secure/HttpOnly/SameSite=Lax with
Path=/ and no Domain. Keep credentialed CORS disabled. Callback query values must
be unambiguous; read binding only from its cookie. Responses use no-store and
no-referrer. Redirects carry short-lived handoff codes, never JWTs.
