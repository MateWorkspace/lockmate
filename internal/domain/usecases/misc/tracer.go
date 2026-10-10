package domainusecasesmisc

import (
	"context"

	"github.com/google/uuid"
)

// Tracer generates an event ID internally and injects external diagnostic metadata.
// The derived context preserves parent values, cancellation, and deadlines.
type Tracer interface {
	Inject(ctx context.Context, clientSlug string, clientExecId uuid.UUID) (newCtx context.Context, err error)
}
