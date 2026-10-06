# Domain usecases

## Purpose

These interfaces define application behavior using domain models and dependency
contracts. They do not implement application logic or enforce authorization on
their own. Application implementations must follow the rules below. Repository,
caching, and utility adapters remain independent dependencies selected by
composition; domain usecases never import those adapters or transport types.

## Layout

- `auth/session.rs`: login, refresh, and access-token/API-key authentication.
- `management/`: one trait file for each of the eight repository entities;
  role_permission and member_role assignments have their own files.
- `profile/`: account, security, and own-API-key traits.
- `shared/`: safe views, joined results, Page, and `future.rs` with UsecaseFuture.
- `mod.rs` files declare modules and explicitly reexport traits/shared types.
  Keep local CreateRequest/UpdateRequest names in their feature modules.

## Code style

Follow the existing contract layout: explicit trait methods first, requests and
results afterward. Traits are Send + Sync and object safe, with named lifetimes
and boxed Send futures. Keep method names snake_case and data structs simple.
Add dependencies through domain contracts first; do not reference concrete
adapters. Keep implementation and orchestration in the future application layer.
Use shared safe responses rather than cache-named DTOs or hash-bearing entities.

## Calling conventions and responses

Every operation receives `AppContext` first. Protected operations receive
`AccessClaims` next, from a trusted authentication path. Presentation must not
construct a caller from client-supplied IDs or unsigned claims. Claims do not
carry token expiry; their validation must happen for each incoming request.
`AppContext.actor` is logging metadata, not authorization or a numeric audit ID.
Application code derives repository audit actors from `caller.user_id`.

Traits are object safe and return boxed `Send` futures with `UsecaseError`.
Requests are owned, with selectors and scoped `space_id` passed explicitly as in
repository contracts. Requests and results live after the trait in each file.
Names are local to their modules; for example, `management::user::CreateRequest`.

`UserView` excludes passwords and their hashes. `ApiKeyView` excludes hashes.
Joined results use those safe views. `Page<T>` contains items and the total
matching the repository result, without promising a stronger database snapshot.
`CreatedApiKey` returns metadata and the raw key only at creation. Token results,
password requests, and raw-key results intentionally omit `Debug`; never log
passwords, hashes, raw API keys, or tokens.

Updates preserve omission and clearing: `None` leaves a nullable field unchanged,
`Some(Some(value))` sets it, and `Some(None)` clears it. Slugs are immutable.
User verification resets remain atomic repository behavior. Preference updates
replace the stored JSON value, following repository semantics. Avatar paths are
metadata only; no object-storage operations are implied.

Validate IDs and supplied fields through existing validation contracts where
available. Pagination follows existing repository page/limit behavior; reject
arithmetic overflow rather than adding a different normalization scheme.

## Auth/session

`login` checks a username/password in an existing space membership. It does not
create a user or enroll a membership. `refresh` validates the token for the
requested space and checks that the live membership still belongs to its claimed
user and space. Both require an existing, non-deleted user, active non-deleted
space, and active non-deleted membership. Both produce current user data and
current effective role/permission slugs through repository reads before issuing
new access and refresh tokens. Refresh does not revoke previously issued tokens.

`authenticate_access_token` validates token signature, purpose, expiry, and space,
then checks live user/membership/space data and returns current effective grants,
rather than treating signed grant snapshots as current authority.
`authenticate_api_key` validates format, hashes the raw key, and uses
`read_active_by_hash` in the requested space. It returns the live owner's current
effective grants. No authentication operation consumes management caches.

Missing login identity, wrong passwords, or inactive/missing login membership
produce `Unauthorized` with a generic message and an optional retained cause.
Invalid/missing API-key identity also produces generic `Unauthorized`. Token
invalid/expired errors retain their existing classification. Invalid request
shape is `BadArgs`; database and utility operational failures remain failures.
Do not turn a database outage into invalid credentials.

Registration, forgot-password delivery, OAuth, persisted sessions, and server-side
logout/revocation are outside these interfaces.

## Management and authorization

Before every protected operation, recheck the caller's live user, active space,
active membership, ownership of that membership, and effective permission slugs
through repository contracts. Missing or inactive caller identity is unauthorized;
an authenticated caller without the required authority is forbidden. Management
cache values and caller-supplied grant snapshots must never authorize a request.

Application construction selects the control space by its immutable slug,
defaulting to the seeded `lockmate` slug. Resolve it through the Space repository;
never grant authority merely because a role is named `super` or `admin`. Environment
configuration and construction are deferred to application implementation.

Global User and Space management requires a caller authenticated in the control
space, plus the corresponding current control-space permission. Scoped entity
management permits either a caller in the target space with its corresponding
permission, or a control-space caller with that permission in the control space.
A caller in another ordinary space is forbidden. Control callers must also
resolve a non-deleted target space; inactive target spaces remain manageable.

