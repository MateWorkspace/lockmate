package domainusecasesauth

import (
	"context"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

// Handoff exchanges an opaque single-use code after authenticating its client.
type Handoff interface {
	Exchange(ctx context.Context, client *domainmodelsentity.ClientIdentity, request ExchangeAuthHandoffRequest) (*SessionResult, error)
}

type ExchangeAuthHandoffRequest struct{ Code string }
