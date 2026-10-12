# Embedded seed definitions

Embed space.json, permission.json, role.json, user.json, and client.json through
seeder.go. Load() decodes all documents into domainmodelsseeder.Data and publishes
global SeedData only on complete success. Reject unknown fields and trailing JSON.
The loader performs no database writes, password hashing, or logging.

space.json is the sole configured source of the control-space slug, defaulting to
lockmate. Composition uses it for management authorization even when seeding is
disabled. Permission and role definitions are scoped to this space. Keep slugs
separate from readable names, and align permission.json with the application
permission catalog. Super has every permission; Admin manages users, credentials,
clients, IP allowlist entries, and memberships but only reads authorization catalogs; User has
self-service profile/security grants and is the declared default role.

User passwords and client secrets are supplied through LOCKMATE_SEEDER_* settings,
never stored in JSON or given embedded defaults. Password and SecretKey are runtime
model fields with json:"-". Composition clones the user/client definitions before
binding credentials; SeedData stays credential-free. Super/Admin/User passwords
map by role_slug; the baseline control client receives SeederClientSecretKey.
Metadata/grant changes require rebuilding, while credential values are runtime
settings. Seed reruns do not rotate persisted credentials. Never print seed data,
passwords, client secrets, or configuration in diagnostics. Runtime .env is not
baked into Docker images. Bootstrap accounts use null audit actors.

Model definitions live under internal/domain/models/seeder; application consumes
those models rather than importing this loader. Create behavior and validation
belong in internal/application/seeder. Tests live under tests/database/seeder and
tests/application/seeder. Verify catalog parity independently of the seed data.
