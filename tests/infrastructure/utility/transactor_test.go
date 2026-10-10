package utility_test

import (
	"context"
	"errors"
	infrastructureutilitytransactor "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/transactor"
	"github.com/MateWorkspace/mate-golang/pgxdt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"strings"
	"testing"
)

type transactionStub struct{ called bool }

func (s *transactionStub) WithTx(ctx context.Context, fn func(context.Context) error) error {
	s.called = true
	return fn(ctx)
}

func TestTransactorPostgres(t *testing.T) {
	url := os.Getenv("LOCKMATE_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("LOCKMATE_TEST_DATABASE_URL is required for isolated PostgreSQL integration")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	schema := "transactor_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := pool.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := pool.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("schema cleanup: %v", err)
		}
	}()
	table := schema + ".items"
	if _, err := pool.Exec(ctx, "CREATE TABLE "+table+" (id BIGINT PRIMARY KEY)"); err != nil {
		t.Fatal(err)
	}
	db := pgxdt.NewPgxdt(pool)
	transactor := infrastructureutilitytransactor.NewPgxdtImpl(pgxdt.NewTransactor(pool))
	if err := transactor.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := db.Exec(txCtx, "INSERT INTO "+table+" VALUES (1)"); err != nil {
			return err
		}
		return transactor.WithTx(txCtx, func(nested context.Context) error {
			var count int
			if err := db.QueryRow(nested, "SELECT count(*) FROM "+table).Scan(&count); err != nil {
				return err
			}
			if count != 1 {
				return errors.New("nested context did not reuse transaction")
			}
			_, err := db.Exec(nested, "INSERT INTO "+table+" VALUES (2)")
			return err
		})
	}); err != nil {
		t.Fatal(err)
	}
	want := errors.New("rollback callback")
	if err := transactor.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := db.Exec(txCtx, "INSERT INTO "+table+" VALUES (3)"); err != nil {
			return err
		}
		return transactor.WithTx(txCtx, func(context.Context) error { return want })
	}); err != want {
		t.Fatalf("callback error changed: %v", err)
	}
	var count int
	if err := db.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&count); err != nil || count != 2 {
		t.Fatalf("commit or rollback failed: count=%d err=%v", count, err)
	}
}
func TestTransactorDelegation(t *testing.T) {
	stub := &transactionStub{}
	adapter := infrastructureutilitytransactor.NewPgxdtImpl(stub)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	want := errors.New("callback failed")
	err := adapter.WithTx(ctx, func(received context.Context) error {
		if received != ctx {
			t.Fatal("context replaced")
		}
		return want
	})
	if !stub.called || err != want {
		t.Fatalf("callback error was changed: %v", err)
	}
}
