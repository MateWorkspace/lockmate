# Domain models

## Purpose

Own framework-independent data and typed errors used by contracts and usecases.
Keep models independent of application, infrastructure, presentation, SQLx,
mate-pgdt, Redis, and transport bindings. Existing serde, JSON, time, and UUID
value types are supported; do not put I/O, configuration loading, or logging here.

## Layout

- `app/`: AppContext, opaque AppTransaction, AppEnv, and AppInfo.
- `entity/`: persisted entities; shared audit types in `audit.rs`; access and
  refresh claims together in `claims.rs`; GeneratedApiKey in `api_key.rs`.
- `error/`: repository, utility, caching, and usecase error enums, each in its
  own file with a stable `code()` method.
- `logger/`: format, level, and metadata types in separate files.
- `caching/`: queries, families/revisions, stamps/results, pages, and safe cache
  projections and joined payloads.
- `seeder/`: typed embedded-definition inputs and their space-scoped graph.
- Maintain explicit modules and public reexports through each `mod.rs`.

## Code style

Use one file per entity or cohesive model, preserving the combined audit and
claims files. Match adjacent derives and field ordering. Public structs expose
plain typed fields; add rustdoc for ownership, scope, and non-obvious semantics.
Use snake_case fields, i64 entity/audit IDs, and UUID trace identifiers.

Keep stored entities distinct from safe cache and usecase views. User contains
password_hash and ApiKey contains hash; safe projections must discard them,
including inside joined responses. Never add Debug to credential-bearing or
raw-token/raw-key types. Preserve deliberate Debug omissions on cache selectors
and stamps. Serde audit fields flatten and timestamps use RFC3339 as established.

Keep display messages and error codes safe and stable. Preserve native causes
only through source fields on operational errors. Use thiserror and typed
conversions; never expose database detail or credentials through Display.
UsecaseError adds authorization classification and wraps dependency errors;
CachingError stays separate because cache failures are best effort.

## Model invariants

Users are global. Spaces own independent role and permission catalogs, each with
slug plus display name. Memberships connect users to spaces; member_role assigns
multiple roles, and role_permission assigns permissions. API keys belong to a
space membership. Core entities are soft-deleted; assignments are physically
deleted. Preserve audit fields and immutable slug semantics.

AccessClaims and RefreshClaims identify one user, space, and membership. Access
claims include role and permission slugs; refresh claims do not carry grants.
Cryptographic validation and live authorization belong outside these structs.

AppContext.actor is an optional logging label; trace_id is optional UUID. The
transaction field holds an opaque AppTransaction backed by std::any, keeping
concrete driver handles in infrastructure. Clone/equality preserve shared handle
identity; Debug must not reveal the handle. Keep the TransactionContext bridge
in the transactor infrastructure rather than importing its crate here.

Use Option for optional data. Nullable update inputs in contracts/usecases must
retain unchanged/set/clear states; do not collapse them into strings. Avatar paths
are optional object-storage keys; no storage operation is implied by the model.
AppEnv is typed runtime configuration: parsing/defaults live in config. Redis
port is u16, database index u8, timeouts and cache TTLs are Duration. AppEnv does
not derive Debug because it contains secrets. AppInfo holds package metadata.

CachedUser and CachedApiKey exclude hashes. CacheRead distinguishes hits from
misses and retains the lookup stamp; CacheStoreOutcome::Superseded is expected
concurrency behavior. Keep CacheFamily order stable for canonical revision keys.
See `../contracts/AGENTS.md` for dependencies and invalidation requirements.

Seed definitions reference users and spaces by username/slug and nest independent
role/permission catalogs. Graph validation is in config; field validation uses
the utility contract. Do not make models execute seeding or read environment files.

## Verification

Run `cargo fmt --check`, `cargo check --all-targets`, and `cargo test` from the
Lockmate root. Model/contract checks live under tests/domain; seed consistency
checks live under tests/config. Align model changes with repository scanners,
migrations, ERD, cache payloads, claims, and validators where relevant.
