package applicationmanagementclient

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
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

const tagPath = "management/client"

type usecase struct {
	user         domaincontractsrepository.User
	space        domaincontractsrepository.Space
	spaceMember  domaincontractsrepository.SpaceMember
	client       domaincontractsrepository.Client
	clientOrigin domaincontractsrepository.ClientOrigin
	guard        domaincontractsutility.PermissionGuard
	validator    domaincontractsutility.Validator
	logger       domaincontractsutility.Logger
	transactor   domaincontractsutility.Transactor
	keygen       domaincontractsutility.Keygen
	identity     *applicationshared.IdentityReader
}

func NewUsecaseImpl(
	user domaincontractsrepository.User,
	space domaincontractsrepository.Space,
	spaceMember domaincontractsrepository.SpaceMember,
	client domaincontractsrepository.Client,
	clientOrigin domaincontractsrepository.ClientOrigin,
	guard domaincontractsutility.PermissionGuard,
	validator domaincontractsutility.Validator,
	logger domaincontractsutility.Logger,
	transactor domaincontractsutility.Transactor,
	keygen domaincontractsutility.Keygen,
	controlSpaceSlug string,
) domainusecasesmanagement.Client {
	return &usecase{
		user:         user,
		space:        space,
		spaceMember:  spaceMember,
		client:       client,
		clientOrigin: clientOrigin,
		guard:        guard,
		validator:    validator,
		logger:       logger,
		transactor:   transactor,
		keygen:       keygen,
		identity:     applicationshared.NewIdentityReader(user, space, spaceMember, controlSpaceSlug),
	}
}

func (u *usecase) Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.CreateClientRequest) (result *domainusecasesmanagement.ClientSecretResult, err error) {
	const tag = tagPath + "/Create"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientCreate}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to create client", domainmodelslogger.Meta{
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

		if err := u.validator.ClientSlug(request.Slug); err != nil {
			u.logger.Error(ctx, tag, "invalid client slug", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		if err := u.validator.ClientName(request.Name); err != nil {
			u.logger.Error(ctx, tag, "invalid client name", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		if request.Description != nil {
			if err := u.validator.ClientDesc(*request.Description); err != nil {
				u.logger.Error(ctx, tag, "invalid client description", domainmodelslogger.Meta{
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
			u.logger.Error(ctx, tag, "invalid client preferences JSON", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		raw, hash, redacted, err := u.keygen.Generate()
		if err != nil {
			u.logger.Error(ctx, tag, "failed to generate client secret", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		id, err := u.client.Create(ctx, spaceId, domaincontractsrepository.CreateClientArgs{
			Slug:              request.Slug,
			Name:              request.Name,
			Description:       request.Description,
			SecretKeyHash:     hash,
			SecretKeyRedacted: redacted,
			IsActive:          request.IsActive,
			Preferences:       request.Preferences,
			By:                &caller.UserId,
		})
		if err != nil {
			u.logger.Error(ctx, tag, "failed to create client", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		value, err := u.client.ReadById(ctx, spaceId, id)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read client by id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
			})

			return err
		}

		if value == nil {
			failure := domainmodelserror.NewError("client result missing", domainmodelserror.Failure, nil)
			u.logger.Error(ctx, tag, "client result missing", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    failure,
				"space_id":               spaceId,
			})

			return failure
		}

		result = &domainusecasesmanagement.ClientSecretResult{
			Client: *applicationshared.ClientView(value),
			Secret: raw,
		}

		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete client creation transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return nil, err
	}

	u.logger.Info(ctx, tag, "created client", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"space_id":               spaceId,
		"id":                     result,
	})

	return result, nil
}

func (u *usecase) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) (result *domainusecasesshared.ClientView, err error) {
	const tag = tagPath + "/ReadById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientReadById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read client by id", domainmodelslogger.Meta{
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

	value, err := u.client.ReadById(ctx, spaceId, id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read client by id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"id":                     id,
		})

		return nil, err
	}

	result = applicationshared.ClientView(value)

	return result, nil
}

func (u *usecase) ReadBySlug(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, selector string) (result *domainusecasesshared.ClientView, err error) {
	const tag = tagPath + "/ReadBySlug"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientReadBySlug}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read client by slug", domainmodelslogger.Meta{
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

	value, err := u.client.ReadBySlug(ctx, selector)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read client by slug", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return nil, err
	}

	if value == nil || value.SpaceId != spaceId {
		failure := domainmodelserror.NewError("client not found", domainmodelserror.ClientNotFound, nil)
		u.logger.Error(ctx, tag, "client not found", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    failure,
			"space_id":               spaceId,
		})

		return nil, failure
	}

	result = applicationshared.ClientView(value)

	return result, nil
}

func (u *usecase) ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.FilterClientRequest) (result domainusecasesshared.Page[domainusecasesshared.ClientView], err error) {
	const tag = tagPath + "/ReadByFilter"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientReadByFilter}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to list clients", domainmodelslogger.Meta{
			"err": err,
		})

		return domainusecasesshared.Page[domainusecasesshared.ClientView]{}, err
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

		return domainusecasesshared.Page[domainusecasesshared.ClientView]{}, err
	}

	if err := applicationshared.ValidatePagination(request.Page, request.Limit); err != nil {
		u.logger.Error(ctx, tag, "invalid client pagination", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return domainusecasesshared.Page[domainusecasesshared.ClientView]{}, err
	}

	items, total, err := u.client.ReadByFilter(ctx, spaceId, domaincontractsrepository.ClientFilter{
		Page:     request.Page,
		Limit:    request.Limit,
		Search:   request.Search,
		IsActive: request.IsActive,
	})
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list clients", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
		})

		return domainusecasesshared.Page[domainusecasesshared.ClientView]{}, err
	}

	if items == nil {
		items = []domainusecasesshared.ClientView{}
	}

	result = domainusecasesshared.Page[domainusecasesshared.ClientView]{Items: items, Total: total}

	return result, nil
}

