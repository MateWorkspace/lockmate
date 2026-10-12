package presentationcronjob

import (
	"context"

	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
)

// AuthFlow purges expired Google launch, attempt, and authentication handoff records.
func AuthFlow(usecase domainusecasesmisc.AuthFlow) func(ctx context.Context) error {
	return usecase.PurgeExpired
}
