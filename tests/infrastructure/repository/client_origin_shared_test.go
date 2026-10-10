package repository_test

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructurerepositoryclientorigin "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/client_origin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestClientOriginBoundInputsAndContext(t *testing.T) {
	question, dollar := builders()
	ctx := context.WithValue(context.Background(), struct{}{}, "caller")
	origin := "https://example.com' OR 1=1 --"
	for _, scoped := range []bool{false, true} {
		t.Run(map[bool]string{false: "global", true: "scoped"}[scoped], func(t *testing.T) {
			dt := &recordingExecutor{row: scanRow(func(dest ...any) error {
				*dest[0].(*bool) = false
				return nil
			})}
			repo := infrastructurerepositoryclientorigin.NewPostgresImpl(dt, question, dollar)
			var allowed bool
			var err error
			if scoped {
				allowed, err = repo.IsAllowedOriginForClient(ctx, 7, 9, origin)
			} else {
				allowed, err = repo.IsAllowedOrigin(ctx, origin)
			}
			require(t, err)
			if allowed || len(dt.calls) != 1 || dt.calls[0].ctx != ctx {
				t.Fatal("nonmatch or context handling failed")
			}
			call := dt.calls[0]
			if strings.Contains(call.sql, origin) || strings.Contains(call.sql, "?") || !strings.Contains(call.sql, "EXISTS (SELECT 1") {
				t.Fatalf("unsafe query or incorrect placeholder composition: %s", call.sql)
			}
			bound := false
			for _, arg := range call.args {
				if arg == origin {
					bound = true
				}
			}
			if !bound {
				t.Fatal("origin not bound")
			}
		})
	}
}

func TestClientOriginDriverErrors(t *testing.T) {
	question, dollar := builders()
	for _, source := range []error{errors.New("driver failed"), context.Canceled, context.DeadlineExceeded} {
		dt := &recordingExecutor{row: scanRow(func(...any) error { return source })}
		repo := infrastructurerepositoryclientorigin.NewPostgresImpl(dt, question, dollar)
		for _, check := range []func() (bool, error){
			func() (bool, error) { return repo.IsAllowedOrigin(context.Background(), "https://example.com") },
			func() (bool, error) {
				return repo.IsAllowedOriginForClient(context.Background(), 1, 2, "https://example.com")
			},
		} {
			allowed, err := check()
			if allowed || !errors.Is(err, source) {
				t.Fatalf("failure treated as an ordinary nonmatch or source lost: %v", err)
			}
			kind := domainmodelserror.Failure
			if source == context.Canceled {
				kind = domainmodelserror.BadState
			} else if source == context.DeadlineExceeded {
				kind = domainmodelserror.Timeout
			}
			if !errors.Is(err, kind) {
				t.Fatalf("wrong error classification: %v", err)
			}
		}
	}

	for _, constraint := range []string{"uq_client_origins_client_id_origin", "prefix_uq_client_origins_client_id_origin"} {
		source := &pgconn.PgError{Code: "23505", ConstraintName: constraint}
		calls := 0
		dt := &recordingExecutor{row: scanRow(func(dest ...any) error {
			calls++
			if calls == 3 {
				return source
			}
			*dest[0].(*int64) = 1
			return nil
		})}
		repo := infrastructurerepositoryclientorigin.NewPostgresImpl(dt, question, dollar)
		id, err := repo.Create(context.Background(), 1, domaincontractsrepository.CreateClientOriginArgs{ClientId: 2, Origin: "https://example.com"})
		kind := domainmodelserror.Conflict
		if constraint == "uq_client_origins_client_id_origin" {
			kind = domainmodelserror.ClientOriginConflict
		}
		if id != 0 || !errors.Is(err, kind) || !errors.Is(err, source) {
			t.Fatalf("incorrect conflict mapping or source lost: %v", err)
		}
		if !strings.Contains(dt.calls[0].sql, "FROM spaces") || !strings.Contains(dt.calls[1].sql, "FROM clients") {
			t.Fatal("parent lock ordering changed")
		}
	}
}

func TestClientOriginListsCloseRows(t *testing.T) {
	question, dollar := builders()
	failure := errors.New("row stream failed")
	for _, filtered := range []bool{false, true} {
		for _, rowError := range []error{nil, failure} {
			rows := &emptyRows{failure: rowError}
			dt := &recordingExecutor{rows: rows, row: scanRow(func(dest ...any) error {
				*dest[0].(*int64) = 0
				return nil
			})}
			repo := infrastructurerepositoryclientorigin.NewPostgresImpl(dt, question, dollar)
			if filtered {
				items, total, err := repo.ReadByFilter(context.Background(), 1, domaincontractsrepository.ClientOriginFilter{Limit: 10})
				if rowError == nil && (err != nil || items == nil || total != 0) || rowError != nil && !errors.Is(err, rowError) {
					t.Fatalf("incorrect filter result: %v", err)
				}
			} else {
				items, err := repo.ReadByClientId(context.Background(), 1, 2)
				if rowError == nil && (err != nil || items == nil) || rowError != nil && !errors.Is(err, rowError) {
					t.Fatalf("incorrect list result: %v", err)
				}
			}
			if !rows.closed {
				t.Fatal("rows were not closed")
			}
		}
	}
}

func TestClientOriginRejectedArgumentsDoNotExecute(t *testing.T) {
	question, dollar := builders()
	dt := &recordingExecutor{}
	repo := infrastructurerepositoryclientorigin.NewPostgresImpl(dt, question, dollar)
	ctx := context.Background()
	for _, fail := range []func() error{
		func() error {
			_, err := repo.Create(ctx, 1, domaincontractsrepository.CreateClientOriginArgs{ClientId: 0})
			return err
		},
		func() error {
			_, err := repo.Create(ctx, 1, domaincontractsrepository.CreateClientOriginArgs{ClientId: 2, By: ptr(int64(-1))})
			return err
		},
		func() error { _, err := repo.ReadById(ctx, 0, 2); return err },
		func() error { _, err := repo.ReadByClientIdAndOrigin(ctx, 1, 0, ""); return err },
		func() error { _, err := repo.ReadByClientId(ctx, 1, 0); return err },
		func() error {
			_, _, err := repo.ReadByFilter(ctx, 1, domaincontractsrepository.ClientOriginFilter{ClientId: ptr(int64(0))})
			return err
		},
		func() error {
			_, _, err := repo.ReadByFilter(ctx, 1, domaincontractsrepository.ClientOriginFilter{Page: math.MaxInt64, Limit: 2})
			return err
		},
		func() error {
			return repo.UpdateById(ctx, 1, 2, domaincontractsrepository.UpdateClientOriginArgs{By: ptr(int64(0))})
		},
		func() error { return repo.DeleteById(ctx, 1, 0, nil) },
		func() error { _, err := repo.IsAllowedOriginForClient(ctx, 1, 0, ""); return err },
	} {
		if err := fail(); !errors.Is(err, domainmodelserror.BadArgs) {
			t.Fatalf("expected BadArgs: %v", err)
		}
	}
	if len(dt.calls) != 0 {
		t.Fatal("invalid arguments reached database")
	}
}

func TestClientOriginExactNotFound(t *testing.T) {
	question, dollar := builders()
	dt := &recordingExecutor{row: scanRow(func(...any) error { return pgx.ErrNoRows })}
	repo := infrastructurerepositoryclientorigin.NewPostgresImpl(dt, question, dollar)
	_, err := repo.ReadById(context.Background(), 1, 2)
	if !errors.Is(err, domainmodelserror.ClientOriginNotFound) || !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("not-found classification or source lost: %v", err)
	}
}
