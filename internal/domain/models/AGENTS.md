# Domain models

Define the data and error vocabulary shared by contracts and application code.
Keep these models independent of infrastructure implementations. Database
drivers, SQL builders, JWT library claims, bcrypt, and slog do not belong here.

## Layout

- `app/`: application environment, package information, and trace values.
- `entity/`: persistence entities, shared audit fields, access/refresh claims, and
  trusted client authentication identity (in `client.go`).
  Use one file per entity; keep audit structs together in `audit.go` and both
  claim structs together in `claims.go`.
- `error/`: the structured error wrapper and sentinel errors grouped by
  responsibility, such as repository, validator, password, token, and tracer.
  Google OAuth errors distinguish invalid attempts, denial, invalid identity,
  and provider failures.
- `logger/`: format, level, and metadata definitions in separate files.
- `seeder/`: typed embedded bootstrap definitions, one file per entity; Data
  assembles them without importing the embedded-data loader. Seed User.Password
  and Client.SecretKey are runtime-only fields with json:"-"; JSON contains no secrets.

Use the existing package names: `domainmodelsapp`, `domainmodelsentity`,
`domainmodelserror`, and `domainmodelslogger`. Prefer simple value types and small
methods. Existing UUID trace fields are domain values; do not introduce driver
or framework state into models.

## Entity conventions

Use `int64` identifiers and the existing `Id`, `UserId`, and `SpaceId` spelling.
Use pointers for nullable database values, `time.Time` for required timestamps,
`*time.Time` for nullable timestamps, and `json.RawMessage` for JSONB values.
Keep `db` and `json` tags in snake_case and use `omitempty` where the existing
nullable representation calls for it.

Embed the existing audit structs instead of duplicating fields. Most entities
use `AuditCreateUpdateDelete`; assignment entities use creation audit fields.
Access repeated audit names explicitly, such as `AuditCreate.At` and
`AuditUpdate.At`, because embedded `At` and `By` fields are ambiguous.

Users contain profile data. Username, password hash, email, and phone belong to
`ClassicAuth`; Google identity and its email snapshot belong to `GoogleAuth`.
Memberships connect global users to spaces. Roles and permissions belong to
spaces, and assignment entities connect memberships to roles and roles to
permissions. Clients contain their space, identity, secret hash, and redaction.
ClientOrigin stores a client-owned canonical browser origin, active state, and
creation/update/deletion audit fields. Its space is derived through the client;
do not embed origin collections in Client or ClientIdentity.

Keep repository-specific filters, arguments, summaries, and joined projections
in repository contracts. Entity models represent stored records. Some entity
structs contain credential hashes with JSON tags: never treat an exact entity
as an automatically safe public response. Presentation code must choose safe
fields, and repository list contracts use summaries that omit hashes.

## Claims, trace, and logging

Access and refresh claims use domain fields rather than JWT library types.
Both identify the user, space, and space member; access claims also carry role
and permission slugs. Keep them usable for users without classic authentication.

`app.Trace` contains `EventId`, `ClientExecId`, and `ClientSlug` with snake_case
JSON tags. Context storage belongs to the tracer adapter, not the trace model.
Keep environment parsing and defaults in configuration code. `app.Env` contains
logger settings, server ports and HTTP lifecycle timeouts, PostgreSQL settings,
JWT secrets/lifetimes, bcrypt
cost, the startup seeder toggle, three seed passwords, and the seed client secret. Control-space slug comes from space.json,
not the environment. Domain fields use standard Go types,
including time.Duration, uint16 ports, and an int32 PostgreSQL connection limit.
Do not include driver-specific configuration objects or log the entire Env model.
Configuration uses a global config.Env value. LoadEnv loads an optional .env
file without overwriting existing process variables, assigns each field through
typed envGet helpers using its existing value as fallback, and calls validate()
at the end. Required/invalid settings use environment sentinels from
error/validator.go.

`entity.ClientIdentity` is produced by client authentication and contains client ID,
space ID, and slug. It is not a request DTO; usecases recheck the live client and
space. JWTs remain space-scoped and do not embed this identity.

Logger format and level models own their valid values and small conversion or
ordering helpers. `Meta` is a map of string keys to arbitrary values. Formatting,
trace extraction, and output remain in logger infrastructure.

## Error conventions

`Error` carries `Message`, `Type`, and optional `Source`. Construct it through
`NewError`. Preserve the source when wrapping failures so `Unwrap` allows callers
to inspect the original error. Its `Is` method classifies the domain sentinel.

Group exported sentinel errors by responsibility and follow the existing
uppercase snake_case error-code strings. Reuse a suitable classification, such
as `Forbidden`, before introducing a new one. Entity-specific not-found errors
do not automatically match a generic not-found sentinel.

Use `errors.Is` for classification and `errors.As` for structured errors or native
causes. `Error()` includes the source text and is intended for diagnostics;
presentation must select a suitable message rather than expose it wholesale.
Do not put passwords, raw keys, or tokens in messages or sources created here.

Match the existing field ordering and explicit import aliases. Keep related
fields readable with blank lines, and format changed Go files with gofmt.

Client.AuthRedirectUri is nullable, separate from CORS origins, and exposed in
safe client views. GoogleOAuthLaunch, GoogleOAuthAttempt, and AuthHandoff are
internal expiring records whose fields have json:"-". They store digests and
trusted scope; handoffs have identity references, not tokens. Env includes public
HTTP origin, optional Google credentials, flow lifetimes, provider timeout, and
cleanup interval. Never serialize these configuration or transient records.
