package domainusecasesmanagement

import (
	"context"
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

// Permission defines permission-protected administration; audit actors come from caller.
// Scoped operations require target-space or control-space authority.
type Permission interface {
	Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request CreatePermissionRequest) (id int64, err error)
	ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) (*domainmodelsentity.Permission, error)
	ReadBySlug(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, value string) (*domainmodelsentity.Permission, error)
	ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request FilterPermissionRequest) (domainusecasesshared.Page[domainmodelsentity.Permission], error)
	UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64, request UpdatePermissionRequest) error
	DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) error
}

type CreatePermissionRequest struct {
	Slug        string
	Name        string
	Description *string
	Preferences *json.RawMessage
}

type FilterPermissionRequest struct {
	Page   int64
	Limit  int64
	Search *string
}

// UpdatePermissionRequest leaves omitted fields unchanged.
type UpdatePermissionRequest struct {
	Name        *string
	Description *string
	Preferences *json.RawMessage
}