| Operations | Required permission |
| --- | --- |
| Entity reads and filters | `<entity>:get` |
| Entity creation | `<entity>:add` |
| Entity updates | `<entity>:set` |
| Entity deletion, including bulk assignment deletion | `<entity>:remove` |
| Set default role | `role:set` |
| Reset global user password | `user_password:set` |
| Read a member's effective roles or permissions | `user_permission:get` |

Entity prefixes are `space`, `user`, `permission`, `role`, `space_member`,
`role_permission`, `member_role`, and `api_key`. Target IDs never bypass explicit
space scoping. Effective member grants are exposed by SpaceMember management;
assignment inspection is exposed by the corresponding assignment trait.

User creation hashes its plaintext password through Password. Password reset
validates and hashes the replacement without requiring the target's old password.
API-key creation validates metadata, verifies the target membership, generates a
key through the ApiKey utility, and stores only its hash/redaction. Hash lookup is
available only to authentication, not management callers.

`Role.set_default` delegates to `Role.update_by_id` with `is_default = Some(true)`;
repository locking and replacement of the previous default are preserved. New
membership creation preserves the repository's atomic default-role assignment.
Bulk assignment deletion requires at least one member/role or role/permission
selector and follows the repository's OR semantics when both are supplied.
Soft-deleted parents and existing uniqueness/conflict behavior remain governed
by repository contracts. No additional cascading deletion is introduced.

## Profile and ownership

Profile methods derive user, space, member, and audit IDs from the caller.
They accept no override of those IDs. Profile access requires the same live
caller checks described above and the following current permissions:

| Operations | Required permission |
| --- | --- |
| Read profile, roles, or permissions | `profile:get` |
| Update profile or own membership preferences | `profile:set` |
| Change password | `profile_security:set` |
| Read/list own API keys | `profile_api_key:get` |
| Create own API key | `profile_api_key:add` |
| Update own API-key metadata | `profile_api_key:set` |
| Delete own API key | `profile_api_key:remove` |

Profile results include the global safe user, current space, and current
membership. Changes to the global profile affect that user across spaces.
Password changes verify the current password before hashing and storing the new
password. Verification flags, membership activation, and role/permission grants
are never writable through profile requests.

Own-key filters are always restricted to `caller.member_id`. Read/update/delete
first check that the scoped key belongs to that membership; another member's
key is reported as `RepositoryError::ApiKeyNotFound`. Profile exposes no key hash
lookup. Self-enrollment, leaving spaces, account deletion, contact-verification
workflows, and cross-space membership discovery are not included.

## Dependencies, transactions, caching, and errors

| Usecase | Dependency contracts for future implementations |
| --- | --- |
| Session | Space, User, SpaceMember, MemberRole repositories; Password, Token, ApiKey, Validator utilities |
| Management | Matching repositories and caches; repositories for authorization; Validator; Password/ApiKey where required |
| Account | User, Space, SpaceMember, MemberRole repositories; Validator; eligible management caches |
| Security | Caller authorization repositories; User repository; Password and Validator |
| Own API keys | Caller authorization repositories; ApiKey repository/cache; ApiKey and Validator utilities |

Use Logger, Transactor, and caching Invalidation contracts where orchestration
requires them. No new dependency contracts are needed for the selected API.

Reuse repository atomic operations. For multi-write work, pass the callback's
transactional AppContext to every repository operation. Bypass caching whenever
`context.transaction` is present. Reads outside transactions follow existing
cache-aside contracts after authorization; credential and effective-grant reads
always bypass caches. Collect mutated families and invalidate only after the
outermost successful commit, never during nested callbacks or after rollback.
Membership creation invalidates SpaceMembers and MemberRoles; other changes
follow the caching family dependency table in `../contracts/AGENTS.md`. Cache failures fall back to
the database or leave committed success intact, without duplicating adapter logs.

`UsecaseError` wraps typed dependency errors, preserving codes and sources with
safe display text, and adds bad-arguments/state, unauthorized, and forbidden
classification. Unauthorized display never includes its retained cause. Rendering
uses the safe message/code, never source-chain or Debug output. Caching errors
are not part of the public failure model. Transaction callback implementations
must recover a wrapped UsecaseError when possible to retain its classification;
transaction begin/commit/rollback failures remain Transactor errors.

This module only defines interfaces and response conversions. Authorization,
validation, transaction orchestration, logging, and caching behavior must be
implemented and tested in the application layer before exposing these operations.

## Verification

Run `cargo fmt --check`, `cargo check --all-targets`, and `cargo test` from the
Lockmate root. Interface tests live under tests/domain/usecases.rs. Check object
safety, Send futures, credential-free responses, nullable updates, and error
codes/sources. Runtime authorization and ownership tests belong to application
implementations when added; interfaces alone do not enforce these policies.
