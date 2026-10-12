package applicationseeder

import (
	"context"
	"errors"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	domainmodelsseeder "github.com/MateWorkspace/lockmate/internal/domain/models/seeder"
	domainusecasesseeder "github.com/MateWorkspace/lockmate/internal/domain/usecases/seeder"
)

const tagPath = "seeder/seeder"

type usecase struct {
	space             domaincontractsrepository.Space
	permission        domaincontractsrepository.Permission
	role              domaincontractsrepository.Role
	rolePermission    domaincontractsrepository.RolePermission
	user              domaincontractsrepository.User
	classicAuth       domaincontractsrepository.ClassicAuth
	spaceMember       domaincontractsrepository.SpaceMember
	spaceMemberRole   domaincontractsrepository.SpaceMemberRole
	client            domaincontractsrepository.Client
	clientIpAllowlist domaincontractsrepository.ClientIpAllowlist
	password          domaincontractsutility.Password
	keygen            domaincontractsutility.Keygen
	validator         domaincontractsutility.Validator
	transactor        domaincontractsutility.Transactor
	logger            domaincontractsutility.Logger
	data              *domainmodelsseeder.Data
}

func NewUsecaseImpl(
	space domaincontractsrepository.Space,
	permission domaincontractsrepository.Permission,
	role domaincontractsrepository.Role,
	rolePermission domaincontractsrepository.RolePermission,
	user domaincontractsrepository.User,
	classicAuth domaincontractsrepository.ClassicAuth,
	spaceMember domaincontractsrepository.SpaceMember,
	spaceMemberRole domaincontractsrepository.SpaceMemberRole,
	client domaincontractsrepository.Client,
	clientIpAllowlist domaincontractsrepository.ClientIpAllowlist,
	password domaincontractsutility.Password,
	keygen domaincontractsutility.Keygen,
	validator domaincontractsutility.Validator,
	transactor domaincontractsutility.Transactor,
	logger domaincontractsutility.Logger,
	data *domainmodelsseeder.Data,
) domainusecasesseeder.Seeder {
	return &usecase{
		space:             space,
		permission:        permission,
		role:              role,
		rolePermission:    rolePermission,
		user:              user,
		classicAuth:       classicAuth,
		spaceMember:       spaceMember,
		spaceMemberRole:   spaceMemberRole,
		client:            client,
		clientIpAllowlist: clientIpAllowlist,
		password:          password,
		keygen:            keygen,
		validator:         validator,
		transactor:        transactor,
		logger:            logger,
		data:              data,
	}
}

// Run owns the complete bootstrap commit boundary. Creation logs use only the
// final committed attempt, never records that were rolled back during retry.
func (u *usecase) Run(ctx context.Context) error {
	const tag = tagPath + "/Run"

	if err := u.validate(); err != nil {
		u.logger.Error(ctx, tag, "invalid bootstrap seed definitions", domainmodelslogger.Meta{"err": err})
		return err
	}

	var created changes
	for attempt := 0; attempt < 2; attempt++ {
		created = changes{}
		err := u.transactor.WithTx(context.WithValue(ctx, retryKey{}, attempt == 0), func(txCtx context.Context) error {
			spaceId, err := u.seedSpace(txCtx, &created)
			if err != nil {
				return err
			}

			permissionIds, err := u.seedPermissions(txCtx, spaceId, &created)
			if err != nil {
				return err
			}

			roleIds, err := u.seedRoles(txCtx, spaceId, &created)
			if err != nil {
				return err
			}

			if err := u.seedRolePermissions(txCtx, spaceId, roleIds, permissionIds, &created); err != nil {
				return err
			}

			if err := u.seedUsers(txCtx, spaceId, roleIds, &created); err != nil {
				return err
			}

			return u.seedClients(txCtx, spaceId, &created)
		})
		if err == nil {
			break
		}

		if attempt == 0 && seedConflict(err) {
			u.logger.Warn(ctx, tag, "retrying bootstrap transaction after concurrent creation", domainmodelslogger.Meta{"err": err, "next_attempt": 2})
			continue
		}

		u.logger.Error(ctx, tag, "failed to commit bootstrap seed transaction", domainmodelslogger.Meta{"err": err})
		return err
	}

	for _, item := range created.spaces {
		u.logger.Info(ctx, tag, "created control space", domainmodelslogger.Meta{
			"id":         item.id,
			"space_id":   created.spaceId,
			"identifier": item.identifier,
		})
	}

	for _, item := range created.permissions {
		u.logger.Info(ctx, tag, "created seed permission", domainmodelslogger.Meta{
			"id":         item.id,
			"space_id":   created.spaceId,
			"identifier": item.identifier,
		})
	}

	for _, item := range created.roles {
		u.logger.Info(ctx, tag, "created seed role", domainmodelslogger.Meta{
			"id":         item.id,
			"space_id":   created.spaceId,
			"identifier": item.identifier,
		})
	}

	for _, item := range created.rolePermissions {
		u.logger.Info(ctx, tag, "created seed role-permission assignment", domainmodelslogger.Meta{
			"id":         item.id,
			"space_id":   created.spaceId,
			"identifier": item.identifier,
		})
	}

	for _, item := range created.users {
		u.logger.Info(ctx, tag, "created seed user profile", domainmodelslogger.Meta{
			"id":         item.id,
			"space_id":   created.spaceId,
			"identifier": item.identifier,
		})
	}

	for _, item := range created.classicAuths {
		u.logger.Info(ctx, tag, "created seed classic credentials", domainmodelslogger.Meta{
			"id":         item.id,
			"space_id":   created.spaceId,
			"identifier": item.identifier,
		})
	}

	for _, item := range created.members {
		u.logger.Info(ctx, tag, "created seed membership", domainmodelslogger.Meta{
			"id":         item.id,
			"space_id":   created.spaceId,
			"identifier": item.identifier,
		})
	}

	for _, item := range created.memberRoles {
		u.logger.Info(ctx, tag, "created seed member-role assignment", domainmodelslogger.Meta{
			"id":         item.id,
			"space_id":   created.spaceId,
			"identifier": item.identifier,
		})
	}

	for _, item := range created.clients {
		u.logger.Info(ctx, tag, "created seed client", domainmodelslogger.Meta{
			"id":         item.id,
			"space_id":   created.spaceId,
			"identifier": item.identifier,
		})
	}

	for _, item := range created.ipAllowlist {
		u.logger.Info(ctx, tag, "created seed client IP allowlist", domainmodelslogger.Meta{
			"id":         item.id,
			"space_id":   created.spaceId,
			"identifier": item.identifier,
		})
	}

	return nil
}

func seedConflict(err error) bool {
	return errors.Is(err, domainmodelserror.SpaceSlugConflict) ||
		errors.Is(err, domainmodelserror.PermissionSlugConflict) ||
		errors.Is(err, domainmodelserror.RoleSlugConflict) ||
		errors.Is(err, domainmodelserror.RoleDefaultConflict) ||
		errors.Is(err, domainmodelserror.RolePermissionConflict) ||
		errors.Is(err, domainmodelserror.ClassicAuthUsernameConflict) ||
		errors.Is(err, domainmodelserror.SpaceMemberConflict) ||
		errors.Is(err, domainmodelserror.SpaceMemberRoleConflict) ||
		errors.Is(err, domainmodelserror.ClientSlugConflict) ||
		errors.Is(err, domainmodelserror.ClientIpAllowlistConflict)
}

type retryKey struct{}

func canRetrySeed(ctx context.Context, err error) bool {
	return ctx.Value(retryKey{}) == true && seedConflict(err)
}
