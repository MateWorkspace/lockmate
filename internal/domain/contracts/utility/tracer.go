package domaincontractsutility

import (
	"context"

	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
)

type Tracer interface {
	Inject(ctx context.Context, trace domainmodelsapp.Trace) (newCtx context.Context)
	Extract(ctx context.Context) (trace *domainmodelsapp.Trace, err error)
}
