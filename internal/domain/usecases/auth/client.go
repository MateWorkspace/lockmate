package domainusecasesauth

import (
	"context"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

// Client authenticates consuming applications and resolves their space.
// Unknown, inactive, deleted, or mismatched credentials return Unauthorized;
// repository and utility operational failures preserve their classification.
type Client interface {
	Authenticate(ctx context.Context, request AuthenticateClientRequest) (*domainmodelsentity.ClientIdentity, error)
}

type AuthenticateClientRequest struct {
	Slug   string
	Secret string
}
