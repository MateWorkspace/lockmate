package applicationshared

import (
	"context"
	"errors"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
)

func ResolveSession(ctx context.Context, identity *IdentityReader, grants domaincontractsrepository.SpaceMemberRole, userId, spaceId, memberId int64) (*domainusecasesauth.SessionResult, error) {
	live, err := identity.Read(ctx, &domainmodelsentity.AccessClaims{UserId: userId, SpaceId: spaceId, SpaceMemberId: memberId})
	if err != nil {
		return nil, err
	}

	roles, err := grants.ReadEffectiveRolesBySpaceMemberId(ctx, spaceId, memberId)
	if err != nil {
		return nil, IdentityError(err)
	}

	permissions, err := grants.ReadEffectivePermissionsBySpaceMemberId(ctx, spaceId, memberId)
	if err != nil {
		return nil, IdentityError(err)
	}

	if roles == nil {
		roles = []domainmodelsentity.Role{}
	}

	if permissions == nil {
		permissions = []domainmodelsentity.Permission{}
	}

	return &domainusecasesauth.SessionResult{
		User:        live.User,
		Space:       live.Space,
		SpaceMember: live.SpaceMember,
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

func SessionClaims(result *domainusecasesauth.SessionResult) *domainmodelsentity.AccessClaims {
	roles := make([]string, 0, len(result.Roles))
	permissions := make([]string, 0, len(result.Permissions))
	for _, role := range result.Roles {
		roles = append(roles, role.Slug)
	}

	for _, permission := range result.Permissions {
		permissions = append(permissions, permission.Slug)
	}

	return &domainmodelsentity.AccessClaims{
		UserId:        result.User.Id,
		SpaceId:       result.Space.Id,
		SpaceMemberId: result.SpaceMember.Id,
		Name:          result.User.Name,
		Roles:         roles,
		Permissions:   permissions,
	}
}

func EnsureMembership(ctx context.Context, members domaincontractsrepository.SpaceMember, userId, spaceId int64, by *int64) (memberId int64, created bool, err error) {
	member, err := members.ReadByUserId(ctx, spaceId, userId)
	if errors.Is(err, domainmodelserror.SpaceMemberNotFound) {
		id, err := members.Create(ctx, spaceId, domaincontractsrepository.CreateSpaceMemberArgs{UserId: userId, By: by})
		return id, err == nil, err
	}

	if err != nil {
		return 0, false, err
	}

	if member == nil || member.UserId != userId || member.SpaceId != spaceId || member.Id <= 0 || !member.IsActive || member.AuditDelete.At != nil {
		return 0, false, Unauthorized(nil)
	}

	return member.Id, false, nil
}
