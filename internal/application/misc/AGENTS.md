# Miscellaneous application helpers

Expose small utility operations needed by presentation through domain usecases.
Keep one folder per capability, starting with tracer/usecase.go. Use a private
usecase struct, explicit contract dependencies, and NewUsecaseImpl returning the
domain interface. Presentation calls this interface through application rather
than depending on utility infrastructure.

Tracer receives clientSlug and clientExecId, generates a fresh event UUID through
the tracer utility contract, assembles the domain Trace, and delegates injection.
External callers cannot choose EventId. Empty slug and nil execution UUID remain
absent diagnostic metadata; trace is not authenticated identity.

Preserve the derived context's parent values, cancellation, and deadlines. New
injection replaces the previous trace. Generation failure returns nil context,
does not inject, and logs directly with the specific event-ID generation failure
message and err metadata. Do not use shared logging wrappers or defers. Success
has no log or transaction.

Use domain import aliases, context-first methods, early returns, and readable
whitespace. Keep unrelated helpers separate; add capabilities only when needed.
Tests live in tests/application/misc and use injected contracts without mutating
global UUID randomness.

IpAllowlist lives in ip_allowlist/usecase.go. Inject the ClientIpAllowlist and
Logger contracts. Expose the client-scoped live allowlist check without
authentication, permissions, or transactions. Reject invalid scoped IDs as
BadArgs and short-circuit loopback. Unparseable caller IPs return false, nil
without logging or persistence; unexpected repository failures
return false with the original error and an explicit step-specific log. Never
swallow operational failures or use an allowance as authenticated identity.
Routine matches and nonmatches are quiet; transport IP resolution and caching
are deferred.

auth_flow/usecase.go purges expired launch, attempt, and generic handoff records
in batches of 1000, respecting cancellation. Each failed stage writes its own
explicit warning and returns its error. Cleanup does not authenticate records;
expiration is always enforced during use. The cron presentation owns the schedule.

signing_key/usecase.go implements Reconcile through the SigningKey repository and
Secret utility. It re-seals keys whose fingerprint differs from the current
key-encryption key, creates the active key when missing, rotates the active key
once older than the configured rotation interval, and soft-deletes retiring keys
past retention. Creation and rotation own a strict transaction with a
single-conflict retry; re-seal and cleanup are individual statements. Composition
runs Reconcile on the signing-key reconcile interval. It logs committed
creations, rotations, re-seals, and retirement counts.

jwks/usecase.go returns the active and retiring public JWKs from the SigningKey
repository. It performs no authentication, permissions, or transactions; the
presentation layer gates the endpoint with client credentials. An empty result
lists an empty slice.
