# Shared application helpers

Support feature orchestration without depending on concrete infrastructure.
identity.go owns live identity/client and management-scope checks; validation.go
checks IDs, pagination, JSON, and nullable updates; projection.go maps safe
credential views; session.go resolves grants and assembles token claims.

Helpers depend on domain contracts and models. Keep each responsibility explicit,
with small functions, import aliases, early returns, and whitespace. Do not add
global state, a universal dependency bundle, generic CRUD dispatch, or transport
types. IdentityReader is a focused reusable live identity checker.

Shared checks do not call PermissionGuard: each protected feature checks it
immediately below its tag. Helpers return errors without logging. Feature methods
write direct, specific Error/Warn/Info calls at the handling point with inline
metadata; there are no shared logging wrappers, classification registries, or
logging defers. Required failures remain errors; expected optional-record absence
is not a warning. Do not swallow persistence or authorization errors.

Session helpers require caller-supplied repository contexts and never begin or
commit transactions. They do not enroll during refresh/access authentication.
EnsureMembership is used only after verified sign-in or registration. Feature
methods own retries, commit boundaries, success logs, and credential outputs.

Safe projections exclude hashes by construction and reuse domain repository
summaries. Pagination does not promise a database snapshot. Test doubles and
fixtures live in tests/application/shared, outside production helpers.
