# Google authentication integration

Lockmate's development public origin is `https://lockmate-dev.dhonan.my.id`.
The HTTP server defaults to port `10701`; the existing cloudflared tunnel forwards
public HTTPS to `http://localhost:10701`. URLs are built from configuration,
never from Host or forwarded headers. Google authentication is disabled by default.

## Google Console setup

1. Create/select a development project at [Google Cloud Console](https://console.cloud.google.com/).
2. Open Google Auth Platform. In Branding, use `Lockmate Dev`, supply your support
   and developer contact emails, and add `dhonan.my.id` as the authorized domain.
3. In Audience, select External and leave publishing status Testing. Add your
   development Google account as a test user. Basic OpenID/email/profile sign-in
   has an exception to Google's test-user restriction; adding your account still
   documents the intended development audience.
4. In Clients, create a **Web application** OAuth client named `Lockmate Dev`.
5. Register this exact authorized redirect URI, with no trailing slash:

   ```text
   https://lockmate-dev.dhonan.my.id/api/v1/auth/google/callback
   ```

6. Leave authorized JavaScript origins empty: authorization and code exchange
   are performed server-side. Copy the client ID and client secret into local
   configuration. No service account, Gmail, Drive, or People API is needed.
7. Request only `openid email profile`. Lockmate does not request offline access
   or retain Google access/refresh tokens.

Google requires an exact redirect URI match. See the official
[web-server OAuth guide](https://developers.google.com/identity/protocols/oauth2/web-server),
[OpenID Connect guide](https://developers.google.com/identity/openid-connect/openid-connect),
[branding settings](https://support.google.com/cloud/answer/15549049?hl=en), and
[audience settings](https://support.google.com/cloud/answer/15549945?hl=en).

## Local configuration and migrations

Keep existing PostgreSQL and JWT settings. Add these settings to your ignored `.env`
(or process environment, which takes precedence):

```dotenv
LOCKMATE_PORT_HTTP_SERVER=10701
LOCKMATE_HTTP_PUBLIC_URL=https://lockmate-dev.dhonan.my.id
LOCKMATE_GOOGLE_OAUTH_ENABLED=true
LOCKMATE_GOOGLE_OAUTH_CLIENT_ID=<issued-client-id>
LOCKMATE_GOOGLE_OAUTH_CLIENT_SECRET=<issued-client-secret>
LOCKMATE_GOOGLE_OAUTH_LAUNCH_DURATION=2m
LOCKMATE_GOOGLE_OAUTH_ATTEMPT_DURATION=10m
LOCKMATE_GOOGLE_OAUTH_HTTP_TIMEOUT=10s
LOCKMATE_AUTH_HANDOFF_DURATION=1m
LOCKMATE_AUTH_FLOW_CLEANUP_INTERVAL=5m
```

Docker startup applies pending migrations automatically. For native Go startup,
apply the ordered migrations with your migration runner first. They add `clients.auth_redirect_uri`, `google_oauth_launches`,
`google_oauth_attempts`, and `auth_handoffs`. Existing client rows retain a null
redirect until configured. Application startup seeds the control-space baseline unless LOCKMATE_SEEDER_ENABLED
is false. The default control client has no callback until configured.

Run `go run ./cmd/main` from the repository root. Swagger is available at
`https://lockmate-dev.dhonan.my.id/api/docs/index.html`.

Google routes return 501 while disabled. Enabling Google requires its credentials
and a canonical HTTPS public origin. A development Google client should remain
separate from future production credentials and redirect URIs.

## Configure each integrating client

Create or update the client through existing permission-protected management
endpoints, setting its nullable `auth_redirect_uri`, for example:

```json
{"auth_redirect_uri":"https://project.example.com/auth/callback"}
```

The URL is the integrating application's **backend** callback. Each client has one
exact URI. Clients in the same space can have different URIs. Queries, fragments,
credentials, and wildcards are rejected. HTTPS callback URLs are always accepted;
HTTP requires LOCKMATE_AUTH_REDIRECT_ALLOW_HTTP=true. The default is false,
which requires HTTPS even for localhost callbacks. The URI is independent of
client IP allowlists.

Clear it with `{"auth_redirect_uri_set_null":true}`. Do not supply a replacement
and clearing flag together. Clearing/changing it invalidates flows saved against
the previous value. Null leaves classic JSON authentication available.

## Backend integration flow

1. The application backend generates an unpredictable URL-safe correlation
   `state` (32–512 characters), saves it in the user's server-side login session,
   and calls `POST /api/v1/auth/google/start`:

   ```http
   X-Client-Slug: <client-slug>
   X-Client-Secret: <client-secret>
   X-Client-Exec-Id: <nonzero-UUID>
   Content-Type: application/json
   ```

   ```json
   {"state":"f8d5b65dc01947658b04e768c98771ae"}
   ```

   The result is `{"launch_url":"https://lockmate-dev.dhonan.my.id/api/v1/auth/google/launch?ticket=..."}`.
   Redirect the browser to that URL. No caller-supplied redirect URL is accepted.

2. Lockmate consumes the launch ticket, creates an independent OAuth state/nonce/
   PKCE attempt, sets its Secure/HttpOnly/SameSite=Lax browser-binding cookie,
   and redirects to Google. Concurrent browser flows use separate cookies.

3. Google navigates the browser to Lockmate's fixed callback. Lockmate checks state
   and browser binding before consuming the attempt and exchanging the code.
   It verifies the ID token and matches identity by Google's stable subject,
   never by coincidental email. First-time sign-in creates the user and membership.
   Suspended users/memberships cannot be reactivated by signing in.

4. Lockmate redirects to the client's configured backend:

   ```text
   https://project.example.com/auth/callback?code=<handoff-code>&state=<project-state>
   ```

   The application backend **must compare state with its saved login session**
   before proceeding, then clear that state. The project state is distinct from
   Google's internal OAuth state. Validated Google denial redirects with
   `error=access_denied` and project state; invalid state/binding produces a local
   sanitized error and no application redirect.

5. The application backend sends `POST /api/v1/auth/exchange`, using its client
   headers and a fresh nonzero execution UUID:

   ```json
   {"code":"<handoff-code>"}
   ```

   The response is the same session shape as classic sign-in, including fresh
   access/refresh tokens, user, space, membership, roles, permissions, and
   `is_new_user`. The exchange reloads live grants and consumes the code atomically.
   Wrong-client, expired, or replayed codes cannot create a session. A failed
   transaction rolls back consumption so the backend may retry before expiry.

6. The integrating backend establishes its own user session and redirects to its
   frontend using a destination held in that backend's login session. Clean the
   handoff URL from browser history and avoid rendering third-party resources on
   the callback page. Do not retain codes in access logs.

Client secrets remain on application backends. Lockmate JWTs never appear in
redirect URLs. The temporary Lockmate cookie supports top-level navigation;
business APIs are backend-to-backend and rely on client IP allowlists rather than
browser CORS.

## Persistence, operations, and verification

Flow bearer values are persisted as SHA-256 digests. The PKCE verifier exists only
in its transient attempt. Handoffs contain identity references rather than JWTs.
Expiration is enforced during use, independently of periodic bounded cleanup.
Cleanup failures are warnings; they do not authorize expired records.

Start Docker to migrate/seed (or apply migrations before native startup), configure
the client URI, enable Google, and manually
exercise successful login, first-time registration, and cancellation through the
public domain. Also confirm that replay and a second client's exchange fail.

Automated tests use a local fake OAuth/OIDC provider and isolated PostgreSQL;
they do not demonstrate a real Google login. Run:

```sh
go test -race ./...
go vet ./...
go build ./...
```

Set `LOCKMATE_TEST_DATABASE_URL` to an isolated PostgreSQL instance to run database
integration tests. Regenerate Swagger with `make swagger` after annotation changes.
Hosted/embedded login UI, logout/revocation, recovery/OTP, Redis, and gRPC remain
separate features. Generic `auth_handoffs` and `/auth/exchange` are reusable by a
future classic browser-login flow.
