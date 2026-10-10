package applicationmanagementspace

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

const tagPath = "management/space"

type usecase struct {
	user        domaincontractsrepository.User
	space       domaincontractsrepository.Space
	spaceMember domaincontractsrepository.SpaceMember
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
	guard domaincontractsutility.PermissionGuard,
	validator domaincontractsutility.Validator,
	logger domaincontractsutility.Logger,
	transactor domaincontractsutility.Transactor,
	controlSpaceSlug string,
) domainusecasesmanagement.Space {
	return &usecase{
		user:        user,
		space:       space,
		spaceMember: spaceMember,
		guard:       guard,
		validator:   validator,
		logger:      logger,
		transactor:  transactor,
		identity:    applicationshared.NewIdentityReader(user, space, spaceMember, controlSpaceSlug),
	}
}

func (u *usecase) Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesmanagement.CreateSpaceRequest) (result int64, err error) {
	const tag = tagPath + "/Create"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceCreate}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to create space", domainmodelslogger.Meta{
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
		if _, err := u.identity.Management(ctx, caller, nil); err != nil {
			u.logger.Error(ctx, tag, "failed to authorize management scope", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		if err := u.validator.SpaceSlug(request.Slug); err != nil {
			u.logger.Error(ctx, tag, "invalid space slug", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		if err := u.validator.SpaceName(request.Name); err != nil {
			u.logger.Error(ctx, tag, "invalid space name", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		if request.Description != nil {
			if err := u.validator.SpaceDesc(*request.Description); err != nil {
				u.logger.Error(ctx, tag, "invalid space description", domainmodelslogger.Meta{
					"caller_user_id":         callerUserId,
					"caller_space_id":        callerSpaceId,
					"caller_space_member_id": callerSpaceMemberId,
					"err":                    err,
				})

				return err
			}
		}

		if err := applicationshared.ValidateJSON(request.Preferences); err != nil {
			u.logger.Error(ctx, tag, "invalid space preferences JSON", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		value, err := u.space.Create(ctx, domaincontractsrepository.CreateSpaceArgs{
			Slug:        request.Slug,
			Name:        request.Name,
			Description: request.Description,
			IsActive:    request.IsActive,
			Preferences: request.Preferences,
			By:          &caller.UserId,
		})
		if err != nil {
			u.logger.Error(ctx, tag, "failed to create space", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		result = value
		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete space creation transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return 0, err
	}

	u.logger.Info(ctx, tag, "created space", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"id":                     result,
	})

	return result, nil
}

func (u *usecase) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) (result *domainmodelsentity.Space, err error) {
	const tag = tagPath + "/ReadById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceReadById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read space by id", domainmodelslogger.Meta{
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

	if _, err := u.identity.Management(ctx, caller, nil); err != nil {
		u.logger.Error(ctx, tag, "failed to authorize management scope", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
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
			"id":                     id,
		})

		return nil, err
	}

	value, err := u.space.ReadById(ctx, id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read space by id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"id":                     id,
		})

		return nil, err
	}

	result = value

	return result, nil
}

func (u *usecase) ReadBySlug(ctx context.Context, caller *domainmodelsentity.AccessClaims, selector string) (result *domainmodelsentity.Space, err error) {
	const tag = tagPath + "/ReadBySlug"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceReadBySlug}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read space by slug", domainmodelslogger.Meta{
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

	if _, err := u.identity.Management(ctx, caller, nil); err != nil {
		u.logger.Error(ctx, tag, "failed to authorize management scope", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return nil, err
	}

	value, err := u.space.ReadBySlug(ctx, selector)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read space by slug", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return nil, err
	}

	result = value

	return result, nil
}

func (u *usecase) ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesmanagement.FilterSpaceRequest) (result domainusecasesshared.Page[domainmodelsentity.Space], err error) {
	const tag = tagPath + "/ReadByFilter"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceReadByFilter}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to list spaces", domainmodelslogger.Meta{
			"err": err,
		})

		return domainusecasesshared.Page[domainmodelsentity.Space]{}, err
	}

	callerUserId, callerSpaceId, callerSpaceMemberId := int64(0), int64(0), int64(0)
	if caller != nil {
		callerUserId = caller.UserId
		callerSpaceId = caller.SpaceId
		callerSpaceMemberId = caller.SpaceMemberId
	}

	if _, err := u.identity.Management(ctx, caller, nil); err != nil {
		u.logger.Error(ctx, tag, "failed to authorize management scope", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainmodelsentity.Space]{}, err
	}

	if err := applicationshared.ValidatePagination(request.Page, request.Limit); err != nil {
		u.logger.Error(ctx, tag, "invalid space pagination", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainmodelsentity.Space]{}, err
	}

	items, total, err := u.space.ReadByFilter(ctx, domaincontractsrepository.SpaceFilter{
		Page:     request.Page,
		Limit:    request.Limit,
		Search:   request.Search,
		IsActive: request.IsActive,
	})
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list spaces", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainmodelsentity.Space]{}, err
	}

	if items == nil {
		items = []domainmodelsentity.Space{}
	}

	result = domainusecasesshared.Page[domainmodelsentity.Space]{Items: items, Total: total}

	return result, nil
}