func (u *usecase) UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64, request domainusecasesmanagement.UpdateClientRequest) (err error) {
	const tag = tagPath + "/UpdateById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientUpdateById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to update by id client", domainmodelslogger.Meta{
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
			if err := u.validator.ClientName(*request.Name); err != nil {
				u.logger.Error(ctx, tag, "invalid client name", domainmodelslogger.Meta{
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
			if err := u.validator.ClientDesc(*request.Description); err != nil {
				u.logger.Error(ctx, tag, "invalid client description", domainmodelslogger.Meta{
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
			u.logger.Error(ctx, tag, "invalid client preferences JSON", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
				"id":                     id,
			})

			return err
		}

		if err := u.client.UpdateById(ctx, spaceId, id, domaincontractsrepository.UpdateClientArgs{
			Name:        request.Name,
			Description: request.Description,
			IsActive:    request.IsActive,
			Preferences: request.Preferences,
			By:          &caller.UserId,
		}); err != nil {
			u.logger.Error(ctx, tag, "failed to update client", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete client update transaction", domainmodelslogger.Meta{
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
		u.logger.Info(ctx, tag, "updated client", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"id":                     id,
		})
	}

	return nil
}

func (u *usecase) RotateSecret(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) (result *domainusecasesmanagement.ClientSecretResult, err error) {
	const tag = tagPath + "/RotateSecret"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientRotateSecret}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to rotate secret client", domainmodelslogger.Meta{
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

		raw, hash, redacted, err := u.keygen.Generate()
		if err != nil {
			u.logger.Error(ctx, tag, "failed to generate client secret", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
				"id":                     id,
			})

			return err
		}

		if err := u.client.UpdateById(ctx, spaceId, id, domaincontractsrepository.UpdateClientArgs{
			SecretKeyHash:     &hash,
			SecretKeyRedacted: &redacted,
			By:                &caller.UserId,
		}); err != nil {
			u.logger.Error(ctx, tag, "failed to update client", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
				"id":                     id,
			})

			return err
		}

		value, err := u.client.ReadById(ctx, spaceId, id)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read client by id", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    err,
				"space_id":               spaceId,
				"id":                     id,
			})

			return err
		}

		if value == nil {
			failure := domainmodelserror.NewError("client result missing", domainmodelserror.Failure, nil)
			u.logger.Error(ctx, tag, "client result missing", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"err":                    failure,
				"space_id":               spaceId,
				"id":                     id,
			})

			return failure
		}

		result = &domainusecasesmanagement.ClientSecretResult{
			Client: *applicationshared.ClientView(value),
			Secret: raw,
		}

		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete client secret rotation transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"id":                     id,
		})

		return nil, err
	}

	u.logger.Info(ctx, tag, "rotated secret for client", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"space_id":               spaceId,
		"id":                     id,
	})

	return result, nil
}

