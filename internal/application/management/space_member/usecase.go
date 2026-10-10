package applicationmanagementspacemember

import (
	"context"

	applicationpermissions "github.com/MateWorkspace/lockmate/internal/application/permissions"
	applicationshared "github.com/MateWorkspace/lockmate/internal/application/shared"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

const tagPath = "management/space_member"

type usecase struct {
	user            domaincontractsrepository.User
	space           domaincontractsrepository.Space
	spaceMember     domaincontractsrepository.SpaceMember
	spaceMemberRole domaincontractsrepository.SpaceMemberRole
	guard           domaincontractsutility.PermissionGuard
	logger          domaincontractsutility.Logger
	transactor      domaincontractsutility.Transactor
	identity        *applicationshared.IdentityReader
}

func NewUsecaseImpl(
	user domaincontractsrepository.User,
	space domaincontractsrepository.Space,
	spaceMember domaincontractsrepository.SpaceMember,
	spaceMemberRole domaincontractsrepository.SpaceMemberRole,
	guard domaincontractsutility.PermissionGuard,
	logger domaincontractsutility.Logger,
	transactor domaincontractsutility.Transactor,
	controlSpaceSlug string,
) domainusecasesmanagement.SpaceMember {
	return &usecase{
		user:            user,
		space:           space,
		spaceMember:     spaceMember,
		spaceMemberRole: spaceMemberRole,
		guard:           guard,
		logger:          logger,
		transactor:      transactor,
		identity:        applicationshared.NewIdentityReader(user, space, spaceMember, controlSpaceSlug),
	}
}

func (u *usecase) Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.CreateSpaceMemberRequest) (result int64, err error) {
	const tag = tagPath + "/Create"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceMemberCreate}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to create space member", domainmodelslogger.Meta{
			"err": err,
		})

		return 0, err
	}

	callerUserId, callerSpaceId, callerSpaceMemberId := int64(0), int64(0), int64(0)
	if caller != nil {
		callerUserId = caller.UserId
		callerSpaceId = caller.SpaceId
		callerSpaceMemberId = caller.SpaceMemberId
	}

	err = u.transactor.WithTx(ctx, func(ctx context.Context) error {
		if _, err := u.identity.Management(ctx, caller, &spaceId); err != nil {
			u.logger.Error(ctx, tag, "failed to authorize management scope", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		if err := applicationshared.ValidateIds(request.UserId); err != nil {
			u.logger.Error(ctx, tag, "invalid user id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		if err := applicationshared.ValidateJSON(request.Preferences); err != nil {
			u.logger.Error(ctx, tag, "invalid space member preferences JSON", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		value, err := u.spaceMember.Create(ctx, spaceId, domaincontractsrepository.CreateSpaceMemberArgs{
			UserId:      request.UserId,
			IsActive:    request.IsActive,
			Preferences: request.Preferences,
			By:          &caller.UserId,
		})
		if err != nil {
			u.logger.Error(ctx, tag, "failed to create space member", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		result = value
		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete space member creation transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return 0, err
	}

	u.logger.Info(ctx, tag, "created space member", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"space_id":               spaceId,
		"id":                     result,
	})

	return result, nil
}

func (u *usecase) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) (result *domainmodelsentity.SpaceMember, err error) {
	const tag = tagPath + "/ReadById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceMemberReadById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read space member by id", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	callerUserId, callerSpaceId, callerSpaceMemberId := int64(0), int64(0), int64(0)
	if caller != nil {
		callerUserId = caller.UserId
		callerSpaceId = caller.SpaceId
		callerSpaceMemberId = caller.SpaceMemberId
	}

	if _, err := u.identity.Management(ctx, caller, &spaceId); err != nil {
		u.logger.Error(ctx, tag, "failed to authorize management scope", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"id":                     id,
		})

		return nil, err
	}

	if err := applicationshared.ValidateIds(id); err != nil {
		u.logger.Error(ctx, tag, "invalid id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"id":                     id,
		})

		return nil, err
	}

	value, err := u.spaceMember.ReadById(ctx, spaceId, id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read space member by id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"id":                     id,
		})

		return nil, err
	}

	result = value

	return result, nil
}

func (u *usecase) ReadByUserId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, selector int64) (result *domainmodelsentity.SpaceMember, err error) {
	const tag = tagPath + "/ReadByUserId"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceMemberReadByUserId}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read space member by user id", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	callerUserId, callerSpaceId, callerSpaceMemberId := int64(0), int64(0), int64(0)
	if caller != nil {
		callerUserId = caller.UserId
		callerSpaceId = caller.SpaceId
		callerSpaceMemberId = caller.SpaceMemberId
	}

	if _, err := u.identity.Management(ctx, caller, &spaceId); err != nil {
		u.logger.Error(ctx, tag, "failed to authorize management scope", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"user_id":                selector,
		})

		return nil, err
	}

	if err := applicationshared.ValidateIds(selector); err != nil {
		u.logger.Error(ctx, tag, "invalid selector", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"user_id":                selector,
		})

		return nil, err
	}

	value, err := u.spaceMember.ReadByUserId(ctx, spaceId, selector)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read space member by user id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"user_id":                selector,
		})

		return nil, err
	}

	result = value

	return result, nil
}

