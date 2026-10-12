package applicationseeder

import (
	"context"
	"errors"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
)

func (u *usecase) seedPermissions(ctx context.Context, spaceId int64, created *changes) (map[string]int64, error) {
	const tag = tagPath + "/seedPermissions"

	ids := make(map[string]int64, len(u.data.Permissions))
	for _, value := range u.data.Permissions {
		existing, err := u.permission.ReadBySlug(ctx, spaceId, value.Slug)
		if err == nil {
			if existing == nil {
				err := domainmodelserror.NewError("seed permission read returned no record", domainmodelserror.Failure, nil)
				u.logger.Error(ctx, tag, "seed permission read returned no record", domainmodelslogger.Meta{"err": err})

				return nil, err
			}

			ids[value.Slug] = existing.Id
			continue
		}

		if !errors.Is(err, domainmodelserror.PermissionNotFound) {
			u.logger.Error(ctx, tag, "failed to read seed permission", domainmodelslogger.Meta{"err": err})

			return nil, err
		}

		id, err := u.permission.Create(ctx, spaceId, domaincontractsrepository.CreatePermissionArgs{
			Slug:        value.Slug,
			Name:        value.Name,
			Description: &value.Description,
		})
		if err != nil {
			if canRetrySeed(ctx, err) {
				u.logger.Warn(ctx, tag, "failed to create seed permission", domainmodelslogger.Meta{"err": err})
			} else {
				u.logger.Error(ctx, tag, "failed to create seed permission", domainmodelslogger.Meta{"err": err})
			}

			return nil, err
		}

		ids[value.Slug] = id
		created.permissions = append(created.permissions, record{id: id, identifier: value.Slug})
	}

	return ids, nil
}
