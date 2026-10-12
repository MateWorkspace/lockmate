package presentationcronjob

import (
	"context"

	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
)

// SigningKey reconciles the persisted token-signing keys, creating or rotating
// the active key and retiring keys past retention.
func SigningKey(usecase domainusecasesmisc.SigningKey) func(ctx context.Context) error {
	return usecase.Reconcile
}