func (u *usecase) ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.FilterSpaceMemberRequest) (result domainusecasesshared.Page[domaincontractsrepository.SpaceMemberWithUser], err error) {
	const tag = tagPath + "/ReadByFilter"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceMemberReadByFilter}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to list space members", domainmodelslogger.Meta{
			"err": err,
		})

		return domainusecasesshared.Page[domaincontractsrepository.SpaceMemberWithUser]{}, err
	}

	callerUserId, callerSpaceId, callerSpaceMemberId := int64(0), int64(0), int64(0)
	if caller != nil {
		callerUserId = caller.UserId
		callerSpaceId = caller.SpaceId
		callerSpaceMemberId = caller.SpaceMemberId
	}

	if _, err := u.identity.Management(ctx, caller, &spaceId); err != nil {
		u.logger.Error(ctx, tag, "failed to authorize management scope", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return domainusecasesshared.Page[domaincontractsrepository.SpaceMemberWithUser]{}, err
	}

	if err := applicationshared.ValidateOptionalId(request.UserId); err != nil {
		u.logger.Error(ctx, tag, "invalid user id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return domainusecasesshared.Page[domaincontractsrepository.SpaceMemberWithUser]{}, err
	}

	if err := applicationshared.ValidateOptionalId(request.RoleId); err != nil {
		u.logger.Error(ctx, tag, "invalid role id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return domainusecasesshared.Page[domaincontractsrepository.SpaceMemberWithUser]{}, err
	}

	if err := applicationshared.ValidatePagination(request.Page, request.Limit); err != nil {
		u.logger.Error(ctx, tag, "invalid space member pagination", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return domainusecasesshared.Page[domaincontractsrepository.SpaceMemberWithUser]{}, err
	}

	items, total, err := u.spaceMember.ReadByFilter(ctx, spaceId, domaincontractsrepository.SpaceMemberFilter{
		Page:     request.Page,
		Limit:    request.Limit,
		Search:   request.Search,
		UserId:   request.UserId,
		RoleId:   request.RoleId,
		IsActive: request.IsActive,
	})
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list space members", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return domainusecasesshared.Page[domaincontractsrepository.SpaceMemberWithUser]{}, err
	}

	if items == nil {
		items = []domaincontractsrepository.SpaceMemberWithUser{}
	}

	result = domainusecasesshared.Page[domaincontractsrepository.SpaceMemberWithUser]{Items: items, Total: total}

	return result, nil
}

func (u *usecase) UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64, request domainusecasesmanagement.UpdateSpaceMemberRequest) (err error) {
	const tag = tagPath + "/UpdateById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceMemberUpdateById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to update by id space member", domainmodelslogger.Meta{
			"err": err,
		})

		return err
	}

	callerUserId, callerSpaceId, callerSpaceMemberId := int64(0), int64(0), int64(0)
	if caller != nil {
		callerUserId = caller.UserId
		callerSpaceId = caller.SpaceId
		callerSpaceMemberId = caller.SpaceMemberId
	}

	err = u.transactor.WithTx(ctx, func(ctx context.Context) error {
		if _, err := u.identity.Management(ctx, caller, &spaceId); err != nil {
			u.logger.Error(ctx, tag, "failed to authorize management scope", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
				"id":                     id,
			})

			return err
		}

		if err := applicationshared.ValidateIds(id); err != nil {
			u.logger.Error(ctx, tag, "invalid id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
				"id":                     id,
			})

			return err
		}

		if err := applicationshared.ValidateJSON(request.Preferences); err != nil {
			u.logger.Error(ctx, tag, "invalid space member preferences JSON", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
				"id":                     id,
			})

			return err
		}

		if err := u.spaceMember.UpdateById(ctx, spaceId, id, domaincontractsrepository.UpdateSpaceMemberArgs{
			IsActive:    request.IsActive,
			Preferences: request.Preferences,
			By:          &caller.UserId,
		}); err != nil {
			u.logger.Error(ctx, tag, "failed to update space member", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
				"id":                     id,
			})

			return err
		}

		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete space member update transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"id":                     id,
		})

		return err
	}

	if request.IsActive != nil {
		u.logger.Info(ctx, tag, "updated space member", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"id":                     id,
		})
	}

	return nil
}

