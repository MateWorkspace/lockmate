package domaincontractsrepository

import (
	"context"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

// GoogleOAuthLaunch persists internal, expiring records. Consume is atomic and respects
// the supplied transaction context. PurgeExpired deletes at most limit rows.
type GoogleOAuthLaunch interface {
	Create(ctx context.Context, value domainmodelsentity.GoogleOAuthLaunch) error
	Consume(ctx context.Context, ticketHash string) (*domainmodelsentity.GoogleOAuthLaunch, error)
	PurgeExpired(ctx context.Context, limit int64) (int64, error)
}
