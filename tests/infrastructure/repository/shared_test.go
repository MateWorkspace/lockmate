package repository_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructurerepositoryclient "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/client"
	infrastructurerepositoryshared "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/shared"
	infrastructurerepositoryspace "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/space"
	infrastructurerepositoryspacememberrole "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/space_member_role"
	infrastructurerepositoryuser "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"math"
	"strings"
	"testing"
)

type queryCall struct {
	ctx  context.Context
	sql  string
	args []any
}
type recordingExecutor struct {
	calls []queryCall
	row   pgx.Row
	rows  pgx.Rows
}

func (r *recordingExecutor) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	r.calls = append(r.calls, queryCall{ctx, sql, args})
	return pgconn.NewCommandTag("UPDATE 1"), nil
}
func (r *recordingExecutor) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	r.calls = append(r.calls, queryCall{ctx, sql, args})
	return r.row
}
func (r *recordingExecutor) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	r.calls = append(r.calls, queryCall{ctx, sql, args})
	return r.rows, nil
}
func (r *recordingExecutor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type scanRow func(...any) error

func (r scanRow) Scan(dest ...any) error { return r(dest...) }

type emptyRows struct {
	pgx.Rows
	closed  bool
	failure error
}

func (r *emptyRows) Next() bool { return false }
func (r *emptyRows) Err() error { return r.failure }
func (r *emptyRows) Close()     { r.closed = true }

func builders() (*squirrel.StatementBuilderType, *squirrel.StatementBuilderType) {
	question := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)
	dollar := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	return &question, &dollar
}

func TestBoundSearchAndContext(t *testing.T) {
	input := "x' OR 1=1 --"
	dt := &recordingExecutor{row: scanRow(func(...any) error { return pgx.ErrNoRows })}
	question, dollar := builders()
	repo := infrastructurerepositoryspace.NewPostgresImpl(dt, question, dollar)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := repo.ReadBySlug(ctx, input)
	if !errors.Is(err, domainmodelserror.SpaceNotFound) || !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("not-found cause lost: %v", err)
	}
	if len(dt.calls) != 1 || strings.Contains(dt.calls[0].sql, input) || !strings.Contains(dt.calls[0].sql, "$1") || dt.calls[0].args[0] != input || dt.calls[0].ctx != ctx {
		t.Fatal("input interpolated or context changed")
	}
}

func TestSafeProjectionAndRowsClosed(t *testing.T) {
	failure := errors.New("row stream failed")
	for _, rowsError := range []error{nil, failure} {
		rows := &emptyRows{failure: rowsError}
		dt := &recordingExecutor{rows: rows, row: scanRow(func(dest ...any) error { *dest[0].(*int64) = 0; return nil })}
		question, dollar := builders()
		repo := infrastructurerepositoryclient.NewPostgresImpl(dt, question, dollar)
		items, total, err := repo.ReadByFilter(context.Background(), 1, domaincontractsrepository.ClientFilter{Page: 1, Limit: 10, Search: ptr("web")})
		if rowsError == nil {
			if err != nil || items == nil || total != 0 {
				t.Fatalf("unexpected empty page: %v", err)
			}
		} else if !errors.Is(err, failure) || !errors.Is(err, domainmodelserror.Failure) {
			t.Fatalf("row error lost: %v", err)
		}
		if !rows.closed || len(dt.calls) != 2 || strings.Contains(dt.calls[1].sql, "secret_key_hash") || !strings.Contains(dt.calls[1].sql, "e.space_id") {
			t.Fatal("unsafe projection, missing scope, or rows not closed")
		}
	}
}

func TestRejectedArgumentsDoNotExecute(t *testing.T) {
	dt := &recordingExecutor{}
	question, dollar := builders()
	space := infrastructurerepositoryspace.NewPostgresImpl(dt, question, dollar)
	user := infrastructurerepositoryuser.NewPostgresImpl(dt, question, dollar)
	assignment := infrastructurerepositoryspacememberrole.NewPostgresImpl(dt, question, dollar)
	ctx := context.Background()
	failures := []func() error{
		func() error { _, err := space.ReadById(ctx, 0); return err },
		func() error {
			_, _, err := space.ReadByFilter(ctx, domaincontractsrepository.SpaceFilter{Page: math.MaxInt64, Limit: 2})
			return err
		},
		func() error {
			return user.UpdateById(ctx, 1, domaincontractsrepository.UpdateUserArgs{Preferences: ptr(json.RawMessage(`invalid`))})
		},
		func() error {
			return user.UpdateById(ctx, 1, domaincontractsrepository.UpdateUserArgs{AvatarPath: ptr("path"), AvatarPathSetNull: true})
		},
		func() error {
			return assignment.DeleteBySpaceMemberIdOrRoleId(ctx, 1, domaincontractsrepository.DeleteSpaceMemberRoleArgs{})
		},
	}
	for _, fail := range failures {
		if err := fail(); !errors.Is(err, domainmodelserror.BadArgs) {
			t.Fatalf("expected BadArgs: %v", err)
		}
	}
	if len(dt.calls) != 0 {
		t.Fatal("invalid arguments reached database")
	}
}

func TestPgxErrorMapping(t *testing.T) {
	sentinel := errors.New("driver failure")
	for _, test := range []struct {
		name              string
		source, errorType error
	}{
		{"deadline", context.DeadlineExceeded, domainmodelserror.Timeout},
		{"canceled", context.Canceled, domainmodelserror.BadState},
		{"missing", pgx.ErrNoRows, domainmodelserror.NotFound},
		{"foreign key", &pgconn.PgError{Code: "23503"}, domainmodelserror.Conflict},
		{"bad value", &pgconn.PgError{Code: "22P02"}, domainmodelserror.BadArgs},
		{"specific conflict", &pgconn.PgError{Code: "23505", ConstraintName: "uq_clients_slug"}, domainmodelserror.ClientSlugConflict},
		{"exact name required", &pgconn.PgError{Code: "23505", ConstraintName: "prefix_uq_clients_slug"}, domainmodelserror.Conflict},
		{"driver", sentinel, domainmodelserror.Failure},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := infrastructurerepositoryshared.MapPgxError("failed", test.source, infrastructurerepositoryshared.ConflictMatch{Constraint: "uq_clients_slug", Type: domainmodelserror.ClientSlugConflict})
			if !errors.Is(err, test.errorType) || !errors.Is(err, test.source) {
				t.Fatalf("wrong classification or lost source: %v", err)
			}
			if retained := infrastructurerepositoryshared.MapPgxError("outer", err); retained != err {
				t.Fatal("domain error wrapped again")
			}
		})
	}
}
