# Domain usecases

## Purpose and current scope

Define application behavior through interfaces and request/result values.
Implementation belongs in application; domain usecases must not import
infrastructure, transport bindings, database drivers, or provider SDKs.

Application implements the interfaces and enforces the rules below. Interfaces
alone do not enforce authorization. Google adapters/configuration are wired through composition; embedded seed definitions are applied through the startup Seeder.
Gin presentation and main composition expose the implemented classic, profile,
and management operations.

## Layout and style

- `auth/`: `client.go`, `session.go`, and `google.go` for client authentication,
  classic sessions/registration, and Google initiation/callback.
- `profile/`: `account.go` and `security.go` for the caller's own profile,
  membership preferences, authentication summaries, and classic credentials.
- `management/`: one interface file per current entity; Google identities expose
  monitoring reads only. Assignment entities retain separate files.
- `misc/`: small helper usecases exposed to presentation through application,
  including tracer injection, the pre-authentication allowed-origin helper, the
  public JWK set, and signing-key reconciliation.
- `shared/`: `Page[T]` and credential-free classic-auth/client views that reuse
  existing repository summaries.

Use concatenated package names such as `domainusecasesmanagement` and explicit
domain import aliases. Put the interface before its request/result structs.
Request names are unique within the group, such as `CreateRoleRequest` and
`FilterRoleRequest`. Preserve existing `Id`, `UserId`, and `SpaceId` spelling.
Keep logical sections separated by blank lines; format Go changes with gofmt.

Every operation receives context first. Protected operations receive trusted
`*AccessClaims` next, then explicit target scope/IDs where applicable. Session
operations receive a trusted `*entity.ClientIdentity` instead. Presentation cannot
construct either identity from request fields or unsigned claims. Trace fields
are diagnostics, not authenticated identity or numeric audit actors.

Repository-oriented creates return an ID; mutations return an error. Exact reads
return pointers. Filter results use `Page[T]` with items and total, preserving
repository pagination semantics. Client creation/rotation return metadata and
the raw secret once. Session operations return one shared `TokenResult` (tokens
plus `IsNewUser`); the internal identity/grant snapshot stays unexported.

Requests cannot accept audit actors or hashes. Optional updates use pointers;
explicit SetNull flags clear nullable fields. A value and its clear flag together
are invalid. Slugs, ownership IDs, and Google subjects remain immutable.
Preferences replace JSON values; avatar paths change metadata without storage I/O.

## Presentation helpers

Misc usecases provide application entry points for small utility operations that
presentation needs. Presentation calls the usecase rather than a utility contract
or concrete infrastructure adapter. Application implementations delegate
through injected domain utility contracts.

Misc Tracer receives only client slug and execution UUID from presentation. It
generates a fresh event UUID through the tracer utility, assembles the domain
Trace, and returns the injected context plus an error. Presentation cannot choose
the event UUID. Preserve parent values, cancellation, and deadlines; empty slug
and zero execution UUID remain absent metadata. Generation failure does not
inject and emits one sanitized error; success stays quiet. Trace extraction
remains an internal utility operation. Add other helpers only
when presentation has a concrete need; keep unrelated helpers in separate files.

## Client and session authentication

Client authentication looks up the global slug, compares the supplied secret
through Keygen, and requires a live active client and parent space. Its identity
contains client ID, space ID, and slug. Session operations recheck these values
against live records and derive their space from the client, never a request
override. Unknown or disabled identity and mismatched secrets return generic
Unauthorized; operational failures keep their actual error classification.

Tokens retain the existing space scope. Different authenticated clients in the
same space may use the same user's tokens. Client identity is not stored in JWT
claims and client authentication alone grants no user permissions.

Classic SignIn uses username/password. Check credentials before enrolling a
missing membership; repository membership creation assigns the current default
role atomically. Never reactivate suspended memberships. Missing credentials or
wrong passwords return generic Unauthorized without exposing their source.

Register validates name/credentials/optional contacts, hashes the password, and
creates user, classic identity, membership, and default-role assignment in one
transaction. Registration cannot set roles, activation, or verification flags.
Generate both tokens before committing and expose them only after successful
commit. Set IsNewUser only for a new global user, not a new membership.

Refresh and AuthenticateAccessToken validate tokens against the client's space,
then resolve the live user, active space, active membership, and membership
ownership. Reload current grants rather than trust stale signed snapshots.
Neither operation enrolls users. Refresh returns a new pair without revoking the
old refresh token. Preserve TokenInvalid/TokenExpired and operational failures.

