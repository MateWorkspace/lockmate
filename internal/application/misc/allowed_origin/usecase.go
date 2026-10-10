package applicationmiscallowedorigin

import (
	"context"
	"errors"

	applicationshared "github.com/MateWorkspace/lockmate/internal/application/shared"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
)

const tagPath = "misc/allowed_origin"

type usecase struct {
	clientOrigin domaincontractsrepository.ClientOrigin
	validator    domaincontractsutility.Validator
	logger       domaincontractsutility.Logger
}

func NewUsecaseImpl(
	clientOrigin domaincontractsrepository.ClientOrigin,
	validator domaincontractsutility.Validator,
	logger domaincontractsutility.Logger,
) domainusecasesmisc.AllowedOrigin {
	return &usecase{clientOrigin: clientOrigin, validator: validator, logger: logger}
}

func (u *usecase) IsAllowedOrigin(ctx context.Context, origin string) (bool, error) {
	const tag = tagPath + "/IsAllowedOrigin"

	if err := u.validator.ClientOrigin(origin); err != nil {
		if errors.Is(err, domainmodelserror.ClientOriginInvalid) || errors.Is(err, domainmodelserror.ClientOriginTooLong) {
			return false, nil
		}

		u.logger.Error(ctx, tag, "failed to validate origin for global allowlist check", domainmodelslogger.Meta{
			"err": err,
		})

		return false, err
	}

	allowed, err := u.clientOrigin.IsAllowedOrigin(ctx, origin)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to check global origin allowlist", domainmodelslogger.Meta{
			"err": err,
		})

		return false, err
	}

	return allowed, nil
}

func (u *usecase) IsAllowedOriginForClient(ctx context.Context, spaceId, clientId int64, origin string) (bool, error) {
	const tag = tagPath + "/IsAllowedOriginForClient"

	if err := applicationshared.ValidateIds(spaceId, clientId); err != nil {
		u.logger.Error(ctx, tag, "invalid identifiers for client origin allowlist check", domainmodelslogger.Meta{
			"space_id":  spaceId,
			"client_id": clientId,
			"err":       err,
		})

		return false, err
	}

	if err := u.validator.ClientOrigin(origin); err != nil {
		if errors.Is(err, domainmodelserror.ClientOriginInvalid) || errors.Is(err, domainmodelserror.ClientOriginTooLong) {
			return false, nil
		}

		u.logger.Error(ctx, tag, "failed to validate origin for client allowlist check", domainmodelslogger.Meta{
			"space_id":  spaceId,
			"client_id": clientId,
			"err":       err,
		})

		return false, err
	}

	allowed, err := u.clientOrigin.IsAllowedOriginForClient(ctx, spaceId, clientId, origin)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to check client origin allowlist", domainmodelslogger.Meta{
			"space_id":  spaceId,
			"client_id": clientId,
			"err":       err,
		})

		return false, err
	}

	return allowed, nil
}