func (u *usecase) UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64, request domainusecasesmanagement.UpdateSpaceRequest) (err error) {
	const tag = tagPath + "/UpdateById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceUpdateById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to update by id space", domainmodelslogger.Meta{
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
		if _, err := u.identity.Management(ctx, caller, nil); err != nil {
			u.logger.Error(ctx, tag, "failed to authorize management scope", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
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
				"id":                     id,
			})

			return err
		}

		if request.Name != nil {
			if err := u.validator.SpaceName(*request.Name); err != nil {
				u.logger.Error(ctx, tag, "invalid space name", domainmodelslogger.Meta{
					"caller_user_id":         callerUserId,
					"caller_space_id":        callerSpaceId,
					"caller_space_member_id": callerSpaceMemberId,
					"err":                    err,
					"id":                     id,
				})

				return err
			}
		}

		if request.Description != nil {
			if err := u.validator.SpaceDesc(*request.Description); err != nil {
				u.logger.Error(ctx, tag, "invalid space description", domainmodelslogger.Meta{
					"caller_user_id":         callerUserId,
					"caller_space_id":        callerSpaceId,
					"caller_space_member_id": callerSpaceMemberId,
					"err":                    err,
					"id":                     id,
				})

				return err
			}
		}

		if err := applicationshared.ValidateJSON(request.Preferences); err != nil {
			u.logger.Error(ctx, tag, "invalid space preferences JSON", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"id":                     id,
			})

			return err
		}

		if err := u.space.UpdateById(ctx, id, domaincontractsrepository.UpdateSpaceArgs{
			Name:        request.Name,
			Description: request.Description,
			IsActive:    request.IsActive,
			Preferences: request.Preferences,
			By:          &caller.UserId,
		}); err != nil {
			u.logger.Error(ctx, tag, "failed to update space", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"id":                     id,
			})

			return err
		}

		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete space update transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"id":                     id,
		})

		return err
	}

	if request.IsActive != nil {
		u.logger.Info(ctx, tag, "updated space", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"id":                     id,
		})
	}

	return nil
}

func (u *usecase) DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) (err error) {
	const tag = tagPath + "/DeleteById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementSpaceDeleteById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to delete by id space", domainmodelslogger.Meta{
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
		if _, err := u.identity.Management(ctx, caller, nil); err != nil {
			u.logger.Error(ctx, tag, "failed to authorize management scope", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
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
				"id":                     id,
			})

			return err
		}

		if err := u.space.DeleteById(ctx, id, &caller.UserId); err != nil {
			u.logger.Error(ctx, tag, "failed to delete space by id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"id":                     id,
			})

			return err
		}

		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete space deletion transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"id":                     id,
		})

		return err
	}

	u.logger.Info(ctx, tag, "deleted space", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"id":                     id,
	})

	return nil
}
