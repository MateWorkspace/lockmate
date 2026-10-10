package domainusecasesmisc

import "context"

// AllowedOrigin exposes live CORS allowlist checks before authentication.
// Invalid origins are ordinary nonmatches; operational failures remain errors.
// An allowance never authenticates or identifies a client.
type AllowedOrigin interface {
	IsAllowedOrigin(ctx context.Context, origin string) (bool, error)
	IsAllowedOriginForClient(ctx context.Context, spaceId, clientId int64, origin string) (bool, error)
}
