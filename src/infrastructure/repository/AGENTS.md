# PostgreSQL repositories

## Purpose

Implement domain repository contracts using PostgreSQL.
Keep business orchestration and authorization in application; keep driver types
out of domain signatures. Preserve the existing behavior described below.

## Layout

- One directory per entity, with `mod.rs`, `postgres.rs`, and `postgres_query.rs`.
- `shared/`: query/pagination helpers, driver error classification, and per-entity
  scanners under `scan/`.
- Root `mod.rs`: explicit exports of concrete Postgres adapters.

## Code style

Implement the domain trait in postgres.rs; keep SQL construction in
postgres_query.rs with bound SQLx QueryBuilder arguments. Query builders do no
I/O or logging. Use explicit implementations rather than production macros or a
generic repository abstraction. Alias colliding entity/contract names clearly.

Pass AppContext unchanged except for supplied transaction callbacks. Inject
clones of one Pgdt owner and Arc<dyn Logger> through constructors. Map driver
errors through shared helpers, and consume query streams before returning. Keep
blank lines between query construction, execution, decoding, and error handling.
Update scanner column order with model/query changes; preserve typed joined DTOs.

## Behavior and verification

Each entity keeps execution and its contract implementation in `postgres.rs`.
Its `postgres_query.rs` builds native SQLx queries and bindings, without database
access or logging. Shared helpers decode rows, classify errors, and normalize
pagination. The domain contains no SQLx types.

Create one `mate_pgdt::Pgdt` from the pool and pass clones to every repository.
Derive the application transactor from that same wrapper. Rewrapping the pool
creates a different owner and transaction handles will be rejected.

Users have global identities. Spaces own roles and permissions, and space
memberships receive roles through member-role assignments. Slugs are immutable;
display names can change. API keys belong to a membership and one space.

Every space-owned operation receives `space_id` after `AppContext`. ID lookups,
mutations, filters, hash lookups, and assignments all use that scope. Guarded
relationship inserts reject missing, deleted, or cross-space parents with
`BadArgs`. Single-record lookups and mutations with a wrong-space ID return the
entity's not-found error. Empty list reads succeed. Callers still authorize who
may operate within the supplied space.

Management reads include disabled spaces and suspended memberships, while
excluding deleted records and deleted parents. Effective member-role and
permission reads require an active membership and active space with a live user;
unavailable memberships return `SpaceMemberNotFound`, inactive access returns
`BadState`. Effective permissions are deduplicated and ordered by slug then ID.
`read_active_by_hash` returns the key, membership, user, and space only when all
access conditions hold; otherwise it returns `ApiKeyNotFound`. Ordinary API-key
hash reads remain available for management.

Default-role replacement acquires a transaction-level advisory lock scoped to
one space before clearing the previous default and applying the new write.
Creating a membership uses the same lock and atomically assigns the current
default role. It returns `RoleNotFound` without inserting a membership if no
default exists. Inactive memberships also receive the default assignment.
Changing the default does not rewrite existing assignments.

Both operations reuse a supplied transaction, whose owner remains responsible
for committing or rolling back. Propagate failed repository operations out of
transaction callbacks when their changes must be rolled back; nested operations
have no savepoints. Internal helper queries do not call other public repository
methods, so a failure produces one repository log.

Soft deletion preserves child records, which become unavailable through joined
reads. Recreating a membership creates a new ID and does not restore old keys or
assignments. Assignment deletions are physical and permit cleanup even when
parents were soft-deleted. Bulk assignment deletion uses OR when both IDs are
provided, requires at least one ID, and stays within the supplied space.

Repository methods consume their streams before returning, so transaction locks
are released before the next operation. Filter totals and page contents use
separate statements, matching Nadi; they do not promise a shared snapshot.
Pagination is one-based, nonpositive limits return no rows, and overflowing
offsets return `BadArgs`. Search retains SQL ILIKE wildcard behavior.

Failed public methods log once with safe messages and driver metadata, without
SQL, parameters, password hashes, or API key hashes. Classified unit error
variants retain their established shape; `Failure` and `Timeout` preserve the
native SQLx cause.

Integration tests apply all eight migrations in order in isolated schemas.
Against a throwaway PostgreSQL database with pg_trgm available in public, run:

```sh
LOCKMATE_TEST_DATABASE_URL='postgres://postgres:password@localhost/test_db' \
    cargo test infrastructure::repository -- --include-ignored
```

The ignored database tests fail if the URL is absent. Normal `cargo test` runs
all tests that do not require a database, including pure contract/key checks.
JWT claims and embedded seed definitions use the space-aware model. Application
authentication, authorization, and database seeding execution remain separate
work; token validation alone does not query live membership state.

Run `cargo fmt --check`, `cargo check --all-targets`, and normal `cargo test`
from the Lockmate root before relevant isolated integration checks.
