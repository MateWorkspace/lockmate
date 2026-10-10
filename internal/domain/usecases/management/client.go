package domainusecasesmanagement

import (
	"context"
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

// Client defines permission-protected administration; audit actors come from caller.
// Scoped operations require target-space or control-space authority.
type Client interface {
	Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request CreateClientRequest) (*ClientSecretResult, error)
	ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) (*domainusecasesshared.ClientView, error)
	ReadBySlug(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, value string) (*domainusecasesshared.ClientView, error)
	ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request FilterClientRequest) (domainusecasesshared.Page[domainusecasesshared.ClientView], error)
	UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64, request UpdateClientRequest) error

	// RotateSecret returns the raw replacement once, after successful commit.
	RotateSecret(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) (*ClientSecretResult, error)
	DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) error

	// Origin administration always addresses a client within the target space.
	// Origin IDs from another client return ClientOriginNotFound.
	CreateOrigin(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64, request CreateClientOriginRequest) (int64, error)
	ReadOriginById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId, id int64) (*domainmodelsentity.ClientOrigin, error)
	ReadOriginByValue(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64, origin string) (*domainmodelsentity.ClientOrigin, error)
	ReadOriginsByClientId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64) ([]domainmodelsentity.ClientOrigin, error)
	ReadOriginsByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64, request FilterClientOriginRequest) (domainusecasesshared.Page[domainmodelsentity.ClientOrigin], error)
	UpdateOriginById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId, id int64, request UpdateClientOriginRequest) error
	DeleteOriginById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId, id int64) error
}

type CreateClientRequest struct {
	Slug        string
	Name        string
	Description *string
	IsActive    *bool
	Preferences *json.RawMessage
}

type FilterClientRequest struct {
	Page     int64
	Limit    int64
	Search   *string
	IsActive *bool
}

// UpdateClientRequest leaves omitted fields unchanged.
type UpdateClientRequest struct {
	Name        *string
	Description *string
	IsActive    *bool
	Preferences *json.RawMessage
}

// ClientSecretResult exposes a raw secret only at creation or rotation.
type ClientSecretResult struct {
	Client domainusecasesshared.ClientView `json:"client"`
	Secret string                          `json:"secret"`
}

type CreateClientOriginRequest struct {
	Origin   string
	IsActive *bool
}

type UpdateClientOriginRequest struct {
	Origin   *string
	IsActive *bool
}

type FilterClientOriginRequest struct {
	Page     int64
	Limit    int64
	Search   *string
	IsActive *bool
}
