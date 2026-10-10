package domainusecasesmanagement

import (
	"context"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

// RolePermission defines permission-protected administration; audit actors come from caller.
// Scoped operations require target-space or control-space authority.
type RolePermission interface {
	Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request CreateRolePermissionRequest) (id int64, err error)
	ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) (*domaincontractsrepository.RolePermissionDetails, error)
	ReadByRoleIdAndPermissionId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, roleId, permissionId int64) (*domaincontractsrepository.RolePermissionDetails, error)
	ReadByRoleId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, value int64) ([]domaincontractsrepository.RolePermissionWithPermission, error)
	ReadByPermissionId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, value int64) ([]domaincontractsrepository.RolePermissionWithRole, error)
	DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) error
	DeleteByRoleIdOrPermissionId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request DeleteRolePermissionRequest) error
}

type CreateRolePermissionRequest struct {
	RoleId       int64
	PermissionId int64
}

// DeleteRolePermissionRequest requires at least one selector; two selectors use OR within the space.
type DeleteRolePermissionRequest struct {
	RoleId       *int64
	PermissionId *int64
}
