package applicationmanagementuser

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

const tagPath = "management/user"

type usecase struct {
	user        domaincontractsrepository.User
	space       domaincontractsrepository.Space
	spaceMember domaincontractsrepository.SpaceMember
	classicAuth domaincontractsrepository.ClassicAuth
	guard       domaincontractsutility.PermissionGuard
	validator   domaincontractsutility.Validator
	logger      domaincontractsutility.Logger
	transactor  domaincontractsutility.Transactor
	password    domaincontractsutility.Password
	identity    *applicationshared.IdentityReader
}

func NewUsecaseImpl(
	user domaincontractsrepository.User,
	space domaincontractsrepository.Space,
	spaceMember domaincontractsrepository.SpaceMember,
	classicAuth domaincontractsrepository.ClassicAuth,
	guard domaincontractsutility.PermissionGuard,
	validator domaincontractsutility.Validator,
	logger domaincontractsutility.Logger,
	transactor domaincontractsutility.Transactor,
	password domaincontractsutility.Password,
	controlSpaceSlug string,
) domainusecasesmanagement.User {
	return &usecase{
		user:        user,
		space:       space,
		spaceMember: spaceMember,
		classicAuth: classicAuth,
		guard:       guard,
		validator:   validator,
		logger:      logger,
		transactor:  transactor,
		password:    password,
		identity:    applicationshared.NewIdentityReader(user, space, spaceMember, controlSpaceSlug),
	}
}

