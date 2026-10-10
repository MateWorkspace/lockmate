package domainusecasesmanagement

import (
	"context"
	"encoding/json"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

// SpaceMember defines permission-protected administration; audit actors come from caller.
// Scoped operations require target-space or control-space authority.
type SpaceMember interface {
	Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request CreateSpaceMemberRequest) (id int64, err error)
	ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) (*domainmodelsentity.SpaceMember, error)
	ReadByUserId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, value int64) (*domainmodelsentity.SpaceMember, error)
	ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request FilterSpaceMemberRequest) (domainusecasesshared.Page[domaincontractsrepository.SpaceMemberWithUser], error)
	UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64, request UpdateSpaceMemberRequest) error
	ReadEffectiveRolesBySpaceMemberId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, spaceMemberId int64) ([]domainmodelsentity.Role, error)
	ReadEffectivePermissionsBySpaceMemberId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, spaceMemberId int64) ([]domainmodelsentity.Permission, error)
	DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) error
}

type CreateSpaceMemberRequest struct {
	UserId      int64
	IsActive    *bool
	Preferences *json.RawMessage
}

type FilterSpaceMemberRequest struct {
	Page     int64
	Limit    int64
	Search   *string
	UserId   *int64
	RoleId   *int64
	IsActive *bool
}

// UpdateSpaceMemberRequest leaves omitted fields unchanged.
type UpdateSpaceMemberRequest struct {
	IsActive    *bool
	Preferences *json.RawMessage
}
