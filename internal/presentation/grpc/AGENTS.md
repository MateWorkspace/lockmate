# gRPC presentation

Expose Lockmate's existing application usecases over gRPC. This mirrors the HTTP
API and reuses the same domain usecases; it holds no business rules.

## Layout and style

- `mapper/`: wire/entity mappers, preferences (raw JSON string), audit, pagination,
  and the centralized error → gRPC status + typed detail mapping.
- `interceptor/`: the unary pipeline and method classification.
- `server/<group>/<name>/server.go`: private `server` structs embedding the
  generated `Unimplemented*Server`, `NewServer` returning the generated
  interface. Handlers validate request shape, read trusted identity from context,
  call the usecase, and map results/errors. No business logic.
- `register/`: per-group dependency structs with nil guards, plus the standard
  unauthenticated `grpc.health.v1.Health` service.

Use concatenated package names (`presentationgrpcservermanagementrole`) and
explicit import aliases.

## Metadata contract

Client credentials and identity travel in gRPC metadata, not request messages:

- `x-client-slug`, `x-client-secret`, `x-client-exec-id` (nonzero UUID) on every
  business RPC.
- `authorization: Bearer <access-token>` on protected profile/management RPCs.
- the configured client-IP header (lowercased, default `cf-connecting-ip`) is read
  only from trusted peers.

## Interceptors and classification

Chain: recovery → trace → client authentication → caller IP allowlist → bearer
authentication. Trace uses `misc.Tracer`; client auth uses `auth.Client`; the IP
allowlist reuses `misc.IpAllowlist` with the shared `presentation/shared`
resolver and `LOCKMATE_TRUSTED_PROXIES`; bearer uses
`Session.AuthenticateAccessToken` and reloads live grants.

Method classification: public = `VersionService.Get` and the health service;
client-auth only = `SessionService.*`, `HandoffService.Exchange`,
`GoogleService.Start`, `JwksService.Get`; client-auth + bearer = `AccountService`,
`SecurityService`, and every management service. Never build trusted identity from
request fields; read it from the interceptor-injected context
(`presentation/shared`).

## Errors

`mapper.Error` classifies domain sentinels into gRPC codes and a typed
`lockmate.v1.Error` detail. Never expose `Error()` or its source chain. Context
cancellation/deadline map to `Canceled`/`DeadlineExceeded`.

## Transport

The server is plaintext HTTP/2 (h2c); a TLS terminator (cloudflared/Caddy/nginx)
may run in front, forwarding the real client IP in metadata from a trusted peer.
Google browser `launch`/`callback` remain HTTP-only. Regenerate bindings from
`../lockmate-proto`; never hand-edit generated files. Tests live under
`tests/presentation/grpc`.
