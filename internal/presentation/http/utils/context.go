package presentationhttputils

import (
	"context"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type contextKey uint8

const (
	clientIdentityContextKey contextKey = iota
	accessClaimsContextKey
)

func InjectClientIdentity(ctx context.Context, client *domainmodelsentity.ClientIdentity) context.Context {
	return context.WithValue(ctx, clientIdentityContextKey, client)
}

func ClientIdentity(ctx context.Context) *domainmodelsentity.ClientIdentity {
	client, _ := ctx.Value(clientIdentityContextKey).(*domainmodelsentity.ClientIdentity)
	return client
}

func InjectAccessClaims(ctx context.Context, claims *domainmodelsentity.AccessClaims) context.Context {
	return context.WithValue(ctx, accessClaimsContextKey, claims)
}

func AccessClaims(ctx context.Context) *domainmodelsentity.AccessClaims {
	claims, _ := ctx.Value(accessClaimsContextKey).(*domainmodelsentity.AccessClaims)
	return claims
}
