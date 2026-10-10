package applicationmanagementclassicauth

import (
	"context"
	"errors"

	applicationpermissions "github.com/MateWorkspace/lockmate/internal/application/permissions"
	applicationshared "github.com/MateWorkspace/lockmate/internal/application/shared"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

const tagPath = "management/classic_auth"

type usecase struct {
	user        domaincontractsrepository.User
	space       domaincontractsrepository.Space
	spaceMember domaincontractsrepository.SpaceMember
	classicAuth domaincontractsrepository.ClassicAuth
	googleAuth  domaincontractsrepository.GoogleAuth
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
	googleAuth domaincontractsrepository.GoogleAuth,
	guard domaincontractsutility.PermissionGuard,
	validator domaincontractsutility.Validator,
	logger domaincontractsutility.Logger,
	transactor domaincontractsutility.Transactor,
	password domaincontractsutility.Password,
	controlSpaceSlug string,
) domainusecasesmanagement.ClassicAuth {
	return &usecase{
		user:        user,
		space:       space,
		spaceMember: spaceMember,
		classicAuth: classicAuth,
		googleAuth:  googleAuth,
		guard:       guard,
		validator:   validator,
		logger:      logger,
		transactor:  transactor,
		password:    password,
		identity:    applicationshared.NewIdentityReader(user, space, spaceMember, controlSpaceSlug),
	}
}

func (u *usecase) Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesmanagement.CreateClassicAuthRequest) (result int64, err error) {
	const tag = tagPath + "/Create"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClassicAuthCreate}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to create classic auth", domainmodelslogger.Meta{
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

		if err := applicationshared.ValidateIds(request.UserId); err != nil {
			u.logger.Error(ctx, tag, "invalid user id", domainmodelslogger.Meta{
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

		if err := u.validator.UserPassword(request.Password); err != nil {
			u.logger.Error(ctx, tag, "invalid user password", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		if err := applicationshared.ValidateJSON(request.Preferences); err != nil {
			u.logger.Error(ctx, tag, "invalid classic auth preferences JSON", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
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

		value, err := u.classicAuth.Create(ctx, domaincontractsrepository.CreateClassicAuthArgs{
			UserId:          request.UserId,
			Username:        request.Username,
			Email:           request.Email,
			Phone:           request.Phone,
			PasswordHash:    hash,
			IsEmailVerified: request.IsEmailVerified,
			IsPhoneVerified: request.IsPhoneVerified,
			Preferences:     request.Preferences,
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

		result = value
		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete classic auth creation transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return 0, err
	}

	u.logger.Info(ctx, tag, "created classic auth", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"id":                     result,
	})

	return result, nil
}

func (u *usecase) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) (result *domainusecasesshared.ClassicAuthView, err error) {
	const tag = tagPath + "/ReadById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClassicAuthReadById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read classic auth by id", domainmodelslogger.Meta{
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

	value, err := u.classicAuth.ReadById(ctx, id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read classic auth by id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"id":                     id,
		})

		return nil, err
	}

	result = applicationshared.ClassicAuthView(value)

	return result, nil
}

func (u *usecase) ReadByUserId(ctx context.Context, caller *domainmodelsentity.AccessClaims, selector int64) (result *domainusecasesshared.ClassicAuthView, err error) {
	const tag = tagPath + "/ReadByUserId"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClassicAuthReadByUserId}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read classic auth by user id", domainmodelslogger.Meta{
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
			"user_id":                selector,
		})

		return nil, err
	}

	value, err := u.classicAuth.ReadByUserId(ctx, selector)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read classic auth by user id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"user_id":                selector,
		})

		return nil, err
	}

	result = applicationshared.ClassicAuthView(value)

	return result, nil
}

func (u *usecase) ReadByUsername(ctx context.Context, caller *domainmodelsentity.AccessClaims, selector string) (result *domainusecasesshared.ClassicAuthView, err error) {
	const tag = tagPath + "/ReadByUsername"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClassicAuthReadByUsername}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read classic auth by username", domainmodelslogger.Meta{
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

	value, err := u.classicAuth.ReadByUsername(ctx, selector)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read classic auth by username", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return nil, err
	}

	result = applicationshared.ClassicAuthView(value)

	return result, nil
}

