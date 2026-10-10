package applicationshared

import (
	"context"
	"errors"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
)

type IdentityReader struct {
	user             domaincontractsrepository.User
	space            domaincontractsrepository.Space
	member           domaincontractsrepository.SpaceMember
	controlSpaceSlug string
}

type Identity struct {
	User        domainmodelsentity.User
	Space       domainmodelsentity.Space
	SpaceMember domainmodelsentity.SpaceMember
}

func NewIdentityReader(user domaincontractsrepository.User, space domaincontractsrepository.Space, member domaincontractsrepository.SpaceMember, controlSpaceSlug string) *IdentityReader {
	if controlSpaceSlug == "" {
		controlSpaceSlug = "lockmate"
	}

	return &IdentityReader{user: user, space: space, member: member, controlSpaceSlug: controlSpaceSlug}
}

func Unauthorized(source error) error {
	return domainmodelserror.NewError("unauthorized", domainmodelserror.Unauthorized, source)
}

func IdentityError(err error) error {
	if errors.Is(err, domainmodelserror.UserNotFound) || errors.Is(err, domainmodelserror.SpaceNotFound) || errors.Is(err, domainmodelserror.SpaceMemberNotFound) || errors.Is(err, domainmodelserror.ClientNotFound) {
		return Unauthorized(err)
	}

	return err
}

func (r *IdentityReader) Read(ctx context.Context, caller *domainmodelsentity.AccessClaims) (*Identity, error) {
	if caller == nil || caller.UserId <= 0 || caller.SpaceId <= 0 || caller.SpaceMemberId <= 0 {
		return nil, Unauthorized(nil)
	}

	user, err := r.user.ReadById(ctx, caller.UserId)
	if err != nil {
		return nil, IdentityError(err)
	}

	space, err := r.space.ReadById(ctx, caller.SpaceId)
	if err != nil {
		return nil, IdentityError(err)
	}

	member, err := r.member.ReadById(ctx, caller.SpaceId, caller.SpaceMemberId)
	if err != nil {
		return nil, IdentityError(err)
	}

	if user == nil || space == nil || member == nil || user.Id != caller.UserId || space.Id != caller.SpaceId || member.Id != caller.SpaceMemberId || member.UserId != caller.UserId || member.SpaceId != caller.SpaceId || user.AuditDelete.At != nil || space.AuditDelete.At != nil || member.AuditDelete.At != nil || !space.IsActive || !member.IsActive {
		return nil, Unauthorized(nil)
	}

	return &Identity{User: *user, Space: *space, SpaceMember: *member}, nil
}

// Management checks identity and space boundaries; the feature checks permissions.
func (r *IdentityReader) Management(ctx context.Context, caller *domainmodelsentity.AccessClaims, targetSpaceId *int64) (*Identity, error) {
	if err := ValidateOptionalId(targetSpaceId); err != nil {
		return nil, err
	}

	identity, err := r.Read(ctx, caller)
	if err != nil {
		return nil, err
	}

	if targetSpaceId != nil && *targetSpaceId == identity.Space.Id {
		return identity, nil
	}

	control, err := r.space.ReadBySlug(ctx, r.controlSpaceSlug)
	if err != nil {
		return nil, err
	}

	if control == nil || control.AuditDelete.At != nil || control.Id != identity.Space.Id {
		return nil, domainmodelserror.NewError("forbidden", domainmodelserror.Forbidden, nil)
	}

	if targetSpaceId != nil {
		target, err := r.space.ReadById(ctx, *targetSpaceId)
		if err != nil {
			return nil, err
		}

		if target == nil || target.Id != *targetSpaceId || target.AuditDelete.At != nil {
			return nil, domainmodelserror.NewError("space not found", domainmodelserror.SpaceNotFound, nil)
		}
	}

	return identity, nil
}

func ReadClient(ctx context.Context, clients domaincontractsrepository.Client, spaces domaincontractsrepository.Space, identity *domainmodelsentity.ClientIdentity) (*domainmodelsentity.Client, error) {
	if identity == nil || identity.Id <= 0 || identity.SpaceId <= 0 || identity.Slug == "" {
		return nil, Unauthorized(nil)
	}

	client, err := clients.ReadById(ctx, identity.SpaceId, identity.Id)
	if err != nil {
		return nil, IdentityError(err)
	}

	if client == nil || client.Id != identity.Id || client.SpaceId != identity.SpaceId || client.Slug != identity.Slug || !client.IsActive || client.AuditDelete.At != nil {
		return nil, Unauthorized(nil)
	}

	space, err := spaces.ReadById(ctx, identity.SpaceId)
	if err != nil {
		return nil, IdentityError(err)
	}

	if space == nil || space.Id != identity.SpaceId || !space.IsActive || space.AuditDelete.At != nil {
		return nil, Unauthorized(nil)
	}

	return client, nil
}
