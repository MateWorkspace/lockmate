package domainusecasesmanagement

import (
	"context"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

// SpaceMemberRole defines permission-protected administration; audit actors come from caller.
// Scoped operations require target-space or control-space authority.
type SpaceMemberRole interface {
	Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request CreateSpaceMemberRoleRequest) (id int64, err error)
	ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) (*domaincontractsrepository.SpaceMemberRoleDetails, error)
	ReadBySpaceMemberIdAndRoleId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, spaceMemberId, roleId int64) (*domaincontractsrepository.SpaceMemberRoleDetails, error)
	ReadBySpaceMemberId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, value int64) ([]domaincontractsrepository.SpaceMemberRoleWithRole, error)
	ReadByRoleId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, value int64) ([]domaincontractsrepository.SpaceMemberRoleWithSpaceMember, error)
	DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) error
	DeleteBySpaceMemberIdOrRoleId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request DeleteSpaceMemberRoleRequest) error
}

type CreateSpaceMemberRoleRequest struct {
	SpaceMemberId int64
	RoleId        int64
}

// DeleteSpaceMemberRoleRequest requires at least one selector; two selectors use OR within the space.
type DeleteSpaceMemberRoleRequest struct {
	SpaceMemberId *int64
	RoleId        *int64
}
