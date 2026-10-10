package applicationmanagementspacememberrole

import (
	"context"

	applicationpermissions "github.com/MateWorkspace/lockmate/internal/application/permissions"
	applicationshared "github.com/MateWorkspace/lockmate/internal/application/shared"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
)

const tagPath = "management/space_member_role"

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
) domainusecasesmanagement.SpaceMemberRole {
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

func (u *usecase) Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.CreateSpaceMemberRoleRequest) (result int64, err error) {
	const tag = tagPath + "/Create"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceMemberRoleCreate}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to create space member role", domainmodelslogger.Meta{
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

		if err := applicationshared.ValidateIds(request.SpaceMemberId); err != nil {
			u.logger.Error(ctx, tag, "invalid space member id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		if err := applicationshared.ValidateIds(request.RoleId); err != nil {
			u.logger.Error(ctx, tag, "invalid role id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		value, err := u.spaceMemberRole.Create(ctx, spaceId, domaincontractsrepository.CreateSpaceMemberRoleArgs{
			SpaceMemberId: request.SpaceMemberId,
			RoleId:        request.RoleId,
			By:            &caller.UserId,
		})
		if err != nil {
			u.logger.Error(ctx, tag, "failed to create space member role", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete space member role creation transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return 0, err
	}

	u.logger.Info(ctx, tag, "created space member role", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"space_id":               spaceId,
		"id":                     result,
	})

	return result, nil
}

func (u *usecase) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) (result *domaincontractsrepository.SpaceMemberRoleDetails, err error) {
	const tag = tagPath + "/ReadById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceMemberRoleReadById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read space member role by id", domainmodelslogger.Meta{
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

	value, err := u.spaceMemberRole.ReadById(ctx, spaceId, id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read space member role by id", domainmodelslogger.Meta{
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

func (u *usecase) ReadBySpaceMemberIdAndRoleId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, spaceMemberId, roleId int64) (result *domaincontractsrepository.SpaceMemberRoleDetails, err error) {
	const tag = tagPath + "/ReadBySpaceMemberIdAndRoleId"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceMemberRoleReadBySpaceMemberIdAndRoleId}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read space member role by space member id and role id", domainmodelslogger.Meta{
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
			"role_id":                roleId,
		})

		return nil, err
	}

	if err := applicationshared.ValidateIds(spaceMemberId, roleId); err != nil {
		u.logger.Error(ctx, tag, "invalid space member id or role id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"space_member_id":        spaceMemberId,
			"role_id":                roleId,
		})

		return nil, err
	}

	value, err := u.spaceMemberRole.ReadBySpaceMemberIdAndRoleId(ctx, spaceId, spaceMemberId, roleId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read space member role by space member id and role id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"space_member_id":        spaceMemberId,
			"role_id":                roleId,
		})

		return nil, err
	}

	result = value

	return result, nil
}

func (u *usecase) ReadBySpaceMemberId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, selector int64) (result []domaincontractsrepository.SpaceMemberRoleWithRole, err error) {
	const tag = tagPath + "/ReadBySpaceMemberId"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceMemberRoleReadBySpaceMemberId}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read space member role by space member id", domainmodelslogger.Meta{
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
			"space_member_id":        selector,
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
			"space_member_id":        selector,
		})

		return nil, err
	}

	value, err := u.spaceMemberRole.ReadBySpaceMemberId(ctx, spaceId, selector)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read space member role by space member id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"space_member_id":        selector,
		})

		return nil, err
	}

	result = value
	if result == nil {
		result = []domaincontractsrepository.SpaceMemberRoleWithRole{}
	}

	return result, nil
}

func (u *usecase) ReadByRoleId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, selector int64) (result []domaincontractsrepository.SpaceMemberRoleWithSpaceMember, err error) {
	const tag = tagPath + "/ReadByRoleId"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceMemberRoleReadByRoleId}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read space member role by role id", domainmodelslogger.Meta{
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
			"role_id":                selector,
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
			"role_id":                selector,
		})

		return nil, err
	}

	value, err := u.spaceMemberRole.ReadByRoleId(ctx, spaceId, selector)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read space member role by role id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"role_id":                selector,
		})

		return nil, err
	}

	result = value
	if result == nil {
		result = []domaincontractsrepository.SpaceMemberRoleWithSpaceMember{}
	}

	return result, nil
}

func (u *usecase) DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) (err error) {
	const tag = tagPath + "/DeleteById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceMemberRoleDeleteById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to delete by id space member role", domainmodelslogger.Meta{
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

		if err := u.spaceMemberRole.DeleteById(ctx, spaceId, id); err != nil {
			u.logger.Error(ctx, tag, "failed to delete space member role by id", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete space member role deletion transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"id":                     id,
		})

		return err
	}

	u.logger.Info(ctx, tag, "deleted space member role", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"space_id":               spaceId,
		"id":                     id,
	})

	return nil
}

func (u *usecase) DeleteBySpaceMemberIdOrRoleId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.DeleteSpaceMemberRoleRequest) (err error) {
	const tag = tagPath + "/DeleteBySpaceMemberIdOrRoleId"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceMemberRoleDeleteBySpaceMemberIdOrRoleId}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to delete by space member id or role id space member role", domainmodelslogger.Meta{
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
			})

			return err
		}

		if err := applicationshared.ValidateOptionalId(request.SpaceMemberId); err != nil {
			u.logger.Error(ctx, tag, "invalid space member id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		if err := applicationshared.ValidateOptionalId(request.RoleId); err != nil {
			u.logger.Error(ctx, tag, "invalid role id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		if request.SpaceMemberId == nil && request.RoleId == nil {
			failure := domainmodelserror.NewError("a selector is required", domainmodelserror.BadArgs, nil)
			u.logger.Error(ctx, tag, "a selector is required", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    failure,
				"space_id":               spaceId,
			})

			return failure
		}

		if err := u.spaceMemberRole.DeleteBySpaceMemberIdOrRoleId(ctx, spaceId, domaincontractsrepository.DeleteSpaceMemberRoleArgs{
			SpaceMemberId: request.SpaceMemberId,
			RoleId:        request.RoleId,
		}); err != nil {
			u.logger.Error(ctx, tag, "failed to delete space member role by space member id or role id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete space member role bulk deletion transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return err
	}

	u.logger.Info(ctx, tag, "deleted assignments for space member role", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"space_id":               spaceId,
	})

	return nil
}
