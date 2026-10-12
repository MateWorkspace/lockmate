package applicationshared

import (
	"context"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
)

// ReadAuthClient rechecks saved scope and exact redirect against live records.
func ReadAuthClient(ctx context.Context, clients domaincontractsrepository.Client, spaces domaincontractsrepository.Space, clientId, spaceId int64, redirect string) (*domainmodelsentity.Client, error) {
	client, err := clients.ReadById(ctx, spaceId, clientId)
	if err != nil {
		return nil, IdentityError(err)
	}

	if client == nil {
		return nil, Unauthorized(nil)
	}

	client, err = ReadClient(ctx, clients, spaces, &domainmodelsentity.ClientIdentity{
		Id:      clientId,
		SpaceId: spaceId,
		Slug:    client.Slug,
	})
	if err != nil {
		return nil, err
	}

	if client.AuthRedirectUri == nil || *client.AuthRedirectUri != redirect {
		return nil, domainmodelserror.NewError("client authentication redirect is no longer configured", domainmodelserror.BadState, nil)
	}

	return client, nil
}
