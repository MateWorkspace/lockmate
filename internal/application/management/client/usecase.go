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

// loopbackCidrs are always permitted for local administration.
var loopbackCidrs = []string{"127.0.0.1/32", "::1/128"}

type usecase struct {
	user              domaincontractsrepository.User
	space             domaincontractsrepository.Space
	spaceMember       domaincontractsrepository.SpaceMember
	client            domaincontractsrepository.Client
	clientIpAllowlist domaincontractsrepository.ClientIpAllowlist
	guard             domaincontractsutility.PermissionGuard
	validator         domaincontractsutility.Validator
	logger            domaincontractsutility.Logger
	transactor        domaincontractsutility.Transactor
	keygen            domaincontractsutility.Keygen
	identity          *applicationshared.IdentityReader
}

func NewUsecaseImpl(
	user domaincontractsrepository.User,
	space domaincontractsrepository.Space,
	spaceMember domaincontractsrepository.SpaceMember,
	client domaincontractsrepository.Client,
	clientIpAllowlist domaincontractsrepository.ClientIpAllowlist,
	guard domaincontractsutility.PermissionGuard,
	validator domaincontractsutility.Validator,
	logger domaincontractsutility.Logger,
	transactor domaincontractsutility.Transactor,
	keygen domaincontractsutility.Keygen,
	controlSpaceSlug string,
) domainusecasesmanagement.Client {
	return &usecase{
		user:              user,
		space:             space,
		spaceMember:       spaceMember,
		client:            client,
		clientIpAllowlist: clientIpAllowlist,
		guard:             guard,
		validator:         validator,
		logger:            logger,
		transactor:        transactor,
		keygen:            keygen,
		identity:          applicationshared.NewIdentityReader(user, space, spaceMember, controlSpaceSlug),
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

		if request.AuthRedirectUri != nil {
			if err := u.validator.ClientAuthRedirectUri(*request.AuthRedirectUri); err != nil {
				u.logger.Error(ctx, tag, "invalid client authentication redirect URI", domainmodelslogger.Meta{"err": err})

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
			AuthRedirectUri:   request.AuthRedirectUri,
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

		// Loopback stays reachable for local administration regardless of the
		// configured allowlist; only loopback is permitted until more entries
		// are added.
		for _, cidr := range loopbackCidrs {
			if _, err := u.clientIpAllowlist.Create(ctx, spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{
				ClientId: id,
				Cidr:     cidr,
				By:       &caller.UserId,
			}); err != nil {
				u.logger.Error(ctx, tag, "failed to create client loopback IP allowlist entry", domainmodelslogger.Meta{
					"caller_user_id":         callerUserId,
					"caller_space_id":        callerSpaceId,
					"caller_space_member_id": callerSpaceMemberId,
					"err":                    err,
					"space_id":               spaceId,
					"client_id":              id,
					"cidr":                   cidr,
				})

				return err
			}
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
		"id":                     result.Client.Id,
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

		if request.AuthRedirectUri != nil && request.SetAuthRedirectUriNull {
			failure := domainmodelserror.NewError("redirect replacement and clearing cannot be combined", domainmodelserror.BadArgs, nil)
			u.logger.Error(ctx, tag, "conflicting client redirect update", domainmodelslogger.Meta{"err": failure})

			return failure
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

		if request.AuthRedirectUri != nil {
			if err := u.validator.ClientAuthRedirectUri(*request.AuthRedirectUri); err != nil {
				u.logger.Error(ctx, tag, "invalid client authentication redirect URI", domainmodelslogger.Meta{"err": err})

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
			AuthRedirectUri:        request.AuthRedirectUri,
			SetAuthRedirectUriNull: request.SetAuthRedirectUriNull,
			Name:                   request.Name,
			Description:            request.Description,
			IsActive:               request.IsActive,
			Preferences:            request.Preferences,
			By:                     &caller.UserId,
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

	if request.AuthRedirectUri != nil || request.SetAuthRedirectUriNull {
		u.logger.Info(ctx, tag, "updated client authentication redirect", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"id":                     id,
		})
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

func (u *usecase) CreateIpAllowlist(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64, request domainusecasesmanagement.CreateClientIpAllowlistRequest) (int64, error) {
	const tag = tagPath + "/CreateIpAllowlist"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientCreateIpAllowlist}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to create client IP allowlist", domainmodelslogger.Meta{
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
			u.logger.Error(ctx, tag, "failed to authorize management scope to create client IP allowlist", domainmodelslogger.Meta{
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
			u.logger.Error(ctx, tag, "invalid identifiers to create client IP allowlist", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"err":                    err,
			})

			return err
		}

		if err := u.validator.ClientIpAllowlist(request.Cidr); err != nil {
			u.logger.Error(ctx, tag, "invalid cidr for client IP allowlist creation", domainmodelslogger.Meta{
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
			u.logger.Error(ctx, tag, "failed to read client to create client IP allowlist", domainmodelslogger.Meta{
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
			err := domainmodelserror.NewError("client result missing while attempting to create client IP allowlist", domainmodelserror.Failure, nil)
			u.logger.Error(ctx, tag, "client result missing while attempting to create client IP allowlist", domainmodelslogger.Meta{
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
			err := domainmodelserror.NewError("client not found while attempting to create client IP allowlist", domainmodelserror.ClientNotFound, nil)
			u.logger.Error(ctx, tag, "client not found while attempting to create client IP allowlist", domainmodelslogger.Meta{
				"caller_user_id":         callerUserId,
				"caller_space_id":        callerSpaceId,
				"caller_space_member_id": callerSpaceMemberId,
				"space_id":               spaceId,
				"client_id":              clientId,
				"err":                    err,
			})

			return err
		}

		id, err = u.clientIpAllowlist.Create(ctx, spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{
			ClientId: clientId,
			Cidr:     request.Cidr,
			IsActive: request.IsActive,
			By:       &caller.UserId,
		})
		if err != nil {
			u.logger.Error(ctx, tag, "failed to create client IP allowlist", domainmodelslogger.Meta{
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
			err := domainmodelserror.NewError("client IP allowlist creation returned invalid identifier", domainmodelserror.Failure, nil)
			u.logger.Error(ctx, tag, "client IP allowlist creation returned invalid identifier", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete client IP allowlist creation transaction", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return 0, err
	}

	u.logger.Info(ctx, tag, "created client IP allowlist", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"space_id":               spaceId,
		"client_id":              clientId,
		"id":                     id,
	})

	return id, nil
}

func (u *usecase) ReadIpAllowlistById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId, id int64) (*domainmodelsentity.ClientIpAllowlist, error) {
	const tag = tagPath + "/ReadIpAllowlistById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientReadIpAllowlistById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read client IP allowlist by id", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to authorize management scope to read client IP allowlist by id", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "invalid identifiers to read client IP allowlist by id", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to read client to read client IP allowlist by id", domainmodelslogger.Meta{
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
		err := domainmodelserror.NewError("client result missing while attempting to read client IP allowlist by id", domainmodelserror.Failure, nil)
		u.logger.Error(ctx, tag, "client result missing while attempting to read client IP allowlist by id", domainmodelslogger.Meta{
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
		err := domainmodelserror.NewError("client not found while attempting to read client IP allowlist by id", domainmodelserror.ClientNotFound, nil)
		u.logger.Error(ctx, tag, "client not found while attempting to read client IP allowlist by id", domainmodelslogger.Meta{
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

	item, err := u.clientIpAllowlist.ReadById(ctx, spaceId, id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read IP allowlist entry to read client IP allowlist by id", domainmodelslogger.Meta{
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
		err := domainmodelserror.NewError("client IP allowlist result missing while attempting to read client IP allowlist by id", domainmodelserror.Failure, nil)
		u.logger.Error(ctx, tag, "client IP allowlist result missing while attempting to read client IP allowlist by id", domainmodelslogger.Meta{
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
		err := domainmodelserror.NewError("client IP allowlist not found for requested client while attempting to read client IP allowlist by id", domainmodelserror.ClientIpAllowlistNotFound, nil)
		u.logger.Error(ctx, tag, "client IP allowlist not found for requested client while attempting to read client IP allowlist by id", domainmodelslogger.Meta{
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

func (u *usecase) ReadIpAllowlistByCidr(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64, cidr string) (*domainmodelsentity.ClientIpAllowlist, error) {
	const tag = tagPath + "/ReadIpAllowlistByCidr"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientReadIpAllowlistByCidr}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to read client IP allowlist by value", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to authorize management scope to read client IP allowlist by value", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "invalid identifiers to read client IP allowlist by value", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	if err := u.validator.ClientIpAllowlist(cidr); err != nil {
		u.logger.Error(ctx, tag, "invalid cidr for client IP allowlist lookup", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to read client to read client IP allowlist by value", domainmodelslogger.Meta{
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
		err := domainmodelserror.NewError("client result missing while attempting to read client IP allowlist by value", domainmodelserror.Failure, nil)
		u.logger.Error(ctx, tag, "client result missing while attempting to read client IP allowlist by value", domainmodelslogger.Meta{
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
		err := domainmodelserror.NewError("client not found while attempting to read client IP allowlist by value", domainmodelserror.ClientNotFound, nil)
		u.logger.Error(ctx, tag, "client not found while attempting to read client IP allowlist by value", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	item, err := u.clientIpAllowlist.ReadByClientIdAndCidr(ctx, spaceId, clientId, cidr)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read client IP allowlist by value", domainmodelslogger.Meta{
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
		err := domainmodelserror.NewError("client IP allowlist result missing during value lookup", domainmodelserror.Failure, nil)
		u.logger.Error(ctx, tag, "client IP allowlist result missing during value lookup", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return nil, err
	}

	if item.ClientId != clientId || item.Cidr != cidr || item.AuditDelete.At != nil {
		err := domainmodelserror.NewError("client IP allowlist not found for requested client and value", domainmodelserror.ClientIpAllowlistNotFound, nil)
		u.logger.Error(ctx, tag, "client IP allowlist not found for requested client and value", domainmodelslogger.Meta{
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

func (u *usecase) ReadIpAllowlistsByClientId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64) ([]domainmodelsentity.ClientIpAllowlist, error) {
	const tag = tagPath + "/ReadIpAllowlistsByClientId"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientReadIpAllowlistsByClientId}); err != nil {
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

	items, err := u.clientIpAllowlist.ReadByClientId(ctx, spaceId, clientId)
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
		items = make([]domainmodelsentity.ClientIpAllowlist, 0)
	}

	return items, nil
}

func (u *usecase) ReadIpAllowlistsByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64, request domainusecasesmanagement.FilterClientIpAllowlistRequest) (domainusecasesshared.Page[domainmodelsentity.ClientIpAllowlist], error) {
	const tag = tagPath + "/ReadIpAllowlistsByFilter"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientReadIpAllowlistsByFilter}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to filter client origins", domainmodelslogger.Meta{
			"err": err,
		})

		return domainusecasesshared.Page[domainmodelsentity.ClientIpAllowlist]{}, err
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

		return domainusecasesshared.Page[domainmodelsentity.ClientIpAllowlist]{}, err
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

		return domainusecasesshared.Page[domainmodelsentity.ClientIpAllowlist]{}, err
	}

	if err := applicationshared.ValidatePagination(request.Page, request.Limit); err != nil {
		u.logger.Error(ctx, tag, "invalid client IP allowlist pagination", domainmodelslogger.Meta{
			"caller_user_id":         callerUserId,
			"caller_space_id":        callerSpaceId,
			"caller_space_member_id": callerSpaceMemberId,
			"space_id":               spaceId,
			"client_id":              clientId,
			"err":                    err,
		})

		return domainusecasesshared.Page[domainmodelsentity.ClientIpAllowlist]{}, err
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

		return domainusecasesshared.Page[domainmodelsentity.ClientIpAllowlist]{}, err
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

		return domainusecasesshared.Page[domainmodelsentity.ClientIpAllowlist]{}, err
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

		return domainusecasesshared.Page[domainmodelsentity.ClientIpAllowlist]{}, err
	}

	items, total, err := u.clientIpAllowlist.ReadByFilter(ctx, spaceId, domaincontractsrepository.ClientIpAllowlistFilter{
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

		return domainusecasesshared.Page[domainmodelsentity.ClientIpAllowlist]{}, err
	}

	if items == nil {
		items = make([]domainmodelsentity.ClientIpAllowlist, 0)
	}

	return domainusecasesshared.Page[domainmodelsentity.ClientIpAllowlist]{Items: items, Total: total}, nil
}

func (u *usecase) UpdateIpAllowlistById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId, id int64, request domainusecasesmanagement.UpdateClientIpAllowlistRequest) error {
	const tag = tagPath + "/UpdateIpAllowlistById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientUpdateIpAllowlistById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to update client IP allowlist", domainmodelslogger.Meta{
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
			u.logger.Error(ctx, tag, "failed to authorize management scope to update client IP allowlist", domainmodelslogger.Meta{
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
			u.logger.Error(ctx, tag, "invalid identifiers to update client IP allowlist", domainmodelslogger.Meta{
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

		if request.Cidr != nil {
			if err := u.validator.ClientIpAllowlist(*request.Cidr); err != nil {
				u.logger.Error(ctx, tag, "invalid replacement client IP allowlist", domainmodelslogger.Meta{
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
			u.logger.Error(ctx, tag, "failed to read client to update client IP allowlist", domainmodelslogger.Meta{
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
			err := domainmodelserror.NewError("client result missing while attempting to update client IP allowlist", domainmodelserror.Failure, nil)
			u.logger.Error(ctx, tag, "client result missing while attempting to update client IP allowlist", domainmodelslogger.Meta{
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
			err := domainmodelserror.NewError("client not found while attempting to update client IP allowlist", domainmodelserror.ClientNotFound, nil)
			u.logger.Error(ctx, tag, "client not found while attempting to update client IP allowlist", domainmodelslogger.Meta{
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

		item, err := u.clientIpAllowlist.ReadById(ctx, spaceId, id)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read IP allowlist entry to update client IP allowlist", domainmodelslogger.Meta{
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
			err := domainmodelserror.NewError("client IP allowlist result missing while attempting to update client IP allowlist", domainmodelserror.Failure, nil)
			u.logger.Error(ctx, tag, "client IP allowlist result missing while attempting to update client IP allowlist", domainmodelslogger.Meta{
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
			err := domainmodelserror.NewError("client IP allowlist not found for requested client while attempting to update client IP allowlist", domainmodelserror.ClientIpAllowlistNotFound, nil)
			u.logger.Error(ctx, tag, "client IP allowlist not found for requested client while attempting to update client IP allowlist", domainmodelslogger.Meta{
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

		if err := u.clientIpAllowlist.UpdateById(ctx, spaceId, id, domaincontractsrepository.UpdateClientIpAllowlistArgs{Cidr: request.Cidr, IsActive: request.IsActive, By: &caller.UserId}); err != nil {
			u.logger.Error(ctx, tag, "failed to update client IP allowlist", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete client IP allowlist update transaction", domainmodelslogger.Meta{
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

	if request.Cidr != nil || request.IsActive != nil {
		u.logger.Info(ctx, tag, "updated client IP allowlist", domainmodelslogger.Meta{
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

func (u *usecase) DeleteIpAllowlistById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId, id int64) error {
	const tag = tagPath + "/DeleteIpAllowlistById"

	if err := u.guard.Check(caller, []string{applicationpermissions.ManagementClientDeleteIpAllowlistById}); err != nil {
		u.logger.Error(ctx, tag, "permission denied to delete client IP allowlist", domainmodelslogger.Meta{
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
			u.logger.Error(ctx, tag, "failed to authorize management scope to delete client IP allowlist", domainmodelslogger.Meta{
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
			u.logger.Error(ctx, tag, "invalid identifiers to delete client IP allowlist", domainmodelslogger.Meta{
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
			u.logger.Error(ctx, tag, "failed to read client to delete client IP allowlist", domainmodelslogger.Meta{
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
			err := domainmodelserror.NewError("client result missing while attempting to delete client IP allowlist", domainmodelserror.Failure, nil)
			u.logger.Error(ctx, tag, "client result missing while attempting to delete client IP allowlist", domainmodelslogger.Meta{
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
			err := domainmodelserror.NewError("client not found while attempting to delete client IP allowlist", domainmodelserror.ClientNotFound, nil)
			u.logger.Error(ctx, tag, "client not found while attempting to delete client IP allowlist", domainmodelslogger.Meta{
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

		item, err := u.clientIpAllowlist.ReadById(ctx, spaceId, id)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read IP allowlist entry to delete client IP allowlist", domainmodelslogger.Meta{
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
			err := domainmodelserror.NewError("client IP allowlist result missing while attempting to delete client IP allowlist", domainmodelserror.Failure, nil)
			u.logger.Error(ctx, tag, "client IP allowlist result missing while attempting to delete client IP allowlist", domainmodelslogger.Meta{
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
			err := domainmodelserror.NewError("client IP allowlist not found for requested client while attempting to delete client IP allowlist", domainmodelserror.ClientIpAllowlistNotFound, nil)
			u.logger.Error(ctx, tag, "client IP allowlist not found for requested client while attempting to delete client IP allowlist", domainmodelslogger.Meta{
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

		if err := u.clientIpAllowlist.DeleteById(ctx, spaceId, id, &caller.UserId); err != nil {
			u.logger.Error(ctx, tag, "failed to delete client IP allowlist", domainmodelslogger.Meta{
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
		u.logger.Error(ctx, tag, "failed to complete client IP allowlist deletion transaction", domainmodelslogger.Meta{
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

	u.logger.Info(ctx, tag, "deleted client IP allowlist", domainmodelslogger.Meta{
		"caller_user_id":         callerUserId,
		"caller_space_id":        callerSpaceId,
		"caller_space_member_id": callerSpaceMemberId,
		"space_id":               spaceId,
		"client_id":              clientId,
		"id":                     id,
	})

	return nil
}