Access tokens are ES256-signed by the persisted active signing key; issuance
resolves and opens that key. Misc.SigningKey.Reconcile creates the active key
when missing, rotates a stale active key, re-seals keys sealed under a previous
key-encryption key, and removes keys past retention. Misc.Jwks exposes the active
and retiring public JWKs. Both are invoked by composition-owned scheduling, never
by HTTP callers.

## Google authentication

Use the GoogleOAuth utility contract, never a concrete provider adapter.
Start rechecks the authenticated client and its configured auth_redirect_uri,
then saves a short-lived launch ticket with project correlation state. Launch
consumes the ticket and generates trusted browser binding. The GoogleOAuth adapter
owns expiring single-use state, nonce/PKCE, provider exchange, and verified claims.
Binding comes from the Lockmate cookie rather than an arbitrary callback parameter.

Callback receives code/state/binding or a provider error. Complete validates the
attempt and binding before provider work, consumes the attempt even on denial,
and returns verified identity with original client/space IDs. Recheck that client
and space before using them. A provider error and code together are BadArgs;
state and binding are required, and code is required absent a provider error.
Do not include provider error text, state, codes, or bindings in returned messages.

Match by Google subject only. Unknown subjects create a new global user and
Google identity even if their email matches a classic account. Explicit linking
is deferred. A valid provider display name initializes the user; otherwise use
`Google user`. Leave avatar storage unset. Existing user profile edits and classic
contacts are never overwritten. Synchronize email and its verification flag only
on GoogleAuth; unverified provider emails are rejected before identity persistence.

Finish provider communication before the database transaction. Identity creation,
Google contact synchronization, missing membership enrollment, and handoff creation
are atomic. Return a code redirect only after commit. Retry subject/membership
conflicts once with a fresh transaction, without repeating provider exchange.
Generic Handoff.Exchange reloads live identity/grants and generates both tokens
inside the single-use consumption transaction. Return tokens only after commit.

Classify state/binding/replay failures as GoogleOAuthStateInvalid and invalid
verified-token claims as GoogleOAuthInvalid. Validated provider denial returns a
fixed error redirect, while provider operational failures use GoogleOAuthFailed.
Native provider diagnostics must be sanitized because SDK errors can contain
codes, token response bodies, and credential-bearing URLs.

## Profile ownership

Derive user, space, membership, and audit actor IDs exclusively from the caller.
GetProfile includes global user, current space, and current membership. Roles and
permissions are current effective grants. Global profile and classic-credential
edits affect the user across spaces; membership preferences affect this space only.

GetAuthMethods returns optional safe classic and Google summaries. Missing methods
are absent values, not failures; database failures still propagate. Classic
operations require an existing classic identity and return ClassicAuthNotFound
for a Google-only user. Contact changes reset verification in repository SQL;
profile cannot override flags. ChangePassword requires CurrentPassword and returns
PasswordMismatch when comparison fails before hashing/storing NewPassword.

No profile request may activate membership, choose grants, override ownership,
add/delete authentication methods, or set verification flags.

## Management and permissions

Control-space slug comes from management constructor arguments, defaulting to `lockmate`.
Resolve it by immutable slug; never grant authority just because a role is named
super or admin. Global User, Space, ClassicAuth, and GoogleAuth management requires
control-space permissions. Scoped management allows either target-space authority
or control-space authority; callers from another ordinary space are forbidden.
Client ReadBySlug also verifies target-space ownership despite global lookup.

Before protected operations, recheck the caller's live user, active space, active
membership, and membership ownership. Incoming permissions must come from the
trusted AuthenticateAccessToken path that reloads effective grants per request.
Call PermissionGuard with those permissions. Missing/inactive caller identity is
Unauthorized; authenticated insufficient authority is Forbidden. Deleted targets
follow repository visibility rules; inactive targets remain manageable.

| Operations | Required permission |
| --- | --- |
| Management exact/list/filter reads | `<entity>:get` |
| Management creates | `<entity>:add` |
| Management updates and default-role changes | `<entity>:set` |
| Management deletes, including bulk assignments | `<entity>:remove` |
| ClassicAuth.ResetPassword | `user_password:set` |
| Client.RotateSecret | `client_secret:set` |
| SpaceMember effective grant reads | `user_permission:get` |
| Profile and own grant reads | `profile:get` |
| Profile and own membership preference updates | `profile:set` |
| Own authentication summaries | `profile_security:get` |
| Own classic credential/password changes | `profile_security:set` |

Entity prefixes are space, permission, role, user, classic_auth, google_auth,
client, client_origin, space_member, role_permission, and space_member_role.
Embedded seed definitions follow this catalog; independent tests verify parity.

