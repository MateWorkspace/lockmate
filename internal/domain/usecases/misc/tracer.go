package domainusecasesmisc

import (
	"context"

	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
)

// Tracer exposes trace injection to presentation through an application usecase.
// Application delegates to the tracer utility contract, preserving the parent
// context's cancellation and deadlines. Trace values are diagnostic metadata,
// never authenticated identity.
type Tracer interface {
	Inject(ctx context.Context, trace domainmodelsapp.Trace) (newCtx context.Context)
}
