package domainusecasesmanagement

import (
	"context"
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

// Space defines permission-protected administration; audit actors come from caller.
// Global operations require an authenticated control-space caller.
type Space interface {
	Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, request CreateSpaceRequest) (id int64, err error)
	ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) (*domainmodelsentity.Space, error)
	ReadBySlug(ctx context.Context, caller *domainmodelsentity.AccessClaims, value string) (*domainmodelsentity.Space, error)
	ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, request FilterSpaceRequest) (domainusecasesshared.Page[domainmodelsentity.Space], error)
	UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64, request UpdateSpaceRequest) error
	DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) error
}

type CreateSpaceRequest struct {
	Slug        string
	Name        string
	Description *string
	IsActive    *bool
	Preferences *json.RawMessage
}

type FilterSpaceRequest struct {
	Page     int64
	Limit    int64
	Search   *string
	IsActive *bool
}

// UpdateSpaceRequest leaves omitted fields unchanged.
type UpdateSpaceRequest struct {
	Name        *string
	Description *string
	IsActive    *bool
	Preferences *json.RawMessage
}
