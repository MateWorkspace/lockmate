package domaincontractsrepository

import (
	"context"
	"time"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

// SigningKey persists asymmetric token-signing keys. Private material is stored
// sealed by a caller-supplied utility; adapters never perform cryptography.
type SigningKey interface {
	Create(ctx context.Context, value domainmodelsentity.SigningKey) (id int64, err error)
	ReadActive(ctx context.Context) (*domainmodelsentity.SigningKey, error)
	ReadByKid(ctx context.Context, kid string) (*domainmodelsentity.SigningKey, error)
	// ListVerifiable returns nondeleted active and retiring keys in creation order.
	ListVerifiable(ctx context.Context) ([]domainmodelsentity.SigningKey, error)
	// ListStale returns nondeleted keys whose fingerprint differs from the given one.
	ListStale(ctx context.Context, fingerprint string) ([]domainmodelsentity.SigningKey, error)
	// DemoteActive marks the current active key retiring and returns the demoted ID.
	DemoteActive(ctx context.Context, retiringAt time.Time, by *int64) (int64, error)
	UpdateSealed(ctx context.Context, id int64, fingerprint string, cipher []byte, nonce []byte) error
	// DeleteRetiringBefore soft-deletes retiring keys created before the cutoff.
	DeleteRetiringBefore(ctx context.Context, cutoff time.Time, by *int64) (int64, error)
}