func (u *usecase) ReadByEmail(ctx context.Context, caller *domainmodelsentity.AccessClaims, selector string) (result *domainusecasesshared.ClassicAuthView, err error) {
	const tag = tagPath + "/ReadByEmail"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClassicAuthReadByEmail}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read classic auth by email", domainmodelslogger.Meta{
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

	value, err := u.classicAuth.ReadByEmail(ctx, selector)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read classic auth by email", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return nil, err
	}

	result = applicationshared.ClassicAuthView(value)

	return result, nil
}

func (u *usecase) ReadByPhone(ctx context.Context, caller *domainmodelsentity.AccessClaims, selector string) (result *domainusecasesshared.ClassicAuthView, err error) {
	const tag = tagPath + "/ReadByPhone"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClassicAuthReadByPhone}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read classic auth by phone", domainmodelslogger.Meta{
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

	value, err := u.classicAuth.ReadByPhone(ctx, selector)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read classic auth by phone", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return nil, err
	}

	result = applicationshared.ClassicAuthView(value)

	return result, nil
}

func (u *usecase) ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesmanagement.FilterClassicAuthRequest) (result domainusecasesshared.Page[domainusecasesshared.ClassicAuthView], err error) {
	const tag = tagPath + "/ReadByFilter"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClassicAuthReadByFilter}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to list classic authentication methods", domainmodelslogger.Meta{
			"err": err,
		})

		return domainusecasesshared.Page[domainusecasesshared.ClassicAuthView]{}, err
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

		return domainusecasesshared.Page[domainusecasesshared.ClassicAuthView]{}, err
	}

	if err := applicationshared.ValidateOptionalId(request.UserId); err != nil {
		u.logger.Error(ctx, tag, "invalid user id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainusecasesshared.ClassicAuthView]{}, err
	}

	if err := applicationshared.ValidatePagination(request.Page, request.Limit); err != nil {
		u.logger.Error(ctx, tag, "invalid classic auth pagination", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainusecasesshared.ClassicAuthView]{}, err
	}

	items, total, err := u.classicAuth.ReadByFilter(ctx, domaincontractsrepository.ClassicAuthFilter{
		Page:            request.Page,
		Limit:           request.Limit,
		Search:          request.Search,
		UserId:          request.UserId,
		IsEmailVerified: request.IsEmailVerified,
		IsPhoneVerified: request.IsPhoneVerified,
	})
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list classic authentication methods", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainusecasesshared.ClassicAuthView]{}, err
	}

	if items == nil {
		items = []domainusecasesshared.ClassicAuthView{}
	}

	result = domainusecasesshared.Page[domainusecasesshared.ClassicAuthView]{Items: items, Total: total}

	return result, nil
}

func (u *usecase) UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64, request domainusecasesmanagement.UpdateClassicAuthRequest) (err error) {
	const tag = tagPath + "/UpdateById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClassicAuthUpdateById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to update by id classic auth", domainmodelslogger.Meta{
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

		if request.Username != nil {
			if err := u.validator.UserUsername(*request.Username); err != nil {
				u.logger.Error(ctx, tag, "invalid user username", domainmodelslogger.Meta{
					"caller_user_id":         callerUserId,
					"caller_space_id":        callerSpaceId,
					"caller_space_member_id": callerSpaceMemberId,
					"err":                    err,
					"id":                     id,
				})

				return err
			}
		}

		if request.Email != nil {
			if err := u.validator.UserEmail(*request.Email); err != nil {
				u.logger.Error(ctx, tag, "invalid user email", domainmodelslogger.Meta{
					"caller_user_id":         callerUserId,
					"caller_space_id":        callerSpaceId,
					"caller_space_member_id": callerSpaceMemberId,
					"err":                    err,
					"id":                     id,
				})

				return err
			}
		}

		if err := applicationshared.ValidateClear(request.Email, request.EmailSetNull); err != nil {
			u.logger.Error(ctx, tag, "conflicting email value and null inputs", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"id":                     id,
			})

			return err
		}

		if request.Phone != nil {
			if err := u.validator.UserPhone(*request.Phone); err != nil {
				u.logger.Error(ctx, tag, "invalid user phone", domainmodelslogger.Meta{
					"caller_user_id":         callerUserId,
					"caller_space_id":        callerSpaceId,
					"caller_space_member_id": callerSpaceMemberId,
					"err":                    err,
					"id":                     id,
				})

				return err
			}
		}

		if err := applicationshared.ValidateClear(request.Phone, request.PhoneSetNull); err != nil {
			u.logger.Error(ctx, tag, "conflicting phone value and null inputs", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"id":                     id,
			})

			return err
		}

		if err := applicationshared.ValidateJSON(request.Preferences); err != nil {
			u.logger.Error(ctx, tag, "invalid classic auth preferences JSON", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"id":                     id,
			})

			return err
		}

		if err := u.classicAuth.UpdateById(ctx, id, domaincontractsrepository.UpdateClassicAuthArgs{
			Username:        request.Username,
			Email:           request.Email,
			EmailSetNull:    request.EmailSetNull,
			Phone:           request.Phone,
			PhoneSetNull:    request.PhoneSetNull,
			IsEmailVerified: request.IsEmailVerified,
			IsPhoneVerified: request.IsPhoneVerified,
			Preferences:     request.Preferences,
			By:              &caller.UserId,
		}); err != nil {
			u.logger.Error(ctx, tag, "failed to update classic auth", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete classic auth update transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"id":                     id,
		})

		return err
	}

	u.logger.Info(ctx, tag, "updated classic auth", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"id":                     id,
	})

	return nil
}

