# Infrastructure utilities

## Purpose

Implement domain utility contracts with concrete libraries. Accept domain models
and injected contracts at the boundary; keep dependency-specific types here.
Application code selects behavior through contracts, and composition wires
adapters. Do not put business orchestration or presentation mapping in utilities.

## Layout

| Folder | Implementation files |
| --- | --- |
| logger | json.rs and plain.rs each implement Logger; utils.rs shares record/writer helpers |
| password | bcrypt.rs implements Password |
| token | jwt.rs implements Token |
| api_key | generator.rs implements ApiKey |
| validator | regex.rs implements Validator; utils.rs shares field helpers |
| transactor | mate_pgdt.rs implements Transactor; mate_pgdt_txctx_impl.rs bridges AppContext to TransactionContext |

Keep declarations and explicit concrete-adapter exports in each mod.rs. Use one
implementation file per backend and put shared private helpers alongside it;
do not move both logger trait implementations into a generic shared backend.

## Code style

Match adjacent constructors, explicit fields, import aliases, and early-return
error handling. Inject Arc<dyn Logger> where needed rather than global logging.
All contract operations receive AppContext first. Keep synchronous operations
synchronous and asynchronous transactor work boxed with a Send future.
Use module/operation tags such as `utility/password/bcrypt/hash`, lowercase log
messages, and domain LoggerMeta. Preserve readable blank lines between validation,
work, mapping, and logging. Avoid production macros and unnecessary abstraction.

Return the matching typed domain error and preserve native operational causes.
Public adapter failures log once; do not duplicate helper or callback logs.
Follow each adapter's established metadata policy. Never log passwords, hashes,
raw API keys, tokens, JWT secrets, or database parameters; retain causes for
internal diagnosis rather than exposing them in user messages.

## Logger

JsonLogger and PlainLogger independently implement Logger. The contract receives
level, tag, message, and metadata explicitly; no tracing framework is involved.
Print supplied arguments without recursive error-source expansion. Missing actor
prints UNKNOWN; missing trace prints the zero UUID; empty metadata prints {}.
Keep empty supplied actors distinct from absent actors. Preserve level filtering,
plain-output escaping, and synchronized writes to avoid interleaved records.
Shared helpers live in logger/utils.rs; output failures do not panic or recursively
attempt to log themselves.

## Passwords and API keys

BcryptPassword accepts costs 4 through 31 and falls back to 10 for invalid costs.
Reject password inputs over 72 UTF-8 bytes rather than silently truncating.
Mismatch differs from malformed-hash or operational comparison failures. User
password policy validation remains the Validator's responsibility.

ApiKeyGenerator uses `lockmate-` plus 32 random bytes encoded as 64 lowercase hex
characters (72 characters total). Hash the entire raw key with SHA-256 to lowercase
hex and keep its last four characters as redaction. Validate exact prefix, length,
and alphabet. Secure randomness makes collisions extremely unlikely; do not claim
an unconditional uniqueness guarantee from generation alone. Return raw/hash/
redaction only as GeneratedApiKey for storage orchestration; safe usecase views
exclude the hash and expose the raw key only at creation.

## Validation

Compile regexes in RegexValidator construction. Keep field-specific error enums
and length/format precedence. Names/slugs have lengths 3–100, usernames 3–32,
text fields at most 1,000, emails 3–254, and phone digits 8–15 with optional +.
Passwords require at least eight characters and at most 72 UTF-8 bytes, without
mandatory character classes. Follow the existing Unicode display-name rules,
ASCII username/email/phone rules, and entity-specific slug patterns: space and role slugs use lowercase letters,
digits, and separated hyphens; permission slugs additionally allow `_`, `.`,
`:`, and `-` after their initial lowercase letter.
Do not log submitted values. Distinguish optional text validation from absence;
application code decides when to validate optional supplied fields.

## Transactions

Derive MatePgdtTransactor from the same Pgdt instance whose clones repositories
receive. Rewrapping a pool creates a different owner. Keep TransactionContext's
AppContext implementation in mate_pgdt_txctx_impl.rs and concrete handles outside
domain models. Preserve opaque handle identity and foreign/closed-handle checks.

with_tx reuses supplied transactions without savepoints. Pass the callback's
AppContext to every repository call. Nested calls do not commit independently;
the outer owner commits or rolls back. Preserve callback TransactorError unchanged
without logging it again. Map begin/commit/rollback driver failures to Timeout or
Failure, keeping native causes and safe code/SQL-state/constraint metadata. The
private TransactionError distinguishes driver and callback failures; it is not
a new public domain error. Do not add cache invalidation inside this adapter;
application orchestration invalidates only after the outer commit.

## JWT tokens

Access and refresh tokens are bound to one space membership. Both contain
`user_id`, `space_id`, `member_id`, `name`, and `username`. Access claims also
contain `roles` and `permissions`: slugs from that space, rather than display
names. Empty grant lists are valid. Refresh claims contain no grants.

Generation adds a canonical user-ID `sub`, audience `lockmate:space:{space_id}`,
`token_type` (`access` or `refresh`), and integer-second `iat`, `nbf`, and `exp`.
The adapter emits HS256 and accepts the established HS256/HS384/HS512 allowlist.
IDs must be positive. Invalid generation inputs return `BadArgs`; missing secrets,
zero lifetimes, or unusable expiration ranges return `Failure`.

Call `validate_access(context, expected_space_id, token)` or
`validate_refresh(context, expected_space_id, token)` with the caller's expected
space. Validation requires identity, scope, purpose, audience, subject, and time
claims. It checks positive IDs, exact space/audience and token type, canonical
subject consistency, expiration/not-before without clock skew, and
`iat <= nbf < exp` with no future issue time. Each token carries one string
audience. Old unscoped tokens and the previous subject fallback are rejected.

Empty tokens or nonpositive expected-space arguments return `BadArgs`. Expired
tokens return `Expired`; other validation failures return `Invalid`. Access and
refresh purposes are checked independently of signing keys, including when both
configured secrets are equal. Each failed public operation logs once with a safe
message and without tokens, secrets, or claim contents.

The utility validates token contents and does not query repositories. Application
authentication must check live membership state and current privileges when
refreshing or switching spaces. Session orchestration and application
authorization remain separate work. The domain usecase policy requires live
checks for access-token authentication too; see `../../domain/usecases/AGENTS.md`.

## Verification

Run `cargo fmt --check`, `cargo check --all-targets`, and
`cargo test infrastructure::utility` from the Lockmate root. Tests are under
`tests/infrastructure/utility/`. Preserve JSON/plain output, absent-context defaults,
password byte boundaries and Go hash compatibility, key format/hash vectors,
validator field errors, and scoped JWT purpose/audience/time validation.

Transactor integration tests require an isolated PostgreSQL service through
LOCKMATE_TEST_DATABASE_URL. Run `cargo test infrastructure::utility::transactor
-- --include-ignored` with that variable set. Check nested commit/rollback,
cancellation, ownership, callback errors, and safe logging without changing a
developer database.
