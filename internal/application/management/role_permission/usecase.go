package applicationmanagementrolepermission

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

const tagPath = "management/role_permission"

type usecase struct {
	user           domaincontractsrepository.User
	space          domaincontractsrepository.Space
	spaceMember    domaincontractsrepository.SpaceMember
	rolePermission domaincontractsrepository.RolePermission
	guard          domaincontractsutility.PermissionGuard
	logger         domaincontractsutility.Logger
	transactor     domaincontractsutility.Transactor
	identity       *applicationshared.IdentityReader
}

func NewUsecaseImpl(
	user domaincontractsrepository.User,
	space domaincontractsrepository.Space,
	spaceMember domaincontractsrepository.SpaceMember,
	rolePermission domaincontractsrepository.RolePermission,
	guard domaincontractsutility.PermissionGuard,
	logger domaincontractsutility.Logger,
	transactor domaincontractsutility.Transactor,
	controlSpaceSlug string,
) domainusecasesmanagement.RolePermission {
	return &usecase{
		user:           user,
		space:          space,
		spaceMember:    spaceMember,
		rolePermission: rolePermission,
		guard:          guard,
		logger:         logger,
		transactor:     transactor,
		identity:       applicationshared.NewIdentityReader(user, space, spaceMember, controlSpaceSlug),
	}
}

func (u *usecase) Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.CreateRolePermissionRequest) (result int64, err error) {
	const tag = tagPath + "/Create"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementRolePermissionCreate}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to create role permission", domainmodelslogger.Meta{
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

		if err := applicationshared.ValidateIds(request.PermissionId); err != nil {
			u.logger.Error(ctx, tag, "invalid permission id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		value, err := u.rolePermission.Create(ctx, spaceId, domaincontractsrepository.CreateRolePermissionArgs{
			RoleId:       request.RoleId,
			PermissionId: request.PermissionId,
			By:           &caller.UserId,
		})
		if err != nil {
			u.logger.Error(ctx, tag, "failed to create role permission", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete role permission creation transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return 0, err
	}

	u.logger.Info(ctx, tag, "created role permission", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"space_id":               spaceId,
		"id":                     result,
	})

	return result, nil
}

func (u *usecase) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) (result *domaincontractsrepository.RolePermissionDetails, err error) {
	const tag = tagPath + "/ReadById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementRolePermissionReadById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read role permission by id", domainmodelslogger.Meta{
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

	value, err := u.rolePermission.ReadById(ctx, spaceId, id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read role permission by id", domainmodelslogger.Meta{
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

func (u *usecase) ReadByRoleIdAndPermissionId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, roleId, permissionId int64) (result *domaincontractsrepository.RolePermissionDetails, err error) {
	const tag = tagPath + "/ReadByRoleIdAndPermissionId"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementRolePermissionReadByRoleIdAndPermissionId}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read role permission by role id and permission id", domainmodelslogger.Meta{
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
			"role_id":                roleId,
			"permission_id":          permissionId,
		})

		return nil, err
	}

	if err := applicationshared.ValidateIds(roleId, permissionId); err != nil {
		u.logger.Error(ctx, tag, "invalid role id or permission id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"role_id":                roleId,
			"permission_id":          permissionId,
		})

		return nil, err
	}

	value, err := u.rolePermission.ReadByRoleIdAndPermissionId(ctx, spaceId, roleId, permissionId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read role permission by role id and permission id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"role_id":                roleId,
			"permission_id":          permissionId,
		})

		return nil, err
	}

	result = value

	return result, nil
}

func (u *usecase) ReadByRoleId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, selector int64) (result []domaincontractsrepository.RolePermissionWithPermission, err error) {
	const tag = tagPath + "/ReadByRoleId"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementRolePermissionReadByRoleId}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read role permission by role id", domainmodelslogger.Meta{
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

	value, err := u.rolePermission.ReadByRoleId(ctx, spaceId, selector)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read role permission by role id", domainmodelslogger.Meta{
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
		result = []domaincontractsrepository.RolePermissionWithPermission{}
	}

	return result, nil
}

func (u *usecase) ReadByPermissionId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, selector int64) (result []domaincontractsrepository.RolePermissionWithRole, err error) {
	const tag = tagPath + "/ReadByPermissionId"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementRolePermissionReadByPermissionId}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read role permission by permission id", domainmodelslogger.Meta{
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
			"permission_id":          selector,
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
			"permission_id":          selector,
		})

		return nil, err
	}

	value, err := u.rolePermission.ReadByPermissionId(ctx, spaceId, selector)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read role permission by permission id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"permission_id":          selector,
		})

		return nil, err
	}

	result = value
	if result == nil {
		result = []domaincontractsrepository.RolePermissionWithRole{}
	}

	return result, nil
}

func (u *usecase) DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) (err error) {
	const tag = tagPath + "/DeleteById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementRolePermissionDeleteById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to delete by id role permission", domainmodelslogger.Meta{
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

		if err := u.rolePermission.DeleteById(ctx, spaceId, id); err != nil {
			u.logger.Error(ctx, tag, "failed to delete role permission by id", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete role permission deletion transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"id":                     id,
		})

		return err
	}

	u.logger.Info(ctx, tag, "deleted role permission", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"space_id":               spaceId,
		"id":                     id,
	})

	return nil
}

func (u *usecase) DeleteByRoleIdOrPermissionId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.DeleteRolePermissionRequest) (err error) {
	const tag = tagPath + "/DeleteByRoleIdOrPermissionId"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementRolePermissionDeleteByRoleIdOrPermissionId}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to delete by role id or permission id role permission", domainmodelslogger.Meta{
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

		if err := applicationshared.ValidateOptionalId(request.PermissionId); err != nil {
			u.logger.Error(ctx, tag, "invalid permission id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		if request.RoleId == nil && request.PermissionId == nil {
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

		if err := u.rolePermission.DeleteByRoleIdOrPermissionId(ctx, spaceId, domaincontractsrepository.DeleteRolePermissionArgs{
			RoleId:       request.RoleId,
			PermissionId: request.PermissionId,
		}); err != nil {
			u.logger.Error(ctx, tag, "failed to delete role permission by role id or permission id", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete role permission bulk deletion transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return err
	}

	u.logger.Info(ctx, tag, "deleted assignments for role permission", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"space_id":               spaceId,
	})

	return nil
}