func (u *usecase) DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) (err error) {
	const tag = tagPath + "/DeleteById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientDeleteById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to delete by id client", domainmodelslogger.Meta{
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

		if err := u.client.DeleteById(ctx, spaceId, id, &caller.UserId); err != nil {
			u.logger.Error(ctx, tag, "failed to delete client by id", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete client deletion transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"err":                    err,
			"space_id":               spaceId,
			"id":                     id,
		})

		return err
	}

	u.logger.Info(ctx, tag, "deleted client", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"space_id":               spaceId,
		"id":                     id,
	})

	return nil
}

func (u *usecase) CreateOrigin(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64, request domainusecasesmanagement.CreateClientOriginRequest) (int64, error) {
	const tag = tagPath + "/CreateOrigin"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientCreateOrigin}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to create client origin", domainmodelslogger.Meta{
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

	var id int64
	err := u.transactor.WithTx(ctx, func(ctx context.Context) error {
		if _, err := u.identity.Management(ctx, caller, &spaceId); err != nil {
			u.logger.Error(ctx, tag, "failed to authorize management scope to create client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"err":                    err,
			})

			return err
		}

		if err := applicationshared.ValidateIds(spaceId, clientId); err != nil {
			u.logger.Error(ctx, tag, "invalid identifiers to create client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"err":                    err,
			})

			return err
		}

		if err := u.validator.ClientOrigin(request.Origin); err != nil {
			u.logger.Error(ctx, tag, "invalid origin for client origin creation", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"err":                    err,
			})

			return err
		}

		client, err := u.client.ReadById(ctx, spaceId, clientId)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read client to create client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"err":                    err,
			})

			return err
		}

		if client == nil {
			err := domainmodelserror.NewError("client result missing while attempting to create client origin", domainmodelserror.Failure, nil)
			u.logger.Error(ctx, tag, "client result missing while attempting to create client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"err":                    err,
			})

			return err
		}

		if client.Id != clientId || client.SpaceId != spaceId || client.AuditDelete.At != nil {
			err := domainmodelserror.NewError("client not found while attempting to create client origin", domainmodelserror.ClientNotFound, nil)
			u.logger.Error(ctx, tag, "client not found while attempting to create client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"err":                    err,
			})

			return err
		}

		id, err = u.clientOrigin.Create(ctx, spaceId, domaincontractsrepository.CreateClientOriginArgs{
			ClientId: clientId,
			Origin:   request.Origin,
			IsActive: request.IsActive,
			By:       &caller.UserId,
		})
		if err != nil {
			u.logger.Error(ctx, tag, "failed to create client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"err":                    err,
			})

			return err
		}

		if id <= 0 {
			err := domainmodelserror.NewError("client origin creation returned invalid identifier", domainmodelserror.Failure, nil)
			u.logger.Error(ctx, tag, "client origin creation returned invalid identifier", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"err":                    err,
			})

			return err
		}

		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete client origin creation transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return 0, err
	}

	u.logger.Info(ctx, tag, "created client origin", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"space_id":               spaceId,
		"client_id":              clientId,
		"id":                     id,
	})

	return id, nil
}

