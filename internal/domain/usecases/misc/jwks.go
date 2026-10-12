package domainusecasesmisc

import (
	"context"
	"encoding/json"
)

// Jwks exposes the public JSON Web Keys verifiable for the authenticated client.
type Jwks interface {
	PublicKeys(ctx context.Context) ([]json.RawMessage, error)
}
