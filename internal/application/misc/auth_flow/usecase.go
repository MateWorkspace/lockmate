package applicationmiscauthflow

import (
	"context"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
)

type usecase struct {
	launches domaincontractsrepository.GoogleOAuthLaunch
	attempts domaincontractsrepository.GoogleOAuthAttempt
	handoffs domaincontractsrepository.AuthHandoff
	logger   domaincontractsutility.Logger
}

func NewUsecaseImpl(launches domaincontractsrepository.GoogleOAuthLaunch, attempts domaincontractsrepository.GoogleOAuthAttempt, handoffs domaincontractsrepository.AuthHandoff, logger domaincontractsutility.Logger) domainusecasesmisc.AuthFlow {
	return &usecase{
		launches: launches,
		attempts: attempts,
		handoffs: handoffs,
		logger:   logger,
	}
}

func (u *usecase) PurgeExpired(ctx context.Context) error {
	const tag = "misc/auth_flow/PurgeExpired"
	for {
		count, err := u.launches.PurgeExpired(ctx, 1000)
		if err != nil {
			u.logger.Warn(ctx, tag, "failed to purge expired Google launch tickets", domainmodelslogger.Meta{"err": err})

			return err
		}

		if count < 1000 {
			break
		}
	}

	for {
		count, err := u.attempts.PurgeExpired(ctx, 1000)
		if err != nil {
			u.logger.Warn(ctx, tag, "failed to purge expired Google authorization attempts", domainmodelslogger.Meta{"err": err})

			return err
		}

		if count < 1000 {
			break
		}
	}

	for {
		count, err := u.handoffs.PurgeExpired(ctx, 1000)
		if err != nil {
			u.logger.Warn(ctx, tag, "failed to purge expired authentication handoffs", domainmodelslogger.Meta{"err": err})

			return err
		}

		if count < 1000 {
			break
		}
	}

	return nil
}
