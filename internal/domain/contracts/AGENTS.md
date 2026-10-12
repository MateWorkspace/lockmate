# Domain contracts

Define the capabilities that application code depends on. Contracts describe
operations and their inputs and outputs; infrastructure supplies implementations.
Keep database drivers, SQL builders, cryptography implementations, and logging
frameworks outside this scope.

## Layout

- `repository/`: persistence interfaces for space, permission, role, user,
  classic auth, Google auth, client, client origin, space member, role permission, and space
  member role. Keep one contract file per entity.
- `utility/`: tracer, logger, keygen, password, transactor, token, validator,
  permission guard, and Google OAuth interfaces. Keep one file per capability.

Use `domaincontractsrepository` and `domaincontractsutility` as package names.
Import domain packages with the existing explicit aliases. Dependencies belong
to the standard library and domain models; do not expose pgx, Squirrel, slog,
bcrypt, or JWT library types through these interfaces.

## Declaration style

Place the interface first, followed by its argument, filter, summary, and joined
result structs. Use `Create<Entity>Args`, `Update<Entity>Args`, and `<Entity>Filter`
names. Group related fields into an argument struct when a signature becomes
hard to read. Keep identity and space scope explicit in repository signatures.

Follow the existing `Id`, `UserId`, and `SpaceId` spelling. Separate declarations
and logical groups with blank lines. Document behavior that a signature cannot
express, especially null updates, bulk selectors, and transaction expectations.

Use `context.Context` first for repositories, logger, tracer, transactor, and
Google OAuth.
The current value-oriented utilities do not receive context; preserve their
signatures unless a concrete requirement calls for changing the contract.

## Repository behavior

Create operations return the new `int64` ID and an error. Exact reads return a
pointer and an error. Filter reads return a slice, total count, and an error.
Updates and deletes return an error. Missing exact records and zero-row single
mutations return the corresponding domain not-found error; empty lists return
an allocated empty slice. Bulk deletion may succeed without matching rows.

Require space IDs for scoped operations. Users and authentication identities are
global; client slug lookup is global even though clients belong to spaces.
Ordinary operations exclude deleted records and deleted parents. Effective grant
reads also require active spaces and memberships and return deduplicated grants.

Keep credential hashes out of list results through `ClassicAuthSummary` and
`ClientSummary`. Use explicit joined result structs rather than parallel slices.
Exact authentication reads may return hashes needed by trusted application code.

Update structs must not allow changing ownership, stable slugs, or Google
subjects. A pointer represents an optional replacement; a corresponding
`Set...Null` flag explicitly clears a nullable column. Supplying both is invalid.
Client secret hashes and redactions are replaced together. Contact changes reset
verification unless verification is explicitly supplied.

Filters use `int64` page and limit values. Pages at or below one select the first
page; nonpositive limits return no rows while preserving the total. Reject
overflow. Bulk assignment deletion requires at least one selector; two selectors
use OR within the requested space.

## Utility behavior and errors

`Transactor.WithTx` passes the transaction-aware context to its callback. Every
repository call within that callback must receive that context. Propagate failed
operations so the transaction can roll back. Feature transactors reject an
attached transaction; repository helpers reuse the callback transaction without
savepoints. The infrastructure adapter classifies nesting as BadState.

Token contracts use domain claims, and permission checks receive authenticated
access claims. Validation and credential generation do not perform persistence.
Logger calls receive an explicit tag; tracer injection and extraction carry the
domain trace through context.

Tracer also generates fresh event UUIDs through GenerateEventId. The misc
application usecase accepts only external client slug/execution UUID and assembles
the trace internally; its Inject returns context and error.

GoogleOAuth owns provider verification and transient authorization attempts,
including single-use state and browser/session binding. Return only verified
identity and the original client/space association; provider SDK types, tokens,
and configuration remain outside the contract. Its adapter is implemented through OAuth2 and OIDC libraries.

Return domain errors from `models/error`, preserving native causes where useful.
Callers classify errors with `errors.Is` and inspect structured errors with
`errors.As`. Contracts must not prescribe infrastructure logging or expose
credentials in error messages.

When changing a signature, align implementations, callers, and tests together.
Behavior tests belong in `tests/` under the corresponding layer and capability.

ClientOrigin derives space ownership through its client. Every management operation
is space-scoped. Origins must already be canonical and validated through
Validator.ClientOrigin before persistence. Ordinary reads include inactive
records; IsAllowedOriginForClient and the global IsAllowedOrigin require active,
nondeleted origins, clients, and spaces. A global CORS match never authenticates
or identifies a client. Client/origin uniqueness includes inactive records but
excludes deleted records.

Client create/update and safe projections include nullable AuthRedirectUri.
SetAuthRedirectUriNull clears it; clearing and replacement are mutually exclusive.
AuthFlow generates opaque URL-safe secrets and hashes. GoogleOAuth.Begin returns
both state and authorization URL; Complete returns validated scope/redirect and
provider identity, or a validated denial result. Transient launch/attempt/handoff
repositories expose Create, atomic Consume, and bounded PurgeExpired. Handoff
consumption follows the caller's transaction and is committed only with tokens.
