package applicationprofileaccount

import (
	"context"

	applicationpermissions "github.com/MateWorkspace/lockmate/internal/application/permissions"
	applicationshared "github.com/MateWorkspace/lockmate/internal/application/shared"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	domainusecasesprofile "github.com/MateWorkspace/lockmate/internal/domain/usecases/profile"
)

const tagPath = "profile/account"

type usecase struct {
	user            domaincontractsrepository.User
	space           domaincontractsrepository.Space
	spaceMember     domaincontractsrepository.SpaceMember
	spaceMemberRole domaincontractsrepository.SpaceMemberRole
	guard           domaincontractsutility.PermissionGuard
	validator       domaincontractsutility.Validator
	transactor      domaincontractsutility.Transactor
	logger          domaincontractsutility.Logger
	identity        *applicationshared.IdentityReader
}

func NewUsecaseImpl(
	user domaincontractsrepository.User,
	space domaincontractsrepository.Space,
	spaceMember domaincontractsrepository.SpaceMember,
	spaceMemberRole domaincontractsrepository.SpaceMemberRole,
	guard domaincontractsutility.PermissionGuard,
	validator domaincontractsutility.Validator,
	transactor domaincontractsutility.Transactor,
	logger domaincontractsutility.Logger,
) domainusecasesprofile.Account {
	return &usecase{
		user:            user,
		space:           space,
		spaceMember:     spaceMember,
		spaceMemberRole: spaceMemberRole,
		guard:           guard,
		validator:       validator,
		transactor:      transactor,
		logger:          logger,
		identity:        applicationshared.NewIdentityReader(user, space, spaceMember, ""),
	}
}

func (u *usecase) GetProfile(ctx context.Context, caller *domainmodelsentity.AccessClaims) (result *domainusecasesprofile.ProfileResult, err error) {
	const tag = tagPath + "/GetProfile"

	if err := u.guard.Check(caller, []string{applicationpermissions.ProfileAccountGetProfile}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to get profile", domainmodelslogger.Meta{
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

	result = &domainusecasesprofile.ProfileResult{
		User:        identity.User,
		Space:       identity.Space,
		SpaceMember: identity.SpaceMember,
	}

	return result, nil
}

func (u *usecase) GetRoles(ctx context.Context, caller *domainmodelsentity.AccessClaims) (result []domainmodelsentity.Role, err error) {
	const tag = tagPath + "/GetRoles"

	if err := u.guard.Check(caller, []string{applicationpermissions.ProfileAccountGetRoles}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to get roles", domainmodelslogger.Meta{
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

	result, err = u.spaceMemberRole.ReadEffectiveRolesBySpaceMemberId(ctx, identity.Space.Id, identity.SpaceMember.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read profile roles", domainmodelslogger.Meta{
			"err":             err,
			"user_id":         identity.User.Id,
			"space_id":        identity.Space.Id,
			"space_member_id": identity.SpaceMember.Id,
		})

		return nil, err
	}

	if result == nil {
		result = []domainmodelsentity.Role{}
	}

	return result, nil
}

func (u *usecase) GetPermissions(ctx context.Context, caller *domainmodelsentity.AccessClaims) (result []domainmodelsentity.Permission, err error) {
	const tag = tagPath + "/GetPermissions"

	if err := u.guard.Check(caller, []string{applicationpermissions.ProfileAccountGetPermissions}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to get permissions", domainmodelslogger.Meta{
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

	result, err = u.spaceMemberRole.ReadEffectivePermissionsBySpaceMemberId(ctx, identity.Space.Id, identity.SpaceMember.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read profile permissions", domainmodelslogger.Meta{
			"err":             err,
			"user_id":         identity.User.Id,
			"space_id":        identity.Space.Id,
			"space_member_id": identity.SpaceMember.Id,
		})

		return nil, err
	}

	if result == nil {
		result = []domainmodelsentity.Permission{}
	}

	return result, nil
}

func (u *usecase) UpdateProfile(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesprofile.UpdateProfileRequest) (err error) {
	const tag = tagPath + "/UpdateProfile"

	if err := u.guard.Check(caller, []string{applicationpermissions.ProfileAccountUpdateProfile}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to update profile", domainmodelslogger.Meta{
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

		if request.Name != nil {
			if err := u.validator.UserName(*request.Name); err != nil {
				u.logger.Error(ctx, tag, "invalid user name", domainmodelslogger.Meta{
					"caller_user_id":         callerUserId,
					"caller_space_id":        callerSpaceId,
					"caller_space_member_id": callerSpaceMemberId,
					"err":                    err,
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
			})

			return err
		}

		if err := applicationshared.ValidateJSON(request.Preferences); err != nil {
			u.logger.Error(ctx, tag, "invalid account preferences JSON", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		if err := u.user.UpdateById(ctx, identity.User.Id, domaincontractsrepository.UpdateUserArgs{
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
			})

			return err
		}

		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete account profile update transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return err
	}

	return nil
}

func (u *usecase) UpdateMembershipPreferences(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesprofile.UpdateMembershipPreferencesRequest) (err error) {
	const tag = tagPath + "/UpdateMembershipPreferences"

	if err := u.guard.Check(caller, []string{applicationpermissions.ProfileAccountUpdateMembershipPreferences}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to update membership preferences", domainmodelslogger.Meta{
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

		if err := applicationshared.ValidateJSON(&request.Preferences); err != nil {
			u.logger.Error(ctx, tag, "invalid account preferences JSON", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
			})

			return err
		}

		if err := u.spaceMember.UpdateById(ctx, identity.Space.Id, identity.SpaceMember.Id, domaincontractsrepository.UpdateSpaceMemberArgs{
			Preferences: &request.Preferences,
			By:          &caller.UserId,
		}); err != nil {
			u.logger.Error(ctx, tag, "failed to update space member", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete account membership preference update transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return err
	}

	return nil
}
