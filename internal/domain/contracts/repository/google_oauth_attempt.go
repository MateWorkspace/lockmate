package domaincontractsrepository

import (
	"context"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

// GoogleOAuthAttempt persists internal, expiring records. Consume is atomic and respects
// the supplied transaction context. PurgeExpired deletes at most limit rows.
type GoogleOAuthAttempt interface {
	Create(ctx context.Context, value domainmodelsentity.GoogleOAuthAttempt) error
	Consume(ctx context.Context, stateHash string, bindingHash string) (*domainmodelsentity.GoogleOAuthAttempt, error)
	PurgeExpired(ctx context.Context, limit int64) (int64, error)
}
