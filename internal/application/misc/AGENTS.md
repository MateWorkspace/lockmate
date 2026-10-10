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

AllowedOrigin lives in allowed_origin/usecase.go. Inject the ClientOrigin
repository, Validator, and Logger contracts. Expose global and scoped live
allowlist checks without authentication, permissions, or transactions. Reject
invalid scoped IDs as BadArgs. Known invalid/oversized origins return false, nil
without logging or persistence; unexpected validation or repository failures
return false with the original error and an explicit step-specific log. Never
swallow operational failures or use an allowance as authenticated identity.
Routine matches and nonmatches are quiet; middleware and caching are deferred.
