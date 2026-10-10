package domaincontractsrepository

import (
	"context"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type RolePermission interface {
	Create(ctx context.Context, spaceId int64, args CreateRolePermissionArgs) (id int64, err error)
	ReadById(ctx context.Context, spaceId, id int64) (*RolePermissionDetails, error)
	ReadByRoleIdAndPermissionId(ctx context.Context, spaceId, roleId, permissionId int64) (*RolePermissionDetails, error)
	ReadByRoleId(ctx context.Context, spaceId, value int64) ([]RolePermissionWithPermission, error)
	ReadByPermissionId(ctx context.Context, spaceId, value int64) ([]RolePermissionWithRole, error)
	DeleteById(ctx context.Context, spaceId, id int64) error
	DeleteByRoleIdOrPermissionId(ctx context.Context, spaceId int64, args DeleteRolePermissionArgs) error
}

type CreateRolePermissionArgs struct {
	RoleId       int64
	PermissionId int64
	By           *int64
}

// DeleteRolePermissionArgs requires at least one selector; two selectors use OR.
type DeleteRolePermissionArgs struct {
	RoleId       *int64
	PermissionId *int64
}

type RolePermissionDetails struct {
	RolePermission domainmodelsentity.RolePermission
	Role           domainmodelsentity.Role
	Permission     domainmodelsentity.Permission
}

type RolePermissionWithRole struct {
	RolePermission domainmodelsentity.RolePermission
	Role           domainmodelsentity.Role
}

type RolePermissionWithPermission struct {
	RolePermission domainmodelsentity.RolePermission
	Permission     domainmodelsentity.Permission
}
