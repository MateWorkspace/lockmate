// Package presentationshared holds small helpers shared by the HTTP and gRPC
// presentations: trusted identity context values and caller IP resolution.
package presentationshared

import (
	"context"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type contextKey uint8

const (
	clientIdentityContextKey contextKey = iota
	accessClaimsContextKey
)

// InjectClientIdentity stores the interceptor/middleware-authenticated client.
func InjectClientIdentity(ctx context.Context, client *domainmodelsentity.ClientIdentity) context.Context {
	return context.WithValue(ctx, clientIdentityContextKey, client)
}

// ClientIdentity returns the trusted client identity, or nil.
func ClientIdentity(ctx context.Context) *domainmodelsentity.ClientIdentity {
	client, _ := ctx.Value(clientIdentityContextKey).(*domainmodelsentity.ClientIdentity)
	return client
}

// InjectAccessClaims stores the authenticated access claims.
func InjectAccessClaims(ctx context.Context, claims *domainmodelsentity.AccessClaims) context.Context {
	return context.WithValue(ctx, accessClaimsContextKey, claims)
}

// AccessClaims returns the trusted access claims, or nil.
func AccessClaims(ctx context.Context) *domainmodelsentity.AccessClaims {
	claims, _ := ctx.Value(accessClaimsContextKey).(*domainmodelsentity.AccessClaims)
	return claims
}
