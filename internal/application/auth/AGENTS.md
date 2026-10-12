# Authentication application

Implement the Client, Session, and Google domain usecases in `client/`, `session/`,
and `google/`. Each folder has `usecase.go`, a private usecase struct with explicit
dependencies, and NewUsecaseImpl returning its domain interface. Match Nadi's
context-first methods, import aliases, early returns, and readable whitespace.
Google browser initiation lives in launch.go beside its usecase.go.
Shared identity/session helpers live in `../shared`; concrete adapters are not
production dependencies of application.

Clients identify spaces, not user authority. Recheck live client IDs/slug and its
active space. Validate JWTs for that space and resolve current live identity,
membership ownership, and grants. Refresh and access authentication never enroll.
Only credential-authenticated sign-in can create a missing membership; suspension
never becomes activation. Registration returns both tokens after committed
creation of profile, authentication identity, and default-role membership.

Use the strict domain Transactor backed by pgxdt.NewTransactor from the same pool
as repository pgxdt.NewPgxdt. Classic sessions generate tokens inside the transaction and discard
them on failure; Google callbacks commit a handoff instead, with JWTs generated
inside the generic handoff exchange transaction. Retry enrollment conflict once with a fresh transaction. Warn
directly at the failed step and at the retry handler immediately before a retry,
with the next attempt number;
exhausted retries emit the specific step error and the outer transaction error.

Google provider exchange/attempt consumption occurs once before persistence.
Match verified subject, never email. Retry subject/membership conflicts once
without repeating exchange. Warn only when another attempt will occur. Synchronize
Google contacts only; preserve classic contacts and existing profile edits. New
users use a valid provider name or
`Google user`, without copying avatar URLs into storage paths.

Write logger.Error/Warn/Info directly at each handling point with a specific
message and inline err/ID metadata. Do not use deferred logging, shared logging
wrappers, or generic failure messages. Log callback errors at their step and the
outer transaction error with different messages. Recoverable conflicts use warn
at the step and retry handler; only warn about retrying when another attempt will
occur. Access-token and refresh-token generation have separate calls and failure
messages, and outputs are assigned only after both succeed.

Log committed registration/enrollment/contact changes explicitly at info;
successful read-only authentication stays quiet. Native error details are retained
for debugging; never add credentials, tokens, provider codes, attempts, or request
bodies as metadata. Warn directly when an invalid supplied Google name falls back
to the default registration name; an absent name stays quiet. Separate operation
steps with blank lines. Tags follow auth/<capability>/<MethodName>.

Tests live in `tests/application/auth`; isolated PostgreSQL tests require
LOCKMATE_TEST_DATABASE_URL. Provider infrastructure stays behind utility contracts; linking, revocation, OTP,
and forgotten-password workflows remain deferred.

Google Start creates a launch ticket against the client's configured auth_redirect_uri.
Launch consumes it and generates browser binding before calling GoogleOAuth.Begin.
The verified callback rechecks client/space and unchanged URI, then creates an
AuthHandoff atomically with identity changes. Validated denial returns a fixed
error redirect; invalid attempts never redirect. No tokens are issued at callback.
Handoff.Exchange authenticates ownership, consumes in the application transaction,
reloads live identity/grants, and generates tokens before commit. It is independent
of Google enablement so later classic browser flows can reuse it.
