package domaincontractsrepository

import (
	"context"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type SpaceMemberRole interface {
	Create(ctx context.Context, spaceId int64, args CreateSpaceMemberRoleArgs) (id int64, err error)
	ReadById(ctx context.Context, spaceId, id int64) (*SpaceMemberRoleDetails, error)
	ReadBySpaceMemberIdAndRoleId(ctx context.Context, spaceId, spaceMemberId, roleId int64) (*SpaceMemberRoleDetails, error)
	ReadBySpaceMemberId(ctx context.Context, spaceId, value int64) ([]SpaceMemberRoleWithRole, error)
	ReadByRoleId(ctx context.Context, spaceId, value int64) ([]SpaceMemberRoleWithSpaceMember, error)
	ReadEffectiveRolesBySpaceMemberId(ctx context.Context, spaceId, spaceMemberId int64) ([]domainmodelsentity.Role, error)
	ReadEffectivePermissionsBySpaceMemberId(ctx context.Context, spaceId, spaceMemberId int64) ([]domainmodelsentity.Permission, error)
	DeleteById(ctx context.Context, spaceId, id int64) error
	DeleteBySpaceMemberIdOrRoleId(ctx context.Context, spaceId int64, args DeleteSpaceMemberRoleArgs) error
}

type CreateSpaceMemberRoleArgs struct {
	SpaceMemberId int64
	RoleId        int64
	By            *int64
}

// DeleteSpaceMemberRoleArgs requires at least one selector; two selectors use OR.
type DeleteSpaceMemberRoleArgs struct {
	SpaceMemberId *int64
	RoleId        *int64
}

type SpaceMemberRoleDetails struct {
	SpaceMemberRole domainmodelsentity.SpaceMemberRole
	SpaceMember     domainmodelsentity.SpaceMember
	Role            domainmodelsentity.Role
}

type SpaceMemberRoleWithSpaceMember struct {
	SpaceMemberRole domainmodelsentity.SpaceMemberRole
	SpaceMember     domainmodelsentity.SpaceMember
}

type SpaceMemberRoleWithRole struct {
	SpaceMemberRole domainmodelsentity.SpaceMemberRole
	Role            domainmodelsentity.Role
}
