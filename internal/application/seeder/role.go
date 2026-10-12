package applicationseeder

import (
	"context"
	"errors"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
)

func (u *usecase) seedRoles(ctx context.Context, spaceId int64, created *changes) (map[string]int64, error) {
	const tag = tagPath + "/seedRoles"

	currentDefault, err := u.role.ReadDefault(ctx, spaceId)
	if err != nil && !errors.Is(err, domainmodelserror.DefaultRoleNotFound) {
		u.logger.Error(ctx, tag, "failed to read current seed-space default role", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	hasDefault := currentDefault != nil
	ids := make(map[string]int64, len(u.data.Roles))
	for _, value := range u.data.Roles {
		existing, err := u.role.ReadBySlug(ctx, spaceId, value.Slug)
		if err == nil {
			if existing == nil {
				err := domainmodelserror.NewError("seed role read returned no record", domainmodelserror.Failure, nil)
				u.logger.Error(ctx, tag, "seed role read returned no record", domainmodelslogger.Meta{"err": err})

				return nil, err
			}

			ids[value.Slug] = existing.Id
			continue
		}

		if !errors.Is(err, domainmodelserror.RoleNotFound) {
			u.logger.Error(ctx, tag, "failed to read seed role", domainmodelslogger.Meta{"err": err})

			return nil, err
		}

		isDefault := value.IsDefault && !hasDefault
		id, err := u.role.Create(ctx, spaceId, domaincontractsrepository.CreateRoleArgs{
			Slug:        value.Slug,
			Name:        value.Name,
			Description: &value.Description,
			IsDefault:   &isDefault,
		})
		if err != nil {
			if canRetrySeed(ctx, err) {
				u.logger.Warn(ctx, tag, "failed to create seed role", domainmodelslogger.Meta{"err": err})
			} else {
				u.logger.Error(ctx, tag, "failed to create seed role", domainmodelslogger.Meta{"err": err})
			}

			return nil, err
		}

		hasDefault = hasDefault || isDefault
		ids[value.Slug] = id
		created.roles = append(created.roles, record{id: id, identifier: value.Slug})
	}

	if !hasDefault {
		err := domainmodelserror.NewError("control space has no default role; configure one before seeding", domainmodelserror.BadState, nil)
		u.logger.Error(ctx, tag, "control space is missing its default role", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	return ids, nil
}