func (u *usecase) Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesmanagement.CreateUserRequest) (result int64, err error) {
	const tag = tagPath + "/Create"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementUserCreate}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to create user", domainmodelslogger.Meta{
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

		if err := u.validator.UserName(request.Name); err != nil {
			u.logger.Error(ctx, tag, "invalid user name", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		if request.Bio != nil {
			if err := u.validator.UserBio(*request.Bio); err != nil {
				u.logger.Error(ctx, tag, "invalid user bio", domainmodelslogger.Meta{
					"caller_user_id":         callerUserId,
					"caller_space_id":        callerSpaceId,
					"caller_space_member_id": callerSpaceMemberId,
					"err":                    err,
				})

				return err
			}
		}

		if err := applicationshared.ValidateJSON(request.Preferences); err != nil {
			u.logger.Error(ctx, tag, "invalid user preferences JSON", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		if err := u.validator.UserUsername(request.Username); err != nil {
			u.logger.Error(ctx, tag, "invalid user username", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		if err := u.validator.UserPassword(request.Password); err != nil {
			u.logger.Error(ctx, tag, "invalid user password", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		if request.Email != nil {
			if err := u.validator.UserEmail(*request.Email); err != nil {
				u.logger.Error(ctx, tag, "invalid user email", domainmodelslogger.Meta{
					"caller_user_id":         callerUserId,
					"caller_space_id":        callerSpaceId,
					"caller_space_member_id": callerSpaceMemberId,
					"err":                    err,
				})

				return err
			}
		}

		if request.Phone != nil {
			if err := u.validator.UserPhone(*request.Phone); err != nil {
				u.logger.Error(ctx, tag, "invalid user phone", domainmodelslogger.Meta{
					"caller_user_id":         callerUserId,
					"caller_space_id":        callerSpaceId,
					"caller_space_member_id": callerSpaceMemberId,
					"err":                    err,
				})

				return err
			}
		}

		hash, err := u.password.Hash(request.Password)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to hash password", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		id, err := u.user.Create(ctx, domaincontractsrepository.CreateUserArgs{
			Name:        request.Name,
			Bio:         request.Bio,
			AvatarPath:  request.AvatarPath,
			Preferences: request.Preferences,
			By:          &caller.UserId,
		})
		if err != nil {
			u.logger.Error(ctx, tag, "failed to create user", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}
		_, err = u.classicAuth.Create(ctx, domaincontractsrepository.CreateClassicAuthArgs{
			UserId:          id,
			Username:        request.Username,
			PasswordHash:    hash,
			Email:           request.Email,
			Phone:           request.Phone,
			IsEmailVerified: request.IsEmailVerified,
			IsPhoneVerified: request.IsPhoneVerified,
			By:              &caller.UserId,
		})
		if err != nil {
			u.logger.Error(ctx, tag, "failed to create classic auth", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		result = id
		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete user creation transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return 0, err
	}

	u.logger.Info(ctx, tag, "created user", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"id":                     result,
	})

	return result, nil
}

func (u *usecase) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) (result *domainmodelsentity.User, err error) {
	const tag = tagPath + "/ReadById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementUserReadById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read user by id", domainmodelslogger.Meta{
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

	value, err := u.user.ReadById(ctx, id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read user by id", domainmodelslogger.Meta{
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

func (u *usecase) ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesmanagement.FilterUserRequest) (result domainusecasesshared.Page[domainmodelsentity.User], err error) {
	const tag = tagPath + "/ReadByFilter"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementUserReadByFilter}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to list users", domainmodelslogger.Meta{
			"err": err,
		})

		return domainusecasesshared.Page[domainmodelsentity.User]{}, err
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

		return domainusecasesshared.Page[domainmodelsentity.User]{}, err
	}

	if err := applicationshared.ValidatePagination(request.Page, request.Limit); err != nil {
		u.logger.Error(ctx, tag, "invalid user pagination", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainmodelsentity.User]{}, err
	}

	items, total, err := u.user.ReadByFilter(ctx, domaincontractsrepository.UserFilter{
		Page:   request.Page,
		Limit:  request.Limit,
		Search: request.Search,
	})
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list users", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainmodelsentity.User]{}, err
	}

	if items == nil {
		items = []domainmodelsentity.User{}
	}

	result = domainusecasesshared.Page[domainmodelsentity.User]{Items: items, Total: total}

	return result, nil
}

func (u *usecase) UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64, request domainusecasesmanagement.UpdateUserRequest) (err error) {
	const tag = tagPath + "/UpdateById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementUserUpdateById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to update by id user", domainmodelslogger.Meta{
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
			if err := u.validator.UserName(*request.Name); err != nil {
				u.logger.Error(ctx, tag, "invalid user name", domainmodelslogger.Meta{
					"caller_user_id":         callerUserId,
					"caller_space_id":        callerSpaceId,
					"caller_space_member_id": callerSpaceMemberId,
					"err":                    err,
					"id":                     id,
				})

				return err
			}
		}

		if request.Bio != nil {
			if err := u.validator.UserBio(*request.Bio); err != nil {
				u.logger.Error(ctx, tag, "invalid user bio", domainmodelslogger.Meta{
					"caller_user_id":         callerUserId,
					"caller_space_id":        callerSpaceId,
					"caller_space_member_id": callerSpaceMemberId,
					"err":                    err,
					"id":                     id,
				})

				return err
			}
		}

		if err := applicationshared.ValidateClear(request.AvatarPath, request.AvatarPathSetNull); err != nil {
			u.logger.Error(ctx, tag, "conflicting avatar path value and null inputs", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"id":                     id,
			})

			return err
		}

		if err := applicationshared.ValidateJSON(request.Preferences); err != nil {
			u.logger.Error(ctx, tag, "invalid user preferences JSON", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"id":                     id,
			})

			return err
		}

		if err := u.user.UpdateById(ctx, id, domaincontractsrepository.UpdateUserArgs{
			Name:              request.Name,
			Bio:               request.Bio,
			AvatarPath:        request.AvatarPath,
			AvatarPathSetNull: request.AvatarPathSetNull,
			Preferences:       request.Preferences,
			By:                &caller.UserId,
		}); err != nil {
			u.logger.Error(ctx, tag, "failed to update user", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete user update transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"id":                     id,
		})

		return err
	}

	return nil
}

func (u *usecase) DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) (err error) {
	const tag = tagPath + "/DeleteById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementUserDeleteById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to delete by id user", domainmodelslogger.Meta{
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

		if err := u.user.DeleteById(ctx, id, &caller.UserId); err != nil {
			u.logger.Error(ctx, tag, "failed to delete user by id", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete user deletion transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"id":                     id,
		})

		return err
	}

	u.logger.Info(ctx, tag, "deleted user", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"id":                     id,
	})

	return nil
}