User.Create atomically creates a profile and required classic credentials; it
does not issue tokens or enroll membership. ClassicAuth.Create can provision a
Google-only existing user. ResetPassword addresses a user ID, validates/hashes
the replacement, and does not require the old password or implicitly create
credentials. ClassicAuth deletion requires a remaining live Google identity;
otherwise return BadState. GoogleAuth is monitoring-only: no arbitrary subjects,
provider contact edits, linking, or deletion are exposed.

Client creation/rotation generates a secret through Keygen and stores only hash
and redaction. Return raw secret once after commit. Ordinary updates cannot set
secrets or hashes. Default-role replacement and membership creation retain
repository atomicity/lock ordering. Bulk assignment deletion requires a selector;
two selectors use OR within the target space. User deletion retains the
repository's atomic dependent cleanup.

Monitoring covers current records, authentication summaries, memberships, and
grants, not historical logins or persisted sessions.

## Dependencies, errors, and application logging

Use matching repository contracts, Validator, PermissionGuard, Logger, and
Transactor. Client/secret operations also use Keygen; classic credentials use
Password; session issuance/validation uses Token; Google uses GoogleOAuth.
Protected operations use live User, Space, and SpaceMember repositories.
Additional grant resolution uses SpaceMemberRole. Do not introduce driver state
into context models or import concrete adapters into application.

Pass the transaction callback context to every repository operation. Propagate
errors to roll back, including nested failures. Credential-producing feature
operations own their top-level commit boundary. The strict application transactor
rejects an existing transaction with BadState; reusable repository transactions
still compose within its callback. Returned credentials represent committed writes.

Reuse domain errors and preserve native Source. Classify through errors.Is and
inspect with errors.As. Public responses use a safe Message/classification, never
the full Error() source chain. Never log request structs, passwords, hashes,
tokens, raw secrets, OAuth codes/state/bindings, or provider token responses.

Application implementations write explicit error logs at each failure handling
point. Transaction callback failures also receive a distinct outer transaction
failure log; shared helpers do not log. Log successful CREATE/UPDATE/DELETE operations only
after commit, including registration, enrollment, password updates, contact sync
when changed, and secret rotation. Routine reads, successful authentication
without writes, and successful token validation remain quiet.

## Verification and deferred runtime acceptance

Run gofmt on changed Go files, then `go test ./...`, `go vet ./...`, and
`go build ./...` from the Lockmate root. Check the complete operation catalog,
domain dependency direction, and absence of hashes/audit override fields from
feature requests and results. Interface declarations alone cannot prove runtime
business behavior.

Application tests cover registration token issuance and rollback;
first Google registration; automatic enrollment once; suspended membership
denial; callback replay/binding mismatch; concurrent identity/membership creation;
grant refresh without enrollment; old-password requirements versus admin reset;
cross-space denials; last-method deletion protection; and one-time secret output.

Forgotten-password recovery, OTP/verification delivery, authentication-method
linking/unlinking in profile, self-deletion/space departure, persisted sessions,
server-side logout/revocation, API keys, caching, and object storage are deferred.

## Client origins

Client management owns origin administration through CreateOrigin,
ReadOriginById, ReadOriginByValue, ReadOriginsByClientId, ReadOriginsByFilter,
UpdateOriginById, and DeleteOriginById. All require explicit target space/client
IDs and separate client_origin:add/get/set/remove permissions. Requests cannot
supply ownership or audit actors. Exact origin operations verify client ownership;
inactive clients and origins remain manageable. Validate canonical origins with
the validator utility. Mutations own strict transactions and derive audit actors
from the authenticated caller. Origin and activation changes are access updates.

Misc AllowedOrigin exposes global and client-scoped checks before authentication.
It delegates through ClientOrigin and Validator contracts without a permission
check or transaction. Known invalid/oversized origins return false without error,
logging, or database access. Operational failures remain errors and emit explicit
logs; successful matches and ordinary denials stay quiet. Scoped IDs must be
positive. A match does not authenticate or identify a client. Gin/CORS wiring and
caching are outside these usecases.

Google now exposes Start, Launch, and SignInCallback. Start accepts project state,
not a redirect URL; Launch supplies internal browser-cookie data; callback returns
a safe redirect containing a handoff. Auth.Handoff.Exchange returns SessionResult
only after committed single-use consumption and fresh token generation. Misc.AuthFlow
provides bounded expiration cleanup without presentation-specific types.

Startup Seeder.Run is separate from HTTP usecases. It depends on contracts and
seed models, owns an atomic bootstrap boundary, and is invoked only by composition.
Its interface is under seeder/. It does not require user claims or grant a transport
permission bypass. Read its scoped application/data guides before modifying it.
