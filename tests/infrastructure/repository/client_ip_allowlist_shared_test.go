package repository_test

import (
	"context"
	"errors"
	"math"
	"net/netip"
	"strings"
	"testing"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructurerepositoryclientipallowlist "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/client_ip_allowlist"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestClientIpAllowlistBoundInputsAndContext(t *testing.T) {
	question, dollar := builders()
	ctx := context.WithValue(context.Background(), struct{}{}, "caller")
	ip := netip.MustParseAddr("203.0.113.9")
	dt := &recordingExecutor{row: scanRow(func(dest ...any) error {
		*dest[0].(*int) = 1
		return nil
	})}
	repo := infrastructurerepositoryclientipallowlist.NewPostgresImpl(dt, question, dollar)
	allowed, err := repo.IsAllowedForClient(ctx, 7, 9, ip)
	require(t, err)
	if !allowed || len(dt.calls) != 1 || dt.calls[0].ctx != ctx {
		t.Fatal("match or context handling failed")
	}
	call := dt.calls[0]
	if strings.Contains(call.sql, ip.String()) || strings.Contains(call.sql, "?") || !strings.Contains(call.sql, "FROM client_ip_allowlist") {
		t.Fatalf("unsafe query or incorrect placeholder composition: %s", call.sql)
	}
	bound := false
	for _, arg := range call.args {
		if arg == ip.String() {
			bound = true
		}
	}
	if !bound {
		t.Fatal("IP not bound")
	}
}

func TestClientIpAllowlistNoRowsIsOrdinaryNonmatch(t *testing.T) {
	question, dollar := builders()
	dt := &recordingExecutor{row: scanRow(func(...any) error { return pgx.ErrNoRows })}
	repo := infrastructurerepositoryclientipallowlist.NewPostgresImpl(dt, question, dollar)
	allowed, err := repo.IsAllowedForClient(context.Background(), 7, 9, netip.MustParseAddr("203.0.113.9"))
	if err != nil || allowed {
		t.Fatalf("no rows was not an ordinary nonmatch: %v %v", allowed, err)
	}
}

func TestClientIpAllowlistDriverErrors(t *testing.T) {
	question, dollar := builders()
	for _, source := range []error{errors.New("driver failed"), context.Canceled, context.DeadlineExceeded} {
		dt := &recordingExecutor{row: scanRow(func(...any) error { return source })}
		repo := infrastructurerepositoryclientipallowlist.NewPostgresImpl(dt, question, dollar)
		allowed, err := repo.IsAllowedForClient(context.Background(), 1, 2, netip.MustParseAddr("203.0.113.9"))
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

	for _, constraint := range []string{"uq_client_ip_allowlist_client_id_cidr", "prefix_uq_client_ip_allowlist_client_id_cidr"} {
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
		repo := infrastructurerepositoryclientipallowlist.NewPostgresImpl(dt, question, dollar)
		id, err := repo.Create(context.Background(), 1, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: 2, Cidr: "203.0.113.0/24"})
		kind := domainmodelserror.Conflict
		if constraint == "uq_client_ip_allowlist_client_id_cidr" {
			kind = domainmodelserror.ClientIpAllowlistConflict
		}
		if id != 0 || !errors.Is(err, kind) || !errors.Is(err, source) {
			t.Fatalf("incorrect conflict mapping or source lost: %v", err)
		}
		if !strings.Contains(dt.calls[0].sql, "FROM spaces") || !strings.Contains(dt.calls[1].sql, "FROM clients") {
			t.Fatal("parent lock ordering changed")
		}
	}
}

func TestClientIpAllowlistListsCloseRows(t *testing.T) {
	question, dollar := builders()
	failure := errors.New("row stream failed")
	for _, filtered := range []bool{false, true} {
		for _, rowError := range []error{nil, failure} {
			rows := &emptyRows{failure: rowError}
			dt := &recordingExecutor{rows: rows, row: scanRow(func(dest ...any) error {
				*dest[0].(*int64) = 0
				return nil
			})}
			repo := infrastructurerepositoryclientipallowlist.NewPostgresImpl(dt, question, dollar)
			if filtered {
				items, total, err := repo.ReadByFilter(context.Background(), 1, domaincontractsrepository.ClientIpAllowlistFilter{Limit: 10})
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

func TestClientIpAllowlistRejectedArgumentsDoNotExecute(t *testing.T) {
	question, dollar := builders()
	dt := &recordingExecutor{}
	repo := infrastructurerepositoryclientipallowlist.NewPostgresImpl(dt, question, dollar)
	ctx := context.Background()
	for _, fail := range []func() error{
		func() error {
			_, err := repo.Create(ctx, 1, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: 0})
			return err
		},
		func() error {
			_, err := repo.Create(ctx, 1, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: 2, By: ptr(int64(-1))})
			return err
		},
		func() error { _, err := repo.ReadById(ctx, 0, 2); return err },
		func() error { _, err := repo.ReadByClientIdAndCidr(ctx, 1, 0, ""); return err },
		func() error { _, err := repo.ReadByClientId(ctx, 1, 0); return err },
		func() error {
			_, _, err := repo.ReadByFilter(ctx, 1, domaincontractsrepository.ClientIpAllowlistFilter{ClientId: ptr(int64(0))})
			return err
		},
		func() error {
			_, _, err := repo.ReadByFilter(ctx, 1, domaincontractsrepository.ClientIpAllowlistFilter{Page: math.MaxInt64, Limit: 2})
			return err
		},
		func() error {
			return repo.UpdateById(ctx, 1, 2, domaincontractsrepository.UpdateClientIpAllowlistArgs{By: ptr(int64(0))})
		},
		func() error { return repo.DeleteById(ctx, 1, 0, nil) },
		func() error {
			_, err := repo.IsAllowedForClient(ctx, 1, 0, netip.MustParseAddr("203.0.113.9"))
			return err
		},
		func() error { _, err := repo.IsAllowedForClient(ctx, 1, 2, netip.Addr{}); return err },
	} {
		if err := fail(); !errors.Is(err, domainmodelserror.BadArgs) {
			t.Fatalf("expected BadArgs: %v", err)
		}
	}
	if len(dt.calls) != 0 {
		t.Fatal("invalid arguments reached database")
	}
}

func TestClientIpAllowlistExactNotFound(t *testing.T) {
	question, dollar := builders()
	dt := &recordingExecutor{row: scanRow(func(...any) error { return pgx.ErrNoRows })}
	repo := infrastructurerepositoryclientipallowlist.NewPostgresImpl(dt, question, dollar)
	_, err := repo.ReadById(context.Background(), 1, 2)
	if !errors.Is(err, domainmodelserror.ClientIpAllowlistNotFound) || !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("not-found classification or source lost: %v", err)
	}
}
