# Database assets

## Purpose

Maintain the PostgreSQL schema and embedded seed definitions. Runtime database
access belongs to repository adapters; loading and graph validation live in config,
and future seeding execution belongs to application.

## Layout

- `migrations/`: timestamped `.up.sql` / `.down.sql` pairs, one table per migration.
- `seeder/space.json`: spaces with their nested permission and role catalogs.
- `seeder/user.json`: global users and initial account inputs.
- `seeder/space_member.json`: memberships and additional role assignments.

## Code style

Keep SQL readable and consistent with neighboring migrations. Order table
creation before dependent foreign keys and provide matching down migrations.
Preserve PostgreSQL pg_trgm, partial live-record uniqueness, audit fields, scoped
relationships, and established soft-delete behavior. Align schema changes with
`../docs/erd.dbml`, domain entities, repository queries/scanners, and tests.
Never rewrite migrations already applied outside development; this initial schema
was previously edited in place only because it had not been deployed.

Keep seed JSON two-space indented with snake_case fields and no unknown fields.
Use slugs for stable references and names for display. Keep exactly one default
role per seeded space; listed membership roles are additional to that default.
Preserve cross-space references and graph-validation rules described below.

## Migration and seed behavior

`migrations/` contains timestamped PostgreSQL up/down migrations adapted from
Nadi. The initial users table includes nullable `avatar_path`, an object storage
key. Apply migrations before starting the service, for example with SQLx CLI:

```sh
sqlx migrate run --source database/migrations
```

The command uses `DATABASE_URL`. Roll back one migration with
`sqlx migrate revert --source database/migrations`.

`seeder/` contains embedded definitions for spaces, global users, and memberships.
`space.json` nests each space's independent permission and role catalog.
`user.json` stores global account inputs, including plaintext passwords for the
password utility. `space_member.json` connects usernames to space slugs and lists
additional role slugs. References use slugs, while names are display labels.

The baseline defines one active `lockmate` space with 39 permissions and three
roles: super has all 39, admin has 17, and default user has seven. These include
profile API-key create/read/update/delete permissions.
The three existing users each receive a Lockmate membership. Super and admin
list their additional role; user lists none because membership creation already
assigns the default. Listed roles are additive and do not replace that default.
These definitions do not establish application authorization policies.

Rust embeds the JSON files at build time. `config::seeder::load()` returns a
`SeedData` value after JSON parsing and graph validation; `parse()` accepts the
three JSON strings for the same checks. Loading rejects unknown fields, duplicate
definitions/references, missing or multiple default roles, and references outside
the appropriate space. Slug reuse across spaces is valid. Space and membership
active flags default to true; additional membership roles default to empty.
Scalar validation belongs to the validator utility and is separate from graph
validation. Embedded definitions are checked with the field validators in tests.

Loading reads no runtime files and performs no database writes. The future
application seeder will create spaces, their catalogs and role-permission
assignments, global users, and memberships through repository contracts, then
add explicit member-role assignments. Password hashing uses the password contract.

The copied users retain the initial password `12345678`; replace it when preparing
these seed accounts for deployment. Avatar paths are absent and load as `None`.

## Verification

Run `cargo test config::seeder` from the Lockmate root for graph consistency and
field validation. Migration integration tests apply all eight migrations in
order under isolated schemas using LOCKMATE_TEST_DATABASE_URL; keep pg_trgm
available in public. Use throwaway databases for migration, rollback, and seeding
checks; do not modify a developer's database as routine validation.
