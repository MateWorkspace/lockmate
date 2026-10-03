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
    domain::contracts::{repository::Permission, utility::Logger},
    infrastructure::{repository::PostgresPermission, utility::transactor::MatePgdtTransactor},
};
use mate_pgdt::{Pgdt, sqlx::PgPool};

fn dependencies(pool: PgPool, logger: Arc<dyn Logger>) {
    let database = Pgdt::new(pool);
    let permissions: Arc<dyn Permission> =
        Arc::new(PostgresPermission::new(database.clone(), logger.clone()));
    let transactor = MatePgdtTransactor::new(database.transactor(), logger);
}
```

Repository methods consume their streams before returning, so transaction locks
are released before the next operation. Filter totals and page contents use
separate statements, matching Nadi; they do not promise a shared snapshot.
Pagination is one-based, nonpositive limits return no rows, and overflowing
offsets return `BadArgs`. Search retains SQL ILIKE wildcard behavior.

Default-role replacement acquires a transaction-level advisory lock before
clearing the previous default and applying the new write. It reuses any supplied
transaction, whose owner remains responsible for committing or rolling back.
Propagate failed repository operations out of transaction callbacks when their
changes must be rolled back.

Failed public methods log once with safe messages and driver metadata, without
SQL, parameters, password hashes, or API key hashes. Classified unit error
variants retain their established shape; `Failure` and `Timeout` preserve the
native SQLx cause.

Integration tests use the actual migrations in isolated schemas. Against a
throwaway PostgreSQL database with pg_trgm available, run:

```sh
LOCKMATE_TEST_DATABASE_URL='postgres://postgres:password@localhost/test_db' \
    cargo test -- --include-ignored
```

The ignored database tests fail if the URL is absent. Normal `cargo test` runs
all tests that do not require a database, including documentation examples.
