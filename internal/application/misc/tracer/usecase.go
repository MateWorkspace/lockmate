package applicationmisctracer

import (
	"context"

	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
	"github.com/google/uuid"
)

const tagPath = "misc/tracer"

type usecase struct {
	tracer domaincontractsutility.Tracer
	logger domaincontractsutility.Logger
}

func NewUsecaseImpl(
	tracer domaincontractsutility.Tracer,
	logger domaincontractsutility.Logger,
) domainusecasesmisc.Tracer {
	return &usecase{tracer: tracer, logger: logger}
}

func (u *usecase) Inject(ctx context.Context, clientSlug string, clientExecId uuid.UUID) (context.Context, error) {
	const tag = tagPath + "/Inject"
	eventId, err := u.tracer.GenerateEventId()
	if err != nil {

		u.logger.Error(ctx, tag, "failed to generate trace event id", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	trace := domainmodelsapp.Trace{EventId: eventId, ClientSlug: clientSlug, ClientExecId: clientExecId}
	return u.tracer.Inject(ctx, trace), nil
}
