# Domain contracts

## Purpose

Define application dependencies independently of their implementations. Keep this
folder limited to traits, dependency inputs/results, and future/callback aliases.
Import domain models and necessary data types, never infrastructure, application,
presentation, SQLx, mate-pgdt, Redis, JWT driver, or transport bindings.

## Layout

- `repository/`: one file per persisted entity; `future.rs` holds RepositoryFuture.
- `caching/`: matching entity traits, plus key building, invalidation, and futures.
- `utility/`: logger, password, token, API-key, validator, and transactor traits.
- Each `mod.rs` declares modules and explicitly reexports their public contracts
  and associated types. Keep the three dependency groups separate.

## Code style

- Match neighboring explicit trait signatures. Traits are `Send + Sync` and
  object safe; use named lifetimes and boxed `Send` futures for asynchronous I/O.
- Pass `&AppContext` first. Space-owned repository and cache operations take an
  explicit positive `space_id`; global User and Space operations have no scope ID.
- Place input, filter, and result structs after the trait's method list. Preserve
  field order because canonical caching selectors depend on it.
- Use repository verbs `create`, `read_by_id`, `read_by_<selector>`,
  `read_by_filter`, `update_by_id`, and `delete_by_id`. Keep assignment joins typed.
- Preserve optional update states with `Option<Option<T>>` for nullable fields.
  Keep audit actor inputs numeric and separate from AppContext's logging actor.
- Keep synchronous utility operations synchronous. Logger receives the tag as an
  argument; Transactor callbacks receive an owned transactional AppContext and
  return TransactionFuture with the domain TransactorError.
- Use typed domain errors, not driver errors or arbitrary string errors. Retain
  credential-bearing inputs without Debug; never introduce auth cache payloads.
- Introduce a dependency contract before implementing a usecase that needs it.
  Contract changes must align adapters, mocks, models, and exports together.

## Caching behavior

These contracts define cache-aside behavior independently of Redis. They do not
change repository behavior. Redis adapters implement these interfaces separately;
application orchestration remains separate work.

### Read and store

Read the management cache outside a transaction. A hit supplies a typed value;
a miss supplies `None`. Both supply a stamp identifying the lookup and dependency
revisions. On a miss, read PostgreSQL and store a successful result using that
same stamp. Never modify stamps or reuse them for another lookup.

Use `AppEnv.caching_record_ttl` for records and details, and
`AppEnv.caching_list_ttl` for pages and assignment lists. Configure these through
`LOCKMATE_CACHING_RECORD_TTL` (default 60 seconds) and
`LOCKMATE_CACHING_LIST_TTL` (default 15 seconds). TTL must be
positive. Empty lists and pages are successful cacheable results. Record not-found
and repository failures are not cached. Page items and total are one payload,
without strengthening the repository's database snapshot guarantees.

Stores validate that their stamp identifies a supported lookup for that method,
that all required revisions are present exactly once, and that the payload's
scope and lookup fields match. For example, `store_record` for a role supports
ID, slug, and default lookups; a default lookup requires `is_default = true`.
Assignment stores validate the selected member, role, or permission and space.
Wrong lookup shapes, invalid IDs, or mismatching payloads return `BadArgs`.

Adapters atomically compare current dependency revisions and store the payload.
If revisions changed, return `Superseded` and leave the cache unchanged; callers
can still return the database result. This prevents a delayed database read from
refilling an invalidated generation. Cache hits must only use an entry matching
the revisions observed by that cache read.

### Safe management payloads

`CachedUser` omits `password_hash`; `CachedApiKey` omits `hash`. Membership pages
use `CachedSpaceMemberWithUser`. Explicit conversions consume repository entities
and discard credential fields. Other joins contain only safe entity fields.
Serialization is supported by the cache payload models; repository DTOs are
unchanged. These projections are not authentication models.

Password verification, authentication/refresh identity reads, API-key hash
lookups, and effective role/permission checks bypass the management cache and
read PostgreSQL. Never cache passwords, credential hashes, raw API keys, or JWTs.
Cache hits do not authorize callers; application code still enforces access.

### Invalidation dependencies

Families are global users and spaces, or an entity family within a positive
`space_id`. Replace every requested family's revision atomically with a fresh,
non-reusable opaque token. Duplicate families are harmless; an empty request is
a no-op. Revision metadata must not expire. If metadata is absent, establish a
fresh token rather than assuming a fixed generation, including after cache loss.
Old entries become unreachable and expire with their own TTL; no wildcard scans
or key enumeration are required.