func (u *usecase) ReadEffectiveRolesBySpaceMemberId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, spaceMemberId int64) (result []domainmodelsentity.Role, err error) {
	const tag = tagPath + "/ReadEffectiveRolesBySpaceMemberId"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceMemberReadEffectiveRolesBySpaceMemberId}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read effective roles by space member id space member", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	callerUserId, callerSpaceId, callerSpaceMemberId := int64(0), int64(0), int64(0)
	if caller != nil {
		callerUserId = caller.UserId
		callerSpaceId = caller.SpaceId
		callerSpaceMemberId = caller.SpaceMemberId
	}

	if _, err := u.identity.Management(ctx, caller, &spaceId); err != nil {
		u.logger.Error(ctx, tag, "failed to authorize management scope", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"space_member_id":        spaceMemberId,
		})

		return nil, err
	}

	if err := applicationshared.ValidateIds(spaceMemberId); err != nil {
		u.logger.Error(ctx, tag, "invalid space member id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"space_member_id":        spaceMemberId,
		})

		return nil, err
	}

	value, err := u.spaceMemberRole.ReadEffectiveRolesBySpaceMemberId(ctx, spaceId, spaceMemberId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read effective space member roles", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"space_member_id":        spaceMemberId,
		})

		return nil, err
	}

	result = value
	if result == nil {
		result = []domainmodelsentity.Role{}
	}

	return result, nil
}

func (u *usecase) ReadEffectivePermissionsBySpaceMemberId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, spaceMemberId int64) (result []domainmodelsentity.Permission, err error) {
	const tag = tagPath + "/ReadEffectivePermissionsBySpaceMemberId"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceMemberReadEffectivePermissionsBySpaceMemberId}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read effective permissions by space member id space member", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	callerUserId, callerSpaceId, callerSpaceMemberId := int64(0), int64(0), int64(0)
	if caller != nil {
		callerUserId = caller.UserId
		callerSpaceId = caller.SpaceId
		callerSpaceMemberId = caller.SpaceMemberId
	}

	if _, err := u.identity.Management(ctx, caller, &spaceId); err != nil {
		u.logger.Error(ctx, tag, "failed to authorize management scope", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"space_member_id":        spaceMemberId,
		})

		return nil, err
	}

	if err := applicationshared.ValidateIds(spaceMemberId); err != nil {
		u.logger.Error(ctx, tag, "invalid space member id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"space_member_id":        spaceMemberId,
		})

		return nil, err
	}

	value, err := u.spaceMemberRole.ReadEffectivePermissionsBySpaceMemberId(ctx, spaceId, spaceMemberId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read effective space member permissions", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"space_member_id":        spaceMemberId,
		})

		return nil, err
	}

	result = value
	if result == nil {
		result = []domainmodelsentity.Permission{}
	}

	return result, nil
}

func (u *usecase) DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) (err error) {
	const tag = tagPath + "/DeleteById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceMemberDeleteById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to delete by id space member", domainmodelslogger.Meta{
			"err": err,
		})

		return err
	}

	callerUserId, callerSpaceId, callerSpaceMemberId := int64(0), int64(0), int64(0)
	if caller != nil {
		callerUserId = caller.UserId
		callerSpaceId = caller.SpaceId
		callerSpaceMemberId = caller.SpaceMemberId
	}

	err = u.transactor.WithTx(ctx, func(ctx context.Context) error {
		if _, err := u.identity.Management(ctx, caller, &spaceId); err != nil {
			u.logger.Error(ctx, tag, "failed to authorize management scope", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
				"id":                     id,
			})

			return err
		}

		if err := applicationshared.ValidateIds(id); err != nil {
			u.logger.Error(ctx, tag, "invalid id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
				"id":                     id,
			})

			return err
		}

		if err := u.spaceMember.DeleteById(ctx, spaceId, id, &caller.UserId); err != nil {
			u.logger.Error(ctx, tag, "failed to delete space member by id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
				"id":                     id,
			})

			return err
		}

		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete space member deletion transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"id":                     id,
		})

		return err
	}

	u.logger.Info(ctx, tag, "deleted space member", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"space_id":               spaceId,
		"id":                     id,
	})

	return nil
}
