package applicationmanagementgoogleauth

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

const tagPath = "management/google_auth"

type usecase struct {
	user        domaincontractsrepository.User
	space       domaincontractsrepository.Space
	spaceMember domaincontractsrepository.SpaceMember
	googleAuth  domaincontractsrepository.GoogleAuth
	guard       domaincontractsutility.PermissionGuard
	logger      domaincontractsutility.Logger
	identity    *applicationshared.IdentityReader
}

func NewUsecaseImpl(
	user domaincontractsrepository.User,
	space domaincontractsrepository.Space,
	spaceMember domaincontractsrepository.SpaceMember,
	googleAuth domaincontractsrepository.GoogleAuth,
	guard domaincontractsutility.PermissionGuard,
	logger domaincontractsutility.Logger,
	controlSpaceSlug string,
) domainusecasesmanagement.GoogleAuth {
	return &usecase{
		user:        user,
		space:       space,
		spaceMember: spaceMember,
		googleAuth:  googleAuth,
		guard:       guard,
		logger:      logger,
		identity:    applicationshared.NewIdentityReader(user, space, spaceMember, controlSpaceSlug),
	}
}

func (u *usecase) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) (result *domainmodelsentity.GoogleAuth, err error) {
	const tag = tagPath + "/ReadById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementGoogleAuthReadById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read google auth by id", domainmodelslogger.Meta{
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

	value, err := u.googleAuth.ReadById(ctx, id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read google auth by id", domainmodelslogger.Meta{
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

func (u *usecase) ReadByUserId(ctx context.Context, caller *domainmodelsentity.AccessClaims, selector int64) (result *domainmodelsentity.GoogleAuth, err error) {
	const tag = tagPath + "/ReadByUserId"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementGoogleAuthReadByUserId}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read google auth by user id", domainmodelslogger.Meta{
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

	value, err := u.googleAuth.ReadByUserId(ctx, selector)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read google auth by user id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"user_id":                selector,
		})

		return nil, err
	}

	result = value

	return result, nil
}

func (u *usecase) ReadBySubject(ctx context.Context, caller *domainmodelsentity.AccessClaims, selector string) (result *domainmodelsentity.GoogleAuth, err error) {
	const tag = tagPath + "/ReadBySubject"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementGoogleAuthReadBySubject}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read google auth by subject", domainmodelslogger.Meta{
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

	value, err := u.googleAuth.ReadBySubject(ctx, selector)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read google auth by subject", domainmodelslogger.Meta{
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

func (u *usecase) ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesmanagement.FilterGoogleAuthRequest) (result domainusecasesshared.Page[domainmodelsentity.GoogleAuth], err error) {
	const tag = tagPath + "/ReadByFilter"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementGoogleAuthReadByFilter}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to list google authentication methods", domainmodelslogger.Meta{
			"err": err,
		})

		return domainusecasesshared.Page[domainmodelsentity.GoogleAuth]{}, err
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

		return domainusecasesshared.Page[domainmodelsentity.GoogleAuth]{}, err
	}

	if err := applicationshared.ValidateOptionalId(request.UserId); err != nil {
		u.logger.Error(ctx, tag, "invalid user id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainmodelsentity.GoogleAuth]{}, err
	}

	if err := applicationshared.ValidatePagination(request.Page, request.Limit); err != nil {
		u.logger.Error(ctx, tag, "invalid google auth pagination", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainmodelsentity.GoogleAuth]{}, err
	}

	items, total, err := u.googleAuth.ReadByFilter(ctx, domaincontractsrepository.GoogleAuthFilter{
		Page:            request.Page,
		Limit:           request.Limit,
		Search:          request.Search,
		UserId:          request.UserId,
		IsEmailVerified: request.IsEmailVerified,
	})
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list google authentication methods", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainmodelsentity.GoogleAuth]{}, err
	}

	if items == nil {
		items = []domainmodelsentity.GoogleAuth{}
	}

	result = domainusecasesshared.Page[domainmodelsentity.GoogleAuth]{Items: items, Total: total}

	return result, nil
}
