package infrastructureutilitytransactor

import (
	"context"
	"errors"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"github.com/MateWorkspace/mate-golang/pgxdt"
)

type pgxdtImpl struct {
	dt pgxdt.Transactor
}

// NewPgxdtImpl expects the strict pgxdt.NewTransactor implementation from the
// same pool as the repositories. pgxdt.NewPgxdt retains repository reuse.
func NewPgxdtImpl(dt pgxdt.Transactor) domaincontractsutility.Transactor {
	return &pgxdtImpl{dt: dt}
}

func (t *pgxdtImpl) WithTx(ctx context.Context, fn func(context.Context) error) error {
	var callbackErr error
	err := t.dt.WithTx(ctx, func(txCtx context.Context) error {
		callbackErr = fn(txCtx)
		return callbackErr
	})
	if callbackErr != nil {
		return callbackErr
	}
	if err == nil {
		return nil
	}
	kind := domainmodelserror.Failure
	switch {
	case errors.Is(err, pgxdt.ErrTransactionAlreadyActive), errors.Is(err, context.Canceled):
		kind = domainmodelserror.BadState
	case errors.Is(err, context.DeadlineExceeded):
		kind = domainmodelserror.Timeout
	}
	return domainmodelserror.NewError("transaction failed", kind, err)
}
