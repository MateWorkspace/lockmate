# Management application

Implement each management interface in its matching entity folder's usecase.go.
Use private structs with explicit contract dependencies and NewUsecaseImpl
constructors returning domain interfaces. Keep execution explicit rather than
introducing a generic CRUD framework. Preserve Nadi's domain aliases, context-first
signatures, early returns, and whitespace between operation steps.

Shared IdentityReader checks live callers and space boundaries after the initial
permission check. Keep permission slugs in application/permissions, not in methods.
Control-space slug comes from embedded space.json (default lockmate), supplied
by composition and resolved by immutable slug. Global user/space/authentication
management requires control authority; scoped operations allow target-space or
control authority. Role names provide no bypass. Inactive targets are manageable.

Every mutation owns a strict application transaction through domain Transactor.
Repositories retain their reusable transaction helper. Pass callback context to
authorization, repository calls, and all multi-step operations. Derive audit
actors from caller.UserId. Return results and log success only after commit.

Validate supplied fields, IDs, pagination overflow, JSON, and null/value conflicts.
Preserve immutable slugs/ownership/subjects and existing repository concurrency
rules. Default-role assignment/replacement and user-dependent cleanup remain in
repositories; assignment bulk deletion preserves scoped OR selectors.

User creation atomically provisions profile and classic credentials without tokens
or enrollment. ResetPassword requires no old password and no implicit provisioning.
Classic-auth deletion requires another live Google method. Google identities are
monitoring-only. Client creation/rotation exposes raw secrets once after commit;
ordinary updates cannot supply secret hashes. Exact and list credential reads
return safe projections. Global client slug lookup must still verify target space.

Tags follow management/<entity>/<MethodName>. Each protected method declares its
tag, then immediately checks PermissionGuard using the centralized operation
constant. Reject permission failures before identity reads, validation, or starting
transactions. Live identity and space-boundary checks still run afterward.

Write logger.Error/Warn/Info directly at each handling point, with a specific
message and inline metadata. Include err and relevant caller/target IDs. Do not
use deferred logging, shared logging wrappers, or generic failure messages. Log
callback failures at their step and transaction failures at the outer handler with
different messages. Native error details are included for debugging; never add
credentials, tokens, request bodies, or preferences as metadata.

Log info explicitly after committed creates/deletes and security/access updates.
Reads and cosmetic/preference-only updates stay quiet. Status/default updates
emit info when IsActive/IsDefault is supplied, including mixed requests. Log
password resets, classic credential changes, secret rotations, and default-role
replacements only after commit. Separate checks and operation steps with blank
lines. Tests live in tests/application/management, with strict contract doubles
in tests/application/shared.

Client origin administration stays in client/usecase.go and uses the injected
ClientOrigin repository. Require the separate client_origin permission catalog
constants immediately below each method tag. Verify live space/client ownership,
and check each exact origin belongs to that client before exposing or mutating
it. Filter requests always use the method's client ID. Inactive targets remain
manageable. Validate supplied origins without normalization; derive audit actors
from the caller. Log committed creates/deletes and updates supplying Origin or
IsActive; audit-only updates stay quiet. Use distinct explicit failure logs for
each operation step and its outer transaction.

Client.AuthRedirectUri is configured through existing client create/update permissions.
Validate the exact callback URL and reject simultaneous replacement/clearing.
Committed redirect replacement/clearing emits its own specific info log; no raw
client secret or redirect value is included in metadata. Client creation logs the
client ID, never the ClientSecretResult containing its one-time raw secret.
