package domainusecasesmanagement

import (
	"context"
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

// Role defines permission-protected administration; audit actors come from caller.
// Scoped operations require target-space or control-space authority.
type Role interface {
	Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request CreateRoleRequest) (id int64, err error)
	ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) (*domainmodelsentity.Role, error)
	ReadBySlug(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, value string) (*domainmodelsentity.Role, error)
	ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request FilterRoleRequest) (domainusecasesshared.Page[domainmodelsentity.Role], error)
	ReadDefault(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64) (*domainmodelsentity.Role, error)
	UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64, request UpdateRoleRequest) error

	// SetDefault atomically replaces the previous default role.
	SetDefault(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) error
	DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) error
}

type CreateRoleRequest struct {
	Slug        string
	Name        string
	Description *string
	IsDefault   *bool
	Preferences *json.RawMessage
}

type FilterRoleRequest struct {
	Page      int64
	Limit     int64
	Search    *string
	IsDefault *bool
}

// UpdateRoleRequest leaves omitted fields unchanged.
type UpdateRoleRequest struct {
	Name        *string
	Description *string
	IsDefault   *bool
	Preferences *json.RawMessage
}
