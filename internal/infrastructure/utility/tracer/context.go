package infrastructureutilitytracer

import (
	"context"
	"github.com/google/uuid"

	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
)

type traceKey struct{}

type contextImpl struct{}

func NewContextImpl() domaincontractsutility.Tracer {
	return &contextImpl{}
}

func (t *contextImpl) GenerateEventId() (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, domainmodelserror.NewError("failed to generate trace event id", domainmodelserror.TraceEventIdFailed, err)
	}
	return id, nil
}

func (t *contextImpl) Inject(ctx context.Context, trace domainmodelsapp.Trace) context.Context {
	return context.WithValue(ctx, traceKey{}, trace)
}

func (t *contextImpl) Extract(ctx context.Context) (*domainmodelsapp.Trace, error) {
	trace, ok := ctx.Value(traceKey{}).(domainmodelsapp.Trace)
	if !ok {
		return nil, domainmodelserror.NewError("trace is not present in context", domainmodelserror.TraceNotFound, nil)
	}

	return &trace, nil
}
