# PostgreSQL repositories

Each entity keeps execution and its contract implementation in `postgres.rs`.
Its `postgres_query.rs` builds native SQLx queries and bindings, without database
access or logging. Shared helpers decode rows, classify errors, and normalize
pagination. The domain contains no SQLx types.

Create one `mate_pgdt::Pgdt` from the pool and pass clones to every repository.
Derive the application transactor from that same wrapper. Rewrapping the pool
creates a different owner and transaction handles will be rejected.

```rust,no_run
use std::sync::Arc;
use lockmate::{
    domain::{contracts::{repository::Permission, utility::Logger}, models::AppContext},
    infrastructure::{repository::PostgresPermission, utility::transactor::MatePgdtTransactor},
};
use mate_pgdt::{Pgdt, sqlx::PgPool};

async fn dependencies(pool: PgPool, logger: Arc<dyn Logger>, context: &AppContext, space_id: i64) {
    let database = Pgdt::new(pool);
    let permissions: Arc<dyn Permission> =
        Arc::new(PostgresPermission::new(database.clone(), logger.clone()));
    let transactor = MatePgdtTransactor::new(database.transactor(), logger);
    let permission = permissions.read_by_slug(context, space_id, "settings.read").await;
}
```

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
    cargo test -- --include-ignored
```

The ignored database tests fail if the URL is absent. Normal `cargo test` runs
all tests that do not require a database, including documentation examples.
JWT claims and embedded seed definitions use the space-aware model. Application
authentication, authorization, and database seeding execution remain separate
work; token validation alone does not query live membership state.
