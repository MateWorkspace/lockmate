package applicationseeder

import (
	"context"
	"errors"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
)

func (u *usecase) seedSpace(ctx context.Context, created *changes) (int64, error) {
	const tag = tagPath + "/seedSpace"

	existing, err := u.space.ReadBySlug(ctx, u.data.Space.Slug)
	if err == nil {
		if existing == nil {
			err := domainmodelserror.NewError("seed space read returned no record", domainmodelserror.Failure, nil)
			u.logger.Error(ctx, tag, "seed space read returned no record", domainmodelslogger.Meta{"err": err})

			return 0, err
		}

		created.spaceId = existing.Id
		return existing.Id, nil
	}

	if !errors.Is(err, domainmodelserror.SpaceNotFound) {
		u.logger.Error(ctx, tag, "failed to read control space", domainmodelslogger.Meta{"err": err})

		return 0, err
	}

	value := u.data.Space
	id, err := u.space.Create(ctx, domaincontractsrepository.CreateSpaceArgs{
		Slug:        value.Slug,
		Name:        value.Name,
		Description: &value.Description,
	})
	if err != nil {
		if canRetrySeed(ctx, err) {
			u.logger.Warn(ctx, tag, "failed to create control space", domainmodelslogger.Meta{"err": err})
		} else {
			u.logger.Error(ctx, tag, "failed to create control space", domainmodelslogger.Meta{"err": err})
		}

		return 0, err
	}

	created.spaceId = id
	created.spaces = append(created.spaces, record{id: id, identifier: value.Slug})
	return id, nil
}
