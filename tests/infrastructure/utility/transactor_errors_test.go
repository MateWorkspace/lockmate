package utility_test

import (
	"context"
	"errors"
	"testing"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructureutilitytransactor "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/transactor"
	"github.com/MateWorkspace/mate-golang/pgxdt"
)

type failingTransactor struct{ err error }

func (s failingTransactor) WithTx(context.Context, func(context.Context) error) error { return s.err }
func TestTransactorLifecycleErrorClassification(t *testing.T) {
	for _, test := range []struct {
		name         string
		source, kind error
	}{
		{"nested", pgxdt.ErrTransactionAlreadyActive, domainmodelserror.BadState},
		{"deadline", context.DeadlineExceeded, domainmodelserror.Timeout},
		{"canceled", context.Canceled, domainmodelserror.BadState},
		{"driver", errors.New("driver failure"), domainmodelserror.Failure},
	} {
		t.Run(test.name, func(t *testing.T) {
			adapter := infrastructureutilitytransactor.NewPgxdtImpl(failingTransactor{test.source})
			err := adapter.WithTx(context.Background(), func(context.Context) error { t.Fatal("callback reached"); return nil })
			if !errors.Is(err, test.kind) || !errors.Is(err, test.source) {
				t.Fatalf("error classification/source lost: %v", err)
			}
		})
	}
}
