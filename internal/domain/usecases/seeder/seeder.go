package domainusecasesseeder

import "context"

// Seeder fills missing control-space bootstrap records in one transaction.
// Existing metadata and credentials are preserved; baseline grants are restored.
type Seeder interface {
	Run(ctx context.Context) error
}
