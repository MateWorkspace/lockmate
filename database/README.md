# Database assets

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

The baseline defines one active `lockmate` space with 36 permissions and three
roles: super has all 36, admin retains its 14, and default user retains its four.
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
