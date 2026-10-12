package presentationhttputils

import (
	"context"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	presentationshared "github.com/MateWorkspace/lockmate/internal/presentation/shared"
)

func InjectClientIdentity(ctx context.Context, client *domainmodelsentity.ClientIdentity) context.Context {
	return presentationshared.InjectClientIdentity(ctx, client)
}

func ClientIdentity(ctx context.Context) *domainmodelsentity.ClientIdentity {
	return presentationshared.ClientIdentity(ctx)
}

func InjectAccessClaims(ctx context.Context, claims *domainmodelsentity.AccessClaims) context.Context {
	return presentationshared.InjectAccessClaims(ctx, claims)
}

func AccessClaims(ctx context.Context) *domainmodelsentity.AccessClaims {
	return presentationshared.AccessClaims(ctx)
}
