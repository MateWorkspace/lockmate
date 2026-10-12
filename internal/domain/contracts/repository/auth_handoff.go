package domaincontractsrepository

import (
	"context"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type AuthHandoff interface {
	Create(ctx context.Context, value domainmodelsentity.AuthHandoff) error
	Consume(ctx context.Context, codeHash string, clientId int64) (*domainmodelsentity.AuthHandoff, error)
	PurgeExpired(ctx context.Context, limit int64) (int64, error)
}
