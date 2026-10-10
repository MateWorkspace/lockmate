# Domain usecases

## Purpose and current scope

Define application behavior through interfaces and request/result values.
Implementation belongs in application; domain usecases must not import
infrastructure, transport bindings, database drivers, or provider SDKs.

Application implements the interfaces and enforces the rules below. Interfaces
alone do not enforce authorization. Google adapters/configuration, presentation,
composition, and seed definitions remain separate follow-up work.

## Layout and style

- `auth/`: `client.go`, `session.go`, and `google.go` for client authentication,
  classic sessions/registration, and Google initiation/callback.
- `profile/`: `account.go` and `security.go` for the caller's own profile,
  membership preferences, authentication summaries, and classic credentials.
- `management/`: one interface file per current entity; Google identities expose
  monitoring reads only. Assignment entities retain separate files.
- `misc/`: small helper usecases exposed to presentation through application,
  starting with the tracer injection mirror in `tracer.go`.
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
the raw secret once. Session operations return one shared `SessionResult`.

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

## Google authentication

Use the GoogleOAuth utility contract, never a concrete provider adapter.
BeginSignIn rechecks the authenticated client and supplies its IDs and a trusted
browser/session binding. The adapter owns the configured callback URI, expiring
single-use attempt state, nonce/PKCE, code exchange, and verified Google claims.
Binding must originate from a trusted browser/session mechanism rather than an
arbitrary callback parameter. It is separate from the public OAuth state value.

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
on GoogleAuth; an unverified email is stored as unverified, not used for linking.

Finish provider communication before the database transaction. Identity creation,
Google contact synchronization, missing membership enrollment, effective grants,
and token generation form one transaction. A missing default role or token failure
rolls back registration. Return tokens only after commit. On a concurrent subject
or membership conflict, roll back and retry once using the winning records; do
not continue in an aborted transaction. Preserve the conflict if recovery fails.

Classify state/binding/replay failures as GoogleOAuthStateInvalid, provider denial
as GoogleOAuthDenied, invalid verified-token claims as GoogleOAuthInvalid, and
provider operational failures as GoogleOAuthFailed. Preserve diagnostic causes.

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
client, space_member, role_permission, and space_member_role. Seed definitions
must later follow this catalog; no seed implementation is introduced here.

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

Application implementations emit one sanitized error log per failed operation.
Helpers do not duplicate it. Log successful CREATE/UPDATE/DELETE operations only
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
