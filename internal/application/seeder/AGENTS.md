# Startup seeder

Implement domainusecasesseeder.Seeder through domain contracts and seed models.
Keep the private usecase and explicit constructor in usecase.go; use entity-focused
space, permission, role, role_permission, user, and client helpers. validation.go
checks definitions and references before persistence; changes.go stores identifiers
for success logging after commit. Do not import database/seeder or concrete adapters.

Run uses one strict application transaction for the entire bootstrap. Repository
helpers reuse its context. Seed in dependency order and retry a known uniqueness
conflict once using a fresh transaction. Warn at the failed step only if another
attempt is available; exhaustion logs the step error and outer transaction failure.
Do not log success inside the transaction or use a deferred/shared logging wrapper.
Write separate explicit info calls for each committed entity/assignment type.

Preserve existing profiles, credentials, redirect settings, defaults, metadata,
and activation state. Reuse natural keys and verify client space ownership.
Create missing permissions, roles, memberships, clients, IP allowlist entries, and configured
grants. Enabled startup restores removed baseline role/member grants. Membership
creation assigns the current default role automatically; do not duplicate it.
Keep an existing default role; if no default exists after role seeding, fail
rather than reset an existing role's metadata. Do not reactivate disabled records.

Validate all fields and duplicate/reference/default rules before any writes.
Composition injects runtime credentials from LOCKMATE_SEEDER_* settings into the
seed models; do not read configuration or environment variables here. Validate
passwords and client secrets through existing Validator/Keygen contracts, then
hash user passwords through Password and configured client secrets through Keygen.
Redact client display values with the established first/last four-character format.
Never log definitions or secrets. Bootstrap writes have null audit actors.

The composition launcher controls SeederEnabled, default true, and invokes Run
before HTTP or maintenance starts. There is no HTTP seeder route or user permission
bypass: this is an internal startup operation. Tests cover PostgreSQL atomicity,
concurrent startup, idempotence, preservation/restoration, policy, and credentials.
