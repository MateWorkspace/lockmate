package applicationprofilesecurity

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
	domainusecasesprofile "github.com/MateWorkspace/lockmate/internal/domain/usecases/profile"
)

const tagPath = "profile/security"

type usecase struct {
	user        domaincontractsrepository.User
	space       domaincontractsrepository.Space
	spaceMember domaincontractsrepository.SpaceMember
	classicAuth domaincontractsrepository.ClassicAuth
	googleAuth  domaincontractsrepository.GoogleAuth
	password    domaincontractsutility.Password
	guard       domaincontractsutility.PermissionGuard
	validator   domaincontractsutility.Validator
	transactor  domaincontractsutility.Transactor
	logger      domaincontractsutility.Logger
	identity    *applicationshared.IdentityReader
}

func NewUsecaseImpl(
	user domaincontractsrepository.User,
	space domaincontractsrepository.Space,
	spaceMember domaincontractsrepository.SpaceMember,
	classicAuth domaincontractsrepository.ClassicAuth,
	googleAuth domaincontractsrepository.GoogleAuth,
	password domaincontractsutility.Password,
	guard domaincontractsutility.PermissionGuard,
	validator domaincontractsutility.Validator,
	transactor domaincontractsutility.Transactor,
	logger domaincontractsutility.Logger,
) domainusecasesprofile.Security {
	return &usecase{
		user:        user,
		space:       space,
		spaceMember: spaceMember,
		classicAuth: classicAuth,
		googleAuth:  googleAuth,
		password:    password,
		guard:       guard,
		validator:   validator,
		transactor:  transactor,
		logger:      logger,
		identity:    applicationshared.NewIdentityReader(user, space, spaceMember, ""),
	}
}

func (u *usecase) GetAuthMethods(ctx context.Context, caller *domainmodelsentity.AccessClaims) (result *domainusecasesprofile.AuthMethodsResult, err error) {
	const tag = tagPath + "/GetAuthMethods"

	if err := u.guard.Check(caller, []string{applicationpermissions.ProfileSecurityGetAuthMethods}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to get auth methods", domainmodelslogger.Meta{
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

	identity, err := u.identity.Read(ctx, caller)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read caller identity", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return nil, err
	}

	classic, err := u.classicAuth.ReadByUserId(ctx, identity.User.Id)
	if errors.Is(err, domainmodelserror.ClassicAuthNotFound) {
		classic = nil
	} else if err != nil {
		u.logger.Error(ctx, tag, "failed to read classic auth by user id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return nil, err
	}

	google, err := u.googleAuth.ReadByUserId(ctx, identity.User.Id)
	if errors.Is(err, domainmodelserror.GoogleAuthNotFound) {
		google = nil
	} else if err != nil {
		u.logger.Error(ctx, tag, "failed to read google auth by user id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return nil, err
	}

	if classic != nil && (classic.UserId != identity.User.Id || classic.AuditDelete.At != nil) {
		failure := domainmodelserror.NewError("invalid authentication owner", domainmodelserror.BadState, nil)
		u.logger.Error(ctx, tag, "invalid authentication owner", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    failure,
		})

		return nil, failure
	}

	if google != nil && (google.UserId != identity.User.Id || google.AuditDelete.At != nil) {
		failure := domainmodelserror.NewError("invalid authentication owner", domainmodelserror.BadState, nil)
		u.logger.Error(ctx, tag, "invalid authentication owner", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    failure,
		})

		return nil, failure
	}

	result = &domainusecasesprofile.AuthMethodsResult{
		ClassicAuth: applicationshared.ClassicAuthView(classic),
		GoogleAuth:  google,
	}

	return result, nil
}

func (u *usecase) UpdateClassicAuth(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesprofile.UpdateClassicAuthRequest) (err error) {
	const tag = tagPath + "/UpdateClassicAuth"

	if err := u.guard.Check(caller, []string{applicationpermissions.ProfileSecurityUpdateClassicAuth}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to update classic auth", domainmodelslogger.Meta{
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
		identity, err := u.identity.Read(ctx, caller)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read caller identity", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
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

		if err := applicationshared.ValidateClear(request.Email, request.EmailSetNull); err != nil {
			u.logger.Error(ctx, tag, "conflicting email value and null inputs", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		if err := applicationshared.ValidateClear(request.Phone, request.PhoneSetNull); err != nil {
			u.logger.Error(ctx, tag, "conflicting phone value and null inputs", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		auth, err := u.classicAuth.ReadByUserId(ctx, identity.User.Id)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read classic auth by user id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		if auth == nil || auth.UserId != identity.User.Id || auth.AuditDelete.At != nil {
			failure := domainmodelserror.NewError("classic authentication not found", domainmodelserror.ClassicAuthNotFound, nil)
			u.logger.Error(ctx, tag, "classic authentication not found", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    failure,
			})

			return failure
		}

		if err := u.classicAuth.UpdateById(ctx, auth.Id, domaincontractsrepository.UpdateClassicAuthArgs{
			Username:     request.Username,
			Email:        request.Email,
			EmailSetNull: request.EmailSetNull,
			Phone:        request.Phone,
			PhoneSetNull: request.PhoneSetNull,
			By:           &caller.UserId,
		}); err != nil {
			u.logger.Error(ctx, tag, "failed to update classic auth", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete security credential update transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return err
	}

	u.logger.Info(ctx, tag, "updated classic auth", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
	})

	return nil
}

func (u *usecase) ChangePassword(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesprofile.ChangePasswordRequest) (err error) {
	const tag = tagPath + "/ChangePassword"

	if err := u.guard.Check(caller, []string{applicationpermissions.ProfileSecurityChangePassword}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to change password", domainmodelslogger.Meta{
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
		identity, err := u.identity.Read(ctx, caller)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read caller identity", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		auth, err := u.classicAuth.ReadByUserId(ctx, identity.User.Id)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read classic auth by user id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		if auth == nil || auth.UserId != identity.User.Id || auth.AuditDelete.At != nil {
			failure := domainmodelserror.NewError("classic authentication not found", domainmodelserror.ClassicAuthNotFound, nil)
			u.logger.Error(ctx, tag, "classic authentication not found", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    failure,
			})

			return failure
		}

		if err := u.password.Compare(auth.PasswordHash, request.CurrentPassword); err != nil {
			u.logger.Error(ctx, tag, "failed to compare current password", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		if err := u.validator.UserPassword(request.NewPassword); err != nil {
			u.logger.Error(ctx, tag, "invalid new password", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		hash, err := u.password.Hash(request.NewPassword)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to hash password", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		if err := u.classicAuth.UpdateById(ctx, auth.Id, domaincontractsrepository.UpdateClassicAuthArgs{PasswordHash: &hash, By: &caller.UserId}); err != nil {
			u.logger.Error(ctx, tag, "failed to update classic auth", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete security password change transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return err
	}

	u.logger.Info(ctx, tag, "changed password", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
	})

	return nil
}