func (u *usecase) ReadOriginById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId, id int64) (*domainmodelsentity.ClientOrigin, error) {
	const tag = tagPath + "/ReadOriginById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientReadOriginById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read client origin by id", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to authorize management scope to read client origin by id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"id":                     id,
			"err":                    err,
		})

		return nil, err
	}

	if err := applicationshared.ValidateIds(spaceId, clientId, id); err != nil {
		u.logger.Error(ctx, tag, "invalid identifiers to read client origin by id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"id":                     id,
			"err":                    err,
		})

		return nil, err
	}

	client, err := u.client.ReadById(ctx, spaceId, clientId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read client to read client origin by id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"id":                     id,
			"err":                    err,
		})

		return nil, err
	}

	if client == nil {
		err := domainmodelserror.NewError("client result missing while attempting to read client origin by id", domainmodelserror.Failure, nil)
		u.logger.Error(ctx, tag, "client result missing while attempting to read client origin by id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"id":                     id,
			"err":                    err,
		})

		return nil, err
	}

	if client.Id != clientId || client.SpaceId != spaceId || client.AuditDelete.At != nil {
		err := domainmodelserror.NewError("client not found while attempting to read client origin by id", domainmodelserror.ClientNotFound, nil)
		u.logger.Error(ctx, tag, "client not found while attempting to read client origin by id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"id":                     id,
			"err":                    err,
		})

		return nil, err
	}

	item, err := u.clientOrigin.ReadById(ctx, spaceId, id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read origin record to read client origin by id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"id":                     id,
			"err":                    err,
		})

		return nil, err
	}

	if item == nil {
		err := domainmodelserror.NewError("client origin result missing while attempting to read client origin by id", domainmodelserror.Failure, nil)
		u.logger.Error(ctx, tag, "client origin result missing while attempting to read client origin by id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"id":                     id,
			"err":                    err,
		})

		return nil, err
	}

	if item.Id != id || item.ClientId != clientId || item.AuditDelete.At != nil {
		err := domainmodelserror.NewError("client origin not found for requested client while attempting to read client origin by id", domainmodelserror.ClientOriginNotFound, nil)
		u.logger.Error(ctx, tag, "client origin not found for requested client while attempting to read client origin by id", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"id":                     id,
			"err":                    err,
		})

		return nil, err
	}

	return item, nil
}

