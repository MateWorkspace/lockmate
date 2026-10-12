// Package presentationgrpcservershared provides trusted identity access for gRPC
// servers, sourced from interceptor-injected context.
package presentationgrpcservershared

import (
	"context"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	presentationgrpcmapper "github.com/MateWorkspace/lockmate/internal/presentation/grpc/mapper"
	presentationshared "github.com/MateWorkspace/lockmate/internal/presentation/shared"
)

// Client returns the interceptor-authenticated client identity.
func Client(ctx context.Context) (*domainmodelsentity.ClientIdentity, error) {
	client := presentationshared.ClientIdentity(ctx)
	if client == nil {
		return nil, presentationgrpcmapper.Error(domainmodelserror.NewError("client authentication is required", domainmodelserror.Unauthorized, nil))
	}

	return client, nil
}

// Claims returns the interceptor-authenticated access claims.
func Claims(ctx context.Context) (*domainmodelsentity.AccessClaims, error) {
	claims := presentationshared.AccessClaims(ctx)
	if claims == nil {
		return nil, presentationgrpcmapper.Error(domainmodelserror.NewError("bearer authentication is required", domainmodelserror.Unauthorized, nil))
	}

	return claims, nil
}