func (u *usecase) ResetPassword(ctx context.Context, caller *domainmodelsentity.AccessClaims, userId int64, request domainusecasesmanagement.ResetUserPasswordRequest) (err error) {
	const tag = tagPath + "/ResetPassword"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClassicAuthResetPassword}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to reset password classic auth", domainmodelslogger.Meta{
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
				"user_id":                userId,
			})

			return err
		}

		if err := applicationshared.ValidateIds(userId); err != nil {
			u.logger.Error(ctx, tag, "invalid user id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"user_id":                userId,
			})

			return err
		}

		if err := u.validator.UserPassword(request.Password); err != nil {
			u.logger.Error(ctx, tag, "invalid user password", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"user_id":                userId,
			})

			return err
		}

		value, err := u.classicAuth.ReadByUserId(ctx, userId)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read classic auth by user id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"user_id":                userId,
			})

			return err
		}

		if value == nil {
			failure := domainmodelserror.NewError("classic authentication not found", domainmodelserror.ClassicAuthNotFound, nil)
			u.logger.Error(ctx, tag, "classic authentication not found", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    failure,
				"user_id":                userId,
			})

			return failure
		}

		hash, err := u.password.Hash(request.Password)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to hash password", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"user_id":                userId,
			})

			return err
		}

		if err := u.classicAuth.UpdateById(ctx, value.Id, domaincontractsrepository.UpdateClassicAuthArgs{PasswordHash: &hash, By: &caller.UserId}); err != nil {
			u.logger.Error(ctx, tag, "failed to update classic auth", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"user_id":                userId,
			})

			return err
		}

		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete classic auth password reset transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"user_id":                userId,
		})

		return err
	}

	u.logger.Info(ctx, tag, "reset password for classic auth", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"user_id":                userId,
	})

	return nil
}

func (u *usecase) DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) (err error) {
	const tag = tagPath + "/DeleteById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClassicAuthDeleteById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to delete by id classic auth", domainmodelslogger.Meta{
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

		value, err := u.classicAuth.ReadById(ctx, id)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read classic auth by id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"id":                     id,
			})

			return err
		}

		if value == nil {
			failure := domainmodelserror.NewError("classic authentication not found", domainmodelserror.ClassicAuthNotFound, nil)
			u.logger.Error(ctx, tag, "classic authentication not found", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    failure,
				"id":                     id,
			})

			return failure
		}

		google, err := u.googleAuth.ReadByUserId(ctx, value.UserId)
		if errors.Is(err, domainmodelserror.GoogleAuthNotFound) {
			failure := domainmodelserror.NewError("cannot remove last authentication method", domainmodelserror.BadState, err)
			u.logger.Error(ctx, tag, "cannot remove last authentication method", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    failure,
				"id":                     id,
			})

			return failure
		}

		if err != nil {
			u.logger.Error(ctx, tag, "failed to read google auth by user id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"id":                     id,
			})

			return err
		}

		if google == nil || google.UserId != value.UserId || google.AuditDelete.At != nil {
			failure := domainmodelserror.NewError("cannot remove last authentication method", domainmodelserror.BadState, nil)
			u.logger.Error(ctx, tag, "cannot remove last authentication method", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    failure,
				"id":                     id,
			})

			return failure
		}

		if err := u.classicAuth.DeleteById(ctx, id, &caller.UserId); err != nil {
			u.logger.Error(ctx, tag, "failed to delete classic auth by id", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete classic auth deletion transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"id":                     id,
		})

		return err
	}

	u.logger.Info(ctx, tag, "deleted classic auth", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"id":                     id,
	})

	return nil
}