| Cached family | Revision dependencies |
| --- | --- |
| Users | Users |
| Spaces | Spaces |
| Permissions | Spaces, Permissions(space) |
| Roles, including default | Spaces, Roles(space) |
| Space members | Spaces, Users, SpaceMembers(space), MemberRoles(space), Roles(space) |
| Role-permission assignments | Spaces, RolePermissions(space), Roles(space), Permissions(space) |
| Member-role assignments | Spaces, Users, MemberRoles(space), SpaceMembers(space), Roles(space) |
| API-key management | Spaces, Users, ApiKeys(space), SpaceMembers(space) |

Invalidate the mutated entity family after each committed create, update, or
delete. This invalidates records, alternate lookups, pages, and totals together.
Updates can change filters and ordering, so they invalidate lists too.
Membership creation also invalidates member-role assignments because it assigns
a default role. Default-role replacement invalidates the whole role family,
including the previously default role. Bulk assignment deletion invalidates its
family even when no rows were removed. Global user and space revisions invalidate
joined reads across spaces without discovering affected spaces.

### Transactions and failures

Application code bypasses caching whenever `AppContext.transaction` is present.
Cache and key contracts reject transactional contexts with `BadState`, rather
than reading or publishing uncommitted state. Accumulate mutated families during
a transaction and invalidate once after the outermost successful commit. Nested
callbacks do not invalidate early; rollback does not invalidate. This is an
orchestration requirement, not a change to the existing transactor contract.

`CachingError::BadArgs` denotes invalid inputs; `BadState` denotes a transactional
context. `Timeout` retains deadline failures; `Failure` retains connectivity,
encoding/decoding, malformed cached payloads, and other operational causes.
Display messages and codes are safe and static; do not expose underlying causes
to users. No conversion from `RepositoryError` is provided.

Cache read failures fall back to PostgreSQL. Corrupt payloads are failures,
never hits; an adapter may remove them. Store and invalidation failures do not
turn committed database success into a user-facing failure. Future application
orchestration falls back without duplicating adapter logs. Adapters log safe
warnings without keys, selectors, payloads, or credentials.

PostgreSQL commit and cache invalidation are separate operations. Revisions
prevent stale fills after successful invalidation, but do not eliminate the gap
between those operations or failed invalidation. TTL bounds stale entries.
Direct database writes are reflected at expiry unless they also invalidate.

### Key builder

The separate `KeyBuilder` contract performs no I/O. Its infrastructure constructor
receives `redis_namespace`, used as the first key segment. Namespace must be
nonempty and contain no colon, so it remains a single segment.

Entry keys use:

`<namespace>:cache:v1:<scope>:<entity>:<operation>:<selector-digest>:<revision-digest>`

Revision keys use:

`<namespace>:cache:v1:revision:<scope>:<entity>`

Scope is `global` or `space:<positive-id>`. Entity segments are `user`, `space`,
`permission`, `role`, `space_member`, `role_permission`, `member_role`, and
`api_key`. Operation segments correspond to repository verbs such as
`read_by_id`, `read_by_filter`, and `read_default`; pair lookups use their complete
repository method name. Each `CacheQuery` selects exactly one operation.

Use lowercase hexadecimal SHA-256 digests. Selector input is a UTF-8 JSON array
of lookup values in the order declared by the query variant. For filters, expand
the filter into an array of all fields in its repository struct declaration
order. Encode integers as JSON numbers, strings as strings, and options as null
or their contained value. Preserve search strings exactly; do not add trimming,
case folding, or pagination normalization. A default-role query still includes
its space ID. Namespace, scope, entity, and operation distinguish entry families.

Revision input is a JSON array of `[family, space_id_or_null, token]` arrays,
sorted by `CacheFamily`'s derived order. Family strings use the entity segments
above. Reject missing, duplicate, extra dependencies, empty tokens, and invalid
IDs with `BadArgs`. Hashing selectors prevents raw contacts and search values
from appearing in keys; it is not encryption. Actor and trace ID never affect
cache identity. Key and stamp specifications deliberately omit `Debug`.

### Adapter acceptance scenarios

Adapter changes must test hit, miss, expiry, successful empty results, invalid
payloads, positive TTL enforcement, and failed database reads not being stored.
Test space/filter/optional-value key isolation, every mutation dependency,
default-role replacement, bulk assignment deletion, and membership default-role
assignment. Exercise stale-fill rejection, revision loss, corrupt payloads,
outer commit versus rollback, and Redis outage fallback. Authentication paths
must never consume management cache values.

Stampede locks, negative caching, durable invalidation retries, and authentication
caching are deferred.

## Verification

Run checks from the Lockmate root: `cargo fmt --check`,
`cargo check --all-targets`, and `cargo test`. Contract tests live under
`tests/domain/`; adapter tests live under `tests/infrastructure/`.
