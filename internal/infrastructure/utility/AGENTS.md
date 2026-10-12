# Utility infrastructure

Implement `internal/domain/contracts/utility` using concrete libraries and
standard-library facilities. Keep business orchestration and persistence out of
these adapters. Utilities return domain models and errors through their contracts.

## Layout and style

- `keygen/generator.go`: random secret generation, hashing, redaction, and checking.
- `password/bcrypt.go`: password hashing and comparison.
- `transactor/pgxdt.go`: adaptation of the shared mate-golang transaction helper.
- `token/jwt.go`: access and refresh JWT generation and validation.
- `validator/validator.go`: all validator contract methods; `utils.go`: shared
  helpers, including URL and address checks for client origins.
- `permission_guard/claims.go`: permission checks against access claims.
- `tracer/context.go`: UUIDv7 event generation and context injection/extraction.
- `logger/slog.go` and `normalize.go`: slog output and metadata normalization.

Use `infrastructureutility<capability>` package names, with concatenated words
for multiword capabilities. Concrete structs are private and dependencies are
explicit fields. Constructors follow `New<Implementation>Impl` and return the
domain interface. The transactor wrapper preserves callback errors and maps
strict pgxdt lifecycle failures into domain classifications.

Use explicit domain import aliases, constructor-before-method ordering, early
returns, and blank lines between validation, computation, and return steps.
Keep shared helpers beside their implementation. Do not introduce a general
helper package for code used by only one utility. Format Go changes with gofmt.

## Context, transactions, and logging

Tracer uses a private context key and stores the domain trace as a value.
Injection derives a context preserving cancellation and deadlines; extraction
returns a copy or the domain trace-not-found error. Do not generate trace IDs
at presentation. GenerateEventId uses UUIDv7 and preserves generation failures
as TraceEventIdFailed with their native cause.

Transactor wraps `mate-golang/pgxdt.NewTransactor`. Repositories and transactor use
the same pool. Pass the callback's context through every operation
in a transaction. Application nesting is rejected as BadState; repository
NewPgxdt calls retain reuse without savepoints. Preserve callback errors unchanged;
classify lifecycle deadlines as Timeout, cancellation as BadState, and other
driver failures as Failure, preserving native causes. Propagate errors to roll back.

Logger uses slog exclusively. Configure JSON or text output through the supplied
handler. Respect the domain level threshold, including `NONE`; ensure the handler
also enables the requested level. Calls supply `tag` and normalized `meta` without
mutating the caller's metadata. Convert top-level error metadata to diagnostic
strings and normalize absent metadata to an empty map.

`useTrace` controls extraction and the trace attribute. When enabled, inject a
tracer dependency and use a zero-value trace if extraction fails. JSON trace
fields follow the model's tags. Other utility adapters return errors without
emitting logs; logging occurs when a caller invokes the logger contract.

## Credential and authorization behavior

Keygen uses cryptographic randomness. The current raw key is `lockmate-` followed
by 64 lowercase hexadecimal characters, for a total length of 73. Hash with
SHA-256, preserve only the first and last four characters in the redaction, and
compare hashes in constant time. Random generation does not replace a database
uniqueness constraint. Never log raw keys or accept malformed keys silently.

Password uses bcrypt. Preserve native failure causes and distinguish a mismatch
from malformed hashes or hashing failures. Keep the existing default-cost
fallback and the 72-byte password limit; never truncate a password.

Token uses HS256 with separate access and refresh secrets and lifetimes. Preserve
token-type separation, space audience checks, identity checks, and required time
claims. Map expiration separately from invalid tokens and generation failures.
Domain claims stay independent of JWT library types. JWT validation does not
query live memberships, grants, or revocation state; application code handles
those checks.

Permission guard requires every supplied permission to exist exactly in access
claims. Nil claims are invalid; an empty requirement list passes for nonnil
claims. Do not add role-based bypasses, wildcard expansion, or case folding.
The caller must supply authenticated claims.

Validator checks the existing model fields and returns their corresponding
domain validator errors. Preserve accepted formats, bounds, and nullable-field
handling. Validation must not normalize inputs, write data, or check database
uniqueness. Keep the existing `UserUsername`, `UserEmail`, `UserPhone`, and
`UserPassword` contract names for classic-auth credential validation.

## Errors and verification

Classify failures through `internal/domain/models/error` and preserve useful
native causes. Reuse existing errors where appropriate. Keep passwords, raw keys,
and tokens out of constructed messages and diagnostic metadata.

Tests belong in `tests/infrastructure/utility`. Cover contract behavior and failure
cases, especially token boundaries, credential mismatches, trace toggling, and
transaction context reuse. PostgreSQL tests require an isolated database via
`LOCKMATE_TEST_DATABASE_URL` and skip when it is absent.

Run `go test ./...`, `go vet ./...`, and `go build ./...` from the Lockmate root
when changing Go implementations. Documentation-only changes need a content and
whitespace review.

ClientOrigin validation accepts canonical HTTPS origins and HTTP only for
localhost or loopback IP addresses. Accept valid ASCII DNS/punycode names and
canonical IPs, omit default ports, and reject paths, credentials, queries,
fragments, wildcards, null origins, whitespace, and trailing-dot hostnames.
Nondefault ports are 1–65535; the limit is 2048 bytes. Reject noncanonical input
rather than rewriting it, resolving DNS, or checking persistence.

auth_flow/random.go generates 32 cryptographically random bytes encoded as URL-safe
base64 and SHA-256 hex digests. google_oauth/oidc.go implements provider authorization,
PKCE/nonce, single-use attempt consumption, and ID-token verification through a
repository contract. This utility's transient persistence is part of its OAuth
responsibility; identity persistence remains application-owned. It emits no logs.
Never propagate provider errors containing token response bodies or request URLs.

ClientAuthRedirectUri accepts canonical HTTPS callback URLs with paths. Constructor
injection from config.Env.AuthRedirectAllowHttp enables HTTP callbacks explicitly
through LOCKMATE_AUTH_REDIRECT_ALLOW_HTTP, default false. Do not read env/config
in the adapter. Reuse origin-format validation with HTTP enabled for redirects only. Keep its
existing port, host, credential, query, fragment, and path checks. ClientOrigin's
HTTP policy remains limited to localhost and loopback addresses.
