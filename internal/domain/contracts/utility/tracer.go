package domaincontractsutility

import (
	"context"
	"github.com/google/uuid"

	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
)

type Tracer interface {
	GenerateEventId() (uuid.UUID, error)
	Inject(ctx context.Context, trace domainmodelsapp.Trace) (newCtx context.Context)
	Extract(ctx context.Context) (trace *domainmodelsapp.Trace, err error)
}
