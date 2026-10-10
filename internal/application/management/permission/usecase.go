package applicationmanagementpermission

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

const tagPath = "management/permission"

type usecase struct {
	user        domaincontractsrepository.User
	space       domaincontractsrepository.Space
	spaceMember domaincontractsrepository.SpaceMember
	permission  domaincontractsrepository.Permission
	guard       domaincontractsutility.PermissionGuard
	validator   domaincontractsutility.Validator
	logger      domaincontractsutility.Logger
	transactor  domaincontractsutility.Transactor
	identity    *applicationshared.IdentityReader
}

func NewUsecaseImpl(
	user domaincontractsrepository.User,
	space domaincontractsrepository.Space,
	spaceMember domaincontractsrepository.SpaceMember,
	permission domaincontractsrepository.Permission,
	guard domaincontractsutility.PermissionGuard,
	validator domaincontractsutility.Validator,
	logger domaincontractsutility.Logger,
	transactor domaincontractsutility.Transactor,
	controlSpaceSlug string,
) domainusecasesmanagement.Permission {
	return &usecase{
		user:        user,
		space:       space,
		spaceMember: spaceMember,
		permission:  permission,
		guard:       guard,
		validator:   validator,
		logger:      logger,
		transactor:  transactor,
		identity:    applicationshared.NewIdentityReader(user, space, spaceMember, controlSpaceSlug),
	}
}

func (u *usecase) Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.CreatePermissionRequest) (result int64, err error) {
	const tag = tagPath + "/Create"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementPermissionCreate}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to create permission", domainmodelslogger.Meta{
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

		if err := u.validator.PermissionSlug(request.Slug); err != nil {
			u.logger.Error(ctx, tag, "invalid permission slug", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		if err := u.validator.PermissionName(request.Name); err != nil {
			u.logger.Error(ctx, tag, "invalid permission name", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		if request.Description != nil {
			if err := u.validator.PermissionDesc(*request.Description); err != nil {
				u.logger.Error(ctx, tag, "invalid permission description", domainmodelslogger.Meta{
					"caller_user_id":         callerUserId,
					"caller_space_id":        callerSpaceId,
					"caller_space_member_id": callerSpaceMemberId,
					"err":                    err,
					"space_id":               spaceId,
				})

				return err
			}
		}

		if err := applicationshared.ValidateJSON(request.Preferences); err != nil {
			u.logger.Error(ctx, tag, "invalid permission preferences JSON", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		value, err := u.permission.Create(ctx, spaceId, domaincontractsrepository.CreatePermissionArgs{
			Slug:        request.Slug,
			Name:        request.Name,
			Description: request.Description,
			Preferences: request.Preferences,
			By:          &caller.UserId,
		})
		if err != nil {
			u.logger.Error(ctx, tag, "failed to create permission", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete permission creation transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return 0, err
	}

	u.logger.Info(ctx, tag, "created permission", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"space_id":               spaceId,
		"id":                     result,
	})

	return result, nil
}

func (u *usecase) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) (result *domainmodelsentity.Permission, err error) {
	const tag = tagPath + "/ReadById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementPermissionReadById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read permission by id", domainmodelslogger.Meta{
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

	value, err := u.permission.ReadById(ctx, spaceId, id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read permission by id", domainmodelslogger.Meta{
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

func (u *usecase) ReadBySlug(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, selector string) (result *domainmodelsentity.Permission, err error) {
	const tag = tagPath + "/ReadBySlug"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementPermissionReadBySlug}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read permission by slug", domainmodelslogger.Meta{
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
		})

		return nil, err
	}

	value, err := u.permission.ReadBySlug(ctx, spaceId, selector)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read permission by slug", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return nil, err
	}

	result = value

	return result, nil
}

func (u *usecase) ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.FilterPermissionRequest) (result domainusecasesshared.Page[domainmodelsentity.Permission], err error) {
	const tag = tagPath + "/ReadByFilter"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementPermissionReadByFilter}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to list permissions", domainmodelslogger.Meta{
			"err": err,
		})

		return domainusecasesshared.Page[domainmodelsentity.Permission]{}, err
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

		return domainusecasesshared.Page[domainmodelsentity.Permission]{}, err
	}

	if err := applicationshared.ValidatePagination(request.Page, request.Limit); err != nil {
		u.logger.Error(ctx, tag, "invalid permission pagination", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return domainusecasesshared.Page[domainmodelsentity.Permission]{}, err
	}

	items, total, err := u.permission.ReadByFilter(ctx, spaceId, domaincontractsrepository.PermissionFilter{
		Page:   request.Page,
		Limit:  request.Limit,
		Search: request.Search,
	})
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list permissions", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return domainusecasesshared.Page[domainmodelsentity.Permission]{}, err
	}

	if items == nil {
		items = []domainmodelsentity.Permission{}
	}

	result = domainusecasesshared.Page[domainmodelsentity.Permission]{Items: items, Total: total}

	return result, nil
}

func (u *usecase) UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64, request domainusecasesmanagement.UpdatePermissionRequest) (err error) {
	const tag = tagPath + "/UpdateById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementPermissionUpdateById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to update by id permission", domainmodelslogger.Meta{
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

		if request.Name != nil {
			if err := u.validator.PermissionName(*request.Name); err != nil {
				u.logger.Error(ctx, tag, "invalid permission name", domainmodelslogger.Meta{
					"caller_user_id":         callerUserId,
					"caller_space_id":        callerSpaceId,
					"caller_space_member_id": callerSpaceMemberId,
					"err":                    err,
					"space_id":               spaceId,
					"id":                     id,
				})

				return err
			}
		}

		if request.Description != nil {
			if err := u.validator.PermissionDesc(*request.Description); err != nil {
				u.logger.Error(ctx, tag, "invalid permission description", domainmodelslogger.Meta{
					"caller_user_id":         callerUserId,
					"caller_space_id":        callerSpaceId,
					"caller_space_member_id": callerSpaceMemberId,
					"err":                    err,
					"space_id":               spaceId,
					"id":                     id,
				})

				return err
			}
		}

		if err := applicationshared.ValidateJSON(request.Preferences); err != nil {
			u.logger.Error(ctx, tag, "invalid permission preferences JSON", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
				"id":                     id,
			})

			return err
		}

		if err := u.permission.UpdateById(ctx, spaceId, id, domaincontractsrepository.UpdatePermissionArgs{
			Name:        request.Name,
			Description: request.Description,
			Preferences: request.Preferences,
			By:          &caller.UserId,
		}); err != nil {
			u.logger.Error(ctx, tag, "failed to update permission", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete permission update transaction", domainmodelslogger.Meta{
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
}

func (u *usecase) DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) (err error) {
	const tag = tagPath + "/DeleteById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementPermissionDeleteById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to delete by id permission", domainmodelslogger.Meta{
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

		if err := u.permission.DeleteById(ctx, spaceId, id, &caller.UserId); err != nil {
			u.logger.Error(ctx, tag, "failed to delete permission by id", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete permission deletion transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"id":                     id,
		})

		return err
	}

	u.logger.Info(ctx, tag, "deleted permission", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"space_id":               spaceId,
		"id":                     id,
	})

	return nil
}
