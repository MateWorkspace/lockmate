# Redis management caches

Each entity implements its caching trait in its own `redis.rs`. A shared
`RedisBackend` owns one clonable asynchronous connection manager. Key building
and invalidation have separate adapters; private helpers supply canonical query
metadata, payload validation, JSON encoding, revision scripts, and error mapping.

```rust,no_run
use std::sync::Arc;
use lockmate::{
    domain::{contracts::{caching::{Space, KeyBuilder}, utility::Logger}, models::{AppContext, AppEnv, CachingError}},
    infrastructure::caching::{RedisBackend, RedisSpace, RedisInvalidation, Sha256KeyBuilder},
};

async fn caches(env: &AppEnv, logger: Arc<dyn Logger>, context: &AppContext) -> Result<(), CachingError> {
    let keys: Arc<dyn KeyBuilder> = Arc::new(Sha256KeyBuilder::new(&env.redis_namespace)?);
    let backend = RedisBackend::connect(context, env, keys, logger).await?;
    let spaces: Arc<dyn Space> = Arc::new(RedisSpace::new(backend.clone()));
    let invalidation = RedisInvalidation::new(backend);
    let cached = spaces.read_by_id(context, 1).await?;
    // On a miss, application code reads PostgreSQL and stores the successful
    // result using cached.stamp and env.caching_record_ttl.
    Ok(())
}
```

Use standalone Redis with the configured host, port, database index, and optional
password. Connection information is built structurally, without a password URL.
`LOCKMATE_REDIS_CONNECT_TIMEOUT` defaults to 5s;
`LOCKMATE_REDIS_OPERATION_TIMEOUT` defaults to 200ms and bounds the whole public
operation, including reconnection waits. TTL defaults live in AppEnv:
`LOCKMATE_CACHING_RECORD_TTL=60s`, `LOCKMATE_CACHING_LIST_TTL=15s`.
Positive fractional TTL milliseconds round upward; invalid ranges are rejected.
Initial handshake failures return promptly without connection-attempt retries;
the manager still reconnects after connection loss.

Reads initialize missing revision metadata with fresh random 256-bit tokens in
one script, then conditionally fetch an entry against that snapshot in another.
A generation change between scripts yields a miss; its stamp is safely rejected
by a later store. Stores atomically compare all revisions and SET with PX.
Invalidation preflights all revision types/values before replacing revisions in
one script. Revisions do not expire; old payloads expire without key scans.

Malformed revision metadata or cached payloads return Failure rather than a hit.
No raw source errors, keys, selectors, payloads, passwords, or connection settings
are logged. Public Redis failures log once with request context and static code:
operational failures warn, invalid arguments/state error. Helpers and key builder
remain silent. No routine success or miss logs are emitted.

All contracts reject transaction contexts. Application code must bypass cache
inside transactions, collect changed families, and invalidate after the outer
commit; rollback performs no invalidation. PostgreSQL fallback, authentication
bypass, and transaction orchestration are not implemented here. The full
management dependency rules are documented by the domain caching contracts.

Deadlines do not guarantee cancellation of a server-side command. Failed
invalidation and the gap after PostgreSQL commit can permit stale management
reads until TTL expiry. No blind command retries or cross-system atomicity are
provided. Redis Cluster, TLS, ACL usernames, stampede locks, negative caching,
durable invalidation retries, and authentication caching are deferred.

Redis integration tests use unique namespaces and track keys for cleanup. Against
an isolated Redis instance, run:

```sh
LOCKMATE_TEST_REDIS_URL='redis://127.0.0.1:6379/0' cargo test infrastructure::caching -- --include-ignored
```

No FLUSHDB is used against the supplied service.
