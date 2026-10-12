# PostgreSQL repositories

Implement the domain repository contracts using pgx v5, Squirrel, and the shared
`mate-golang/pgxdt` execution wrapper. There is no ORM, database contract, logging,
credential hashing, or permission checking in this layer.

## Layout and style

Each entity folder contains `postgres.go` for its constructor, execution, and
transaction orchestration, and `postgres_query.go` for fixed column lists and
query builders. `shared/` contains BasePostgres, pagination, parent/default-role
locks, explicit scanners, and PostgreSQL error classification.

The entity folders are `space`, `permission`, `role`, `user`, `classic_auth`,
`google_auth`, `client`, `client_ip_allowlist`, `space_member`, `role_permission`,
`space_member_role`, and `signing_key`. Use the existing concatenated
`infrastructurerepository<entity>` package names and private `postgresImpl`
structs embedding `shared.BasePostgres`.

Constructors return domain interfaces and receive the same pgxdt wrapper plus
question and dollar Squirrel builders. Use the existing explicit dependency and
import aliases. Bind values, keep table names and sorting adapter-owned, and
match scanner destinations to the selected column order. Keep code spaced and
use early returns. Close rows and check their final error.

Execute all queries through the injected pgxdt wrapper and pass the caller's
context. Keep reusable query construction in `postgres_query.go`; keep scanning,
error conversion, and multi-step orchestration in `postgres.go`. Shared helpers
must have a concrete purpose across repositories rather than hide entity logic.

Exact reads and single mutations return entity-specific not-found errors when
no row matches. List results use allocated empty slices. Count and page queries
share the same predicates; nonpositive limits retain the count with no rows.
Use deterministic creation-time/ID ordering for ordinary lists and slug/ID
ordering for effective grants. Keep client and classic-auth list projections
free of secret-key and password hashes.

## Scope, concurrency, and errors

Constrain every scoped query by space ID. Reads exclude soft-deleted records and
parents, and writes verify live ownership. Assignment insertion locks endpoints
and checks matching spaces. Safe list queries omit hash columns from SQL itself.

User-owned creation locks the user before the space, when applicable. Parent
locks use FOR SHARE; user deletion locks the user FOR UPDATE before modifying
dependents. Default-role operations lock their live space before obtaining the
shared per-space advisory transaction lock. Membership creation uses that same
lock before selecting and assigning the default role. Preserve this ordering.

Default-role replacement clears the previous default atomically. Membership
creation inserts its default-role assignment in the same transaction and fails
when no live default role exists. User deletion atomically soft-deletes owned
authentication records and memberships and removes membership-role assignments,
using the same audit time and actor. Assignment deletion is physical and permits
cleanup of assignments whose endpoints have been deleted.

Internal transactions reuse the caller context through pgxdt. Never swallow a
failed nested operation: the caller must return the error so the outer transaction
rolls back. Single reads and count/page pairs do not create snapshot transactions.
Field validation and inactive-record access decisions remain in application;
effective-grant reads explicitly reject inactive spaces or memberships.

Use exact unique-index names for conflict classification. Preserve existing domain
errors and native causes. Map cancellation, deadlines, malformed values, missing
records, and foreign-key/unique violations consistently. Do not emit logs or
include credential values in error messages.

Use explicit null-update flags consistently, rejecting a replacement together
with a clear flag. Keep credential generation and validation in utilities:
repositories store supplied hashes. Update client hashes and redactions as a
pair, and reset verification for changed or cleared contacts unless the caller
explicitly supplies the verification value.

## Verification

Tests live in `tests/infrastructure/repository`. Recording executors cover bound
inputs, projections, row cleanup, rejected arguments, and error mapping. Database
tests require `LOCKMATE_TEST_DATABASE_URL` for an isolated PostgreSQL database;
they create and drop unique schemas and ensure pg_trgm is available in public.
Only canonical timestamped `.up.sql` files are loaded; unrelated copy files are
ignored. Do not mutate a developer's database for validation.

Run `go test ./...`, `go vet ./...`, and `go build ./...` from the Lockmate root.
Format changed Go files with gofmt. PostgreSQL tests skip without the test URL.

Client-IP-allowlist writes lock the live space before the live, space-owned client.
Management reads include inactive records but hide deleted parents. Allowlist
checks require the matching entry, client, and space to be active and nondeleted;
false is an ordinary nonmatch, whereas database failures remain errors. Match the
resolved IP with Postgres CIDR containment. Preserve child entries when a parent
is soft-deleted. Client-IP-allowlist uniqueness is per client among all nondeleted records.

Internal google_oauth_launch, google_oauth_attempt, and auth_handoff adapters keep
query building in postgres_query.go. Use atomic DELETE ... RETURNING with digest
and ownership/binding predicates. Launches/attempts require nonexpired rows;
handoff expiry is classified by application inside the exchange transaction.
Bound PurgeExpired to 1–1000 rows. Respect the pgxdt transaction context so failed
exchanges roll back consumption. No general management endpoints expose records.

The signing_key adapter stores sealed private material and never performs
cryptography. Reads exclude soft-deleted rows; `ReadActive` and `DemoteActive`
target live active rows; `ListVerifiable` returns active and retiring keys in
creation order; `ListStale` returns live rows whose fingerprint differs;
`DeleteRetiringBefore` soft-deletes retiring keys past the cutoff. Map the
`uq_signing_keys_kid` and `uq_signing_keys_active` partial unique indexes to
SigningKeyConflict. Return SigningKeyNotFound for missing exact rows.
