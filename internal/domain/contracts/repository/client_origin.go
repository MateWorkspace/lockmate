package domaincontractsrepository

import (
	"context"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

// ClientOrigin persists canonical origins validated by Validator.ClientOrigin.
// Space ownership is derived through the client. Ordinary reads include inactive
// records but exclude deleted origins and parents; allowlist checks require every
// record in the ownership chain to be active and nondeleted.
type ClientOrigin interface {
	Create(ctx context.Context, spaceId int64, args CreateClientOriginArgs) (id int64, err error)
	ReadById(ctx context.Context, spaceId, id int64) (*domainmodelsentity.ClientOrigin, error)
	ReadByClientIdAndOrigin(ctx context.Context, spaceId, clientId int64, origin string) (*domainmodelsentity.ClientOrigin, error)
	ReadByClientId(ctx context.Context, spaceId, clientId int64) ([]domainmodelsentity.ClientOrigin, error)
	ReadByFilter(ctx context.Context, spaceId int64, filter ClientOriginFilter) ([]domainmodelsentity.ClientOrigin, int64, error)
	UpdateById(ctx context.Context, spaceId, id int64, args UpdateClientOriginArgs) error
	DeleteById(ctx context.Context, spaceId, id int64, by *int64) error

	// IsAllowedOriginForClient checks only the specified client within its space.
	// Missing, inactive, or deleted matches return false without a not-found error.
	IsAllowedOriginForClient(ctx context.Context, spaceId, clientId int64, origin string) (bool, error)

	// IsAllowedOrigin checks all spaces for an eligible client owning this origin.
	// This supports shared-service CORS preflight, not authentication or client
	// identification. Missing origins return false; persistence failures are errors.
	IsAllowedOrigin(ctx context.Context, origin string) (bool, error)
}

type CreateClientOriginArgs struct {
	ClientId int64
	Origin   string
	IsActive *bool
	By       *int64
}

type UpdateClientOriginArgs struct {
	Origin   *string
	IsActive *bool
	By       *int64
}

type ClientOriginFilter struct {
	Page     int64
	Limit    int64
	ClientId *int64
	Search   *string
	IsActive *bool
}
