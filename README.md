# Lockmate

One authentication service for many applications.

Lockmate gives Mate projects a shared way to sign users in, manage their profiles,
and control access. Users have one identity, while each project **space** has its
own memberships, roles, and permissions. Each application connects through a
**client** with its own credentials.

Built with Go, Gin, and PostgreSQL.

## Features

- Username/password registration, sign-in, and JWT session refresh.
- Google sign-in with automatic registration for new users.
- User profile and password management.
- Administration of users, spaces, memberships, roles, permissions, and clients.
- Per-client caller IP allowlists with trusted-proxy client IP resolution.
- Client callback URLs and single-use authentication handoffs to application backends.
- Swagger API documentation and structured logging.

## Get started

You need PostgreSQL and either Go **1.26.5 or newer** or Docker with a recent Docker
Compose version. Keep the `mate-golang` checkout beside Lockmate; it supplies the
shared transaction library used by local and Docker builds:

```text
workspace/
├── lockmate/
└── mate-golang/
```

For a fresh checkout, create your configuration:

```sh
cd lockmate
cp .env.example .env
```

Edit `.env` to set your PostgreSQL connection, the refresh-token secret, a
signing-key encryption key, and bootstrap credentials:

```dotenv
LOCKMATE_TOKEN_REFRESH_SECRET=<your-refresh-token-secret>
# openssl rand -base64 32
LOCKMATE_SIGNING_KEY_ENCRYPTION_KEY=<32-byte-base64-key>
LOCKMATE_SEEDER_SUPER_PASSWORD=<your-super-password>
LOCKMATE_SEEDER_ADMIN_PASSWORD=<your-admin-password>
LOCKMATE_SEEDER_USER_PASSWORD=<your-user-password>
LOCKMATE_SEEDER_CLIENT_SECRET_KEY=<lockmate-followed-by-64-lowercase-hex-characters>
```

Docker startup applies the ordered [migrations](database/migrations) automatically.
For native Go startup, apply them with your migration tool first. Lockmate then
seeds its control space, permissions, roles, users, and client before serving HTTP.
Startup also reconciles the ES256 access-token signing key: it creates the active
key when missing and rotates it on the configured interval. See the
[database diagram](docs/erd.dbml) for the relationships.

### Run with Go

```sh
go run ./cmd/main
```

### Run with Docker

If PostgreSQL runs on your host machine, set `LOCKMATE_POSTGRES_HOST=dodol.local`
or `host.docker.internal`, and use its host-accessible port. Compose maps both
names to the host gateway. `127.0.0.1` inside the container refers to the container
itself.

```sh
docker compose up -d --build
docker compose logs -f app
```

The container name and published ports come from:

```dotenv
DOCKER_COMPOSE_CONTAINER_NAME=lockmate-app
DOCKER_COMPOSE_HTTP_HOST_PORT=10701
DOCKER_COMPOSE_GRPC_HOST_PORT=10702
```

The container's target ports come from `LOCKMATE_PORT_HTTP_SERVER` and
`LOCKMATE_PORT_GRPC_SERVER`. The gRPC port is reserved; its presentation layer is
not implemented yet.

## Bootstrap accounts

The default control space and client slug are both `lockmate`. The bundled users
are `super`, `admin`, and `user`; their passwords come from the corresponding
`LOCKMATE_SEEDER_SUPER_PASSWORD`, `LOCKMATE_SEEDER_ADMIN_PASSWORD`, and
`LOCKMATE_SEEDER_USER_PASSWORD` variables. The control client's secret comes from
`LOCKMATE_SEEDER_CLIENT_SECRET_KEY`: `lockmate-` followed by 64 lowercase hexadecimal
characters. All four values are required when seeding is enabled, with no built-in
credential defaults. Passwords use the same validation as normal user passwords.

Super has full administration; Admin manages users, credentials, clients, and
memberships; User has self-service profile/security access.

Edit [the seed JSON](database/seeder) and rebuild to customize metadata and grants.
The control-space slug is defined in `space.json`; it has no env setting. Credentials
are supplied at runtime. Existing passwords, client secrets, and metadata are
preserved on reruns, so changing environment values does not rotate stored secrets.
Missing baseline grants are restored. Set `LOCKMATE_SEEDER_ENABLED=false` to skip
seeding and make its credential settings optional. Rotate existing credentials
through the profile or management API.

## Use the API

With the default HTTP port, open [Swagger](http://localhost:10701/api/docs/index.html)
to explore the endpoints. Check service information at
[`/api/version`](http://localhost:10701/api/version).

Client-authenticated requests require these headers:

```http
X-Client-Slug: <client-slug>
X-Client-Secret: <client-secret>
X-Client-Exec-Id: <nonzero-execution-UUID>
```

Keep client secrets on your application backend. The authenticated client selects
the space; callers do not choose it through request parameters. Use a fresh
execution UUID to correlate each operation.

For example, sign in with classic credentials:

```sh
curl http://localhost:10701/api/v1/auth/sign-in \
  -H 'Content-Type: application/json' \
  -H 'X-Client-Slug: <client-slug>' \
  -H 'X-Client-Secret: <client-secret>' \
  -H 'X-Client-Exec-Id: 60984d94-12ce-4d38-bd60-e01f25994119' \
  -d '{"username":"alex","password":"your-password"}'
```

Sign-in and registration return access and refresh tokens. Add
`Authorization: Bearer <access-token>` alongside the client headers for profile
and management operations. Registration is available at `/api/v1/auth/register`;
refresh uses `/api/v1/auth/refresh`.

Google authentication is disabled by default. Follow the
[Google setup and integration guide](docs/google-auth.md) to configure credentials,
the client callback URL, and the browser-to-backend handoff flow.

## Development

```sh
go test -race ./...
go vet ./...
go build ./...
make swagger
```

Set `LOCKMATE_TEST_DATABASE_URL` to an isolated PostgreSQL instance to include
database integration tests; they skip when it is absent. `make swagger` regenerates
the API documentation after changes.

Architecture and contribution conventions are documented in [AGENTS.md](AGENTS.md)
and the scoped guides linked there.
