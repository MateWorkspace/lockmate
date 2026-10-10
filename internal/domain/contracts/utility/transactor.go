package domaincontractsutility

import "context"

type Transactor interface {
	// WithTx owns a top-level commit boundary and rejects an attached transaction.
	// Repository helpers may reuse its callback context internally.
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
