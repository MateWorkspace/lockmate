# Database assets

`migrations/` contains timestamped PostgreSQL up/down migrations adapted from
Nadi. The initial users table includes nullable `avatar_path`, an object storage
key. Apply migrations before starting the service, for example with SQLx CLI:

```sh
sqlx migrate run --source database/migrations
```

The command uses `DATABASE_URL`. Roll back one migration with
`sqlx migrate revert --source database/migrations`.

`seeder/` contains the initial permissions, roles, and users. Rust embeds these
JSON files at build time; `config::seeder::load()` returns a `SeedData` value for
application composition. It reads no runtime files and performs no database
writes. The application seeder will create records through repository contracts
and hash each seed user's password through the password contract.

The copied users have the initial password `12345678`; replace it when preparing
these seed accounts for deployment. Avatar paths are absent and load as `None`.