func (u *usecase) ReadOriginByValue(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64, origin string) (*domainmodelsentity.ClientOrigin, error) {
	const tag = tagPath + "/ReadOriginByValue"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientReadOriginByValue}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read client origin by value", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to authorize management scope to read client origin by value", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	if err := applicationshared.ValidateIds(spaceId, clientId); err != nil {
		u.logger.Error(ctx, tag, "invalid identifiers to read client origin by value", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	if err := u.validator.ClientOrigin(origin); err != nil {
		u.logger.Error(ctx, tag, "invalid origin for client origin lookup", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	client, err := u.client.ReadById(ctx, spaceId, clientId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read client to read client origin by value", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	if client == nil {
		err := domainmodelserror.NewError("client result missing while attempting to read client origin by value", domainmodelserror.Failure, nil)
		u.logger.Error(ctx, tag, "client result missing while attempting to read client origin by value", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	if client.Id != clientId || client.SpaceId != spaceId || client.AuditDelete.At != nil {
		err := domainmodelserror.NewError("client not found while attempting to read client origin by value", domainmodelserror.ClientNotFound, nil)
		u.logger.Error(ctx, tag, "client not found while attempting to read client origin by value", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	item, err := u.clientOrigin.ReadByClientIdAndOrigin(ctx, spaceId, clientId, origin)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read client origin by value", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	if item == nil {
		err := domainmodelserror.NewError("client origin result missing during value lookup", domainmodelserror.Failure, nil)
		u.logger.Error(ctx, tag, "client origin result missing during value lookup", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	if item.ClientId != clientId || item.Origin != origin || item.AuditDelete.At != nil {
		err := domainmodelserror.NewError("client origin not found for requested client and value", domainmodelserror.ClientOriginNotFound, nil)
		u.logger.Error(ctx, tag, "client origin not found for requested client and value", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	return item, nil
}

func (u *usecase) ReadOriginsByClientId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64) ([]domainmodelsentity.ClientOrigin, error) {
	const tag = tagPath + "/ReadOriginsByClientId"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientReadOriginsByClientId}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to list client origins", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to authorize management scope to list client origins", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	if err := applicationshared.ValidateIds(spaceId, clientId); err != nil {
		u.logger.Error(ctx, tag, "invalid identifiers to list client origins", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	client, err := u.client.ReadById(ctx, spaceId, clientId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read client to list client origins", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	if client == nil {
		err := domainmodelserror.NewError("client result missing while attempting to list client origins", domainmodelserror.Failure, nil)
		u.logger.Error(ctx, tag, "client result missing while attempting to list client origins", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	if client.Id != clientId || client.SpaceId != spaceId || client.AuditDelete.At != nil {
		err := domainmodelserror.NewError("client not found while attempting to list client origins", domainmodelserror.ClientNotFound, nil)
		u.logger.Error(ctx, tag, "client not found while attempting to list client origins", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	items, err := u.clientOrigin.ReadByClientId(ctx, spaceId, clientId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list client origins", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	if items == nil {
		items = make([]domainmodelsentity.ClientOrigin, 0)
	}

	return items, nil
}

func (u *usecase) ReadOriginsByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64, request domainusecasesmanagement.FilterClientOriginRequest) (domainusecasesshared.Page[domainmodelsentity.ClientOrigin], error) {
	const tag = tagPath + "/ReadOriginsByFilter"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientReadOriginsByFilter}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to filter client origins", domainmodelslogger.Meta{
			"err": err,
		})

		return domainusecasesshared.Page[domainmodelsentity.ClientOrigin]{}, err
	}

	callerUserId, callerSpaceId, callerSpaceMemberId := int64(0), int64(0), int64(0)
	if caller != nil {
		callerUserId = caller.UserId
		callerSpaceId = caller.SpaceId
		callerSpaceMemberId = caller.SpaceMemberId
	}

	if _, err := u.identity.Management(ctx, caller, &spaceId); err != nil {
		u.logger.Error(ctx, tag, "failed to authorize management scope to filter client origins", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainmodelsentity.ClientOrigin]{}, err
	}

	if err := applicationshared.ValidateIds(spaceId, clientId); err != nil {
		u.logger.Error(ctx, tag, "invalid identifiers to filter client origins", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainmodelsentity.ClientOrigin]{}, err
	}

	if err := applicationshared.ValidatePagination(request.Page, request.Limit); err != nil {
		u.logger.Error(ctx, tag, "invalid client origin pagination", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainmodelsentity.ClientOrigin]{}, err
	}

	client, err := u.client.ReadById(ctx, spaceId, clientId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read client to filter client origins", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainmodelsentity.ClientOrigin]{}, err
	}

	if client == nil {
		err := domainmodelserror.NewError("client result missing while attempting to filter client origins", domainmodelserror.Failure, nil)
		u.logger.Error(ctx, tag, "client result missing while attempting to filter client origins", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainmodelsentity.ClientOrigin]{}, err
	}

	if client.Id != clientId || client.SpaceId != spaceId || client.AuditDelete.At != nil {
		err := domainmodelserror.NewError("client not found while attempting to filter client origins", domainmodelserror.ClientNotFound, nil)
		u.logger.Error(ctx, tag, "client not found while attempting to filter client origins", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainmodelsentity.ClientOrigin]{}, err
	}

	items, total, err := u.clientOrigin.ReadByFilter(ctx, spaceId, domaincontractsrepository.ClientOriginFilter{
		Page:     request.Page,
		Limit:    request.Limit,
		ClientId: &clientId,
		Search:   request.Search,
		IsActive: request.IsActive,
	})
	if err != nil {
		u.logger.Error(ctx, tag, "failed to filter client origins", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainmodelsentity.ClientOrigin]{}, err
	}

	if items == nil {
		items = make([]domainmodelsentity.ClientOrigin, 0)
	}

	return domainusecasesshared.Page[domainmodelsentity.ClientOrigin]{Items: items, Total: total}, nil
}

func (u *usecase) UpdateOriginById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId, id int64, request domainusecasesmanagement.UpdateClientOriginRequest) error {
	const tag = tagPath + "/UpdateOriginById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientUpdateOriginById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to update client origin", domainmodelslogger.Meta{
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

	err := u.transactor.WithTx(ctx, func(ctx context.Context) error {
		if _, err := u.identity.Management(ctx, caller, &spaceId); err != nil {
			u.logger.Error(ctx, tag, "failed to authorize management scope to update client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		if err := applicationshared.ValidateIds(spaceId, clientId, id); err != nil {
			u.logger.Error(ctx, tag, "invalid identifiers to update client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		if request.Origin != nil {
			if err := u.validator.ClientOrigin(*request.Origin); err != nil {
				u.logger.Error(ctx, tag, "invalid replacement client origin", domainmodelslogger.Meta{
					"caller_user_id":         callerUserId,
					"caller_space_id":        callerSpaceId,
					"caller_space_member_id": callerSpaceMemberId,
					"space_id":               spaceId,
					"client_id":              clientId,
					"id":                     id,
					"err":                    err,
				})

				return err
			}

		}

		client, err := u.client.ReadById(ctx, spaceId, clientId)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read client to update client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		if client == nil {
			err := domainmodelserror.NewError("client result missing while attempting to update client origin", domainmodelserror.Failure, nil)
			u.logger.Error(ctx, tag, "client result missing while attempting to update client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		if client.Id != clientId || client.SpaceId != spaceId || client.AuditDelete.At != nil {
			err := domainmodelserror.NewError("client not found while attempting to update client origin", domainmodelserror.ClientNotFound, nil)
			u.logger.Error(ctx, tag, "client not found while attempting to update client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		item, err := u.clientOrigin.ReadById(ctx, spaceId, id)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read origin record to update client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		if item == nil {
			err := domainmodelserror.NewError("client origin result missing while attempting to update client origin", domainmodelserror.Failure, nil)
			u.logger.Error(ctx, tag, "client origin result missing while attempting to update client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		if item.Id != id || item.ClientId != clientId || item.AuditDelete.At != nil {
			err := domainmodelserror.NewError("client origin not found for requested client while attempting to update client origin", domainmodelserror.ClientOriginNotFound, nil)
			u.logger.Error(ctx, tag, "client origin not found for requested client while attempting to update client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		if err := u.clientOrigin.UpdateById(ctx, spaceId, id, domaincontractsrepository.UpdateClientOriginArgs{Origin: request.Origin, IsActive: request.IsActive, By: &caller.UserId}); err != nil {
			u.logger.Error(ctx, tag, "failed to update client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete client origin update transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"id":                     id,
			"err":                    err,
		})

		return err
	}

	if request.Origin != nil || request.IsActive != nil {
		u.logger.Info(ctx, tag, "updated client origin", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"id":                     id,
		})
	}

	return nil
}

func (u *usecase) DeleteOriginById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId, id int64) error {
	const tag = tagPath + "/DeleteOriginById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientDeleteOriginById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to delete client origin", domainmodelslogger.Meta{
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

	err := u.transactor.WithTx(ctx, func(ctx context.Context) error {
		if _, err := u.identity.Management(ctx, caller, &spaceId); err != nil {
			u.logger.Error(ctx, tag, "failed to authorize management scope to delete client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		if err := applicationshared.ValidateIds(spaceId, clientId, id); err != nil {
			u.logger.Error(ctx, tag, "invalid identifiers to delete client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		client, err := u.client.ReadById(ctx, spaceId, clientId)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read client to delete client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		if client == nil {
			err := domainmodelserror.NewError("client result missing while attempting to delete client origin", domainmodelserror.Failure, nil)
			u.logger.Error(ctx, tag, "client result missing while attempting to delete client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		if client.Id != clientId || client.SpaceId != spaceId || client.AuditDelete.At != nil {
			err := domainmodelserror.NewError("client not found while attempting to delete client origin", domainmodelserror.ClientNotFound, nil)
			u.logger.Error(ctx, tag, "client not found while attempting to delete client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		item, err := u.clientOrigin.ReadById(ctx, spaceId, id)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read origin record to delete client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		if item == nil {
			err := domainmodelserror.NewError("client origin result missing while attempting to delete client origin", domainmodelserror.Failure, nil)
			u.logger.Error(ctx, tag, "client origin result missing while attempting to delete client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		if item.Id != id || item.ClientId != clientId || item.AuditDelete.At != nil {
			err := domainmodelserror.NewError("client origin not found for requested client while attempting to delete client origin", domainmodelserror.ClientOriginNotFound, nil)
			u.logger.Error(ctx, tag, "client origin not found for requested client while attempting to delete client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		if err := u.clientOrigin.DeleteById(ctx, spaceId, id, &caller.UserId); err != nil {
			u.logger.Error(ctx, tag, "failed to delete client origin", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"id":                     id,
				"err":                    err,
			})

			return err
		}

		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete client origin deletion transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"id":                     id,
			"err":                    err,
		})

		return err
	}

	u.logger.Info(ctx, tag, "deleted client origin", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"space_id":               spaceId,
		"client_id":              clientId,
		"id":                     id,
	})

	return nil
}
