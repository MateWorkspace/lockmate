package management_test

import (
	"context"
	"errors"
	"math"
	"testing"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
)

const clientOriginValue = "https://app.example.com"

func originFixture(t *testing.T) *testshared.Fixture {
	t.Helper()
	f := testshared.NewFixture()
	f.ClientOrigin.ReadByIdFn = func(ctx context.Context, spaceId, id int64) (*domainmodelsentity.ClientOrigin, error) {
		return &domainmodelsentity.ClientOrigin{Id: id, ClientId: 30, Origin: clientOriginValue, IsActive: true}, nil
	}
	f.ClientOrigin.ReadByClientIdAndOriginFn = func(ctx context.Context, spaceId, clientId int64, origin string) (*domainmodelsentity.ClientOrigin, error) {
		if spaceId != 10 || clientId != 30 || origin != clientOriginValue {
			t.Fatal("incorrect value lookup scope")
		}
		return &domainmodelsentity.ClientOrigin{Id: 50, ClientId: clientId, Origin: origin, IsActive: true}, nil
	}
	f.ClientOrigin.ReadByClientIdFn = func(ctx context.Context, spaceId, clientId int64) ([]domainmodelsentity.ClientOrigin, error) {
		if spaceId != 10 || clientId != 30 {
			t.Fatal("incorrect origin list scope")
		}
		return nil, nil
	}
	f.ClientOrigin.ReadByFilterFn = func(ctx context.Context, spaceId int64, filter domaincontractsrepository.ClientOriginFilter) ([]domainmodelsentity.ClientOrigin, int64, error) {
		if spaceId != 10 || filter.ClientId == nil || *filter.ClientId != 30 {
			t.Fatal("filter lost the requested client")
		}
		return nil, 7, nil
	}
	f.ClientOrigin.CreateFn = func(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateClientOriginArgs) (int64, error) {
		testshared.RequireTx(t, ctx)
		if spaceId != 10 || args.ClientId != 30 || args.Origin != clientOriginValue || args.By == nil || *args.By != f.Caller.UserId {
			t.Fatal("creation scope, origin, or audit actor incorrect")
		}
		return 50, nil
	}
	f.ClientOrigin.UpdateByIdFn = func(ctx context.Context, spaceId, id int64, args domaincontractsrepository.UpdateClientOriginArgs) error {
		testshared.RequireTx(t, ctx)
		if spaceId != 10 || id != 50 || args.By == nil || *args.By != f.Caller.UserId {
			t.Fatal("update scope or audit actor incorrect")
		}
		return nil
	}
	f.ClientOrigin.DeleteByIdFn = func(ctx context.Context, spaceId, id int64, by *int64) error {
		testshared.RequireTx(t, ctx)
		if spaceId != 10 || id != 50 || by == nil || *by != f.Caller.UserId {
			t.Fatal("deletion scope or audit actor incorrect")
		}
		return nil
	}
	return f
}

func mutateOrigin(f *testshared.Fixture, operation string) (int64, error) {
	u := client(f)
	switch operation {
	case "create":
		return u.CreateOrigin(context.Background(), f.Caller, 10, 30, domainusecasesmanagement.CreateClientOriginRequest{Origin: clientOriginValue})
	case "update":
		origin := "https://preview.example.com"
		return 0, u.UpdateOriginById(context.Background(), f.Caller, 10, 30, 50, domainusecasesmanagement.UpdateClientOriginRequest{Origin: &origin})
	default:
		return 0, u.DeleteOriginById(context.Background(), f.Caller, 10, 30, 50)
	}
}

func TestClientOriginMutationsAndCommitLogs(t *testing.T) {
	for _, operation := range []string{"create", "update", "delete"} {
		for _, commitFailure := range []bool{false, true} {
			t.Run(operation+map[bool]string{false: "/success", true: "/commit failure"}[commitFailure], func(t *testing.T) {
				f := originFixture(t)
				readClient := f.Client.ReadByIdFn
				f.Client.ReadByIdFn = func(ctx context.Context, spaceId, id int64) (*domainmodelsentity.Client, error) {
					testshared.RequireTx(t, ctx)
					value, err := readClient(ctx, spaceId, id)
					value.IsActive = false
					return value, err
				}
				readUser := f.User.ReadByIdFn
				f.User.ReadByIdFn = func(ctx context.Context, id int64) (*domainmodelsentity.User, error) {
					testshared.RequireTx(t, ctx)
					return readUser(ctx, id)
				}
				readOrigin := f.ClientOrigin.ReadByIdFn
				f.ClientOrigin.ReadByIdFn = func(ctx context.Context, spaceId, id int64) (*domainmodelsentity.ClientOrigin, error) {
					testshared.RequireTx(t, ctx)
					return readOrigin(ctx, spaceId, id)
				}
				if commitFailure {
					f.CommitErr = testshared.NotFound(domainmodelserror.Failure)
				}
				withTx := f.Transactor.WithTxFn
				f.Transactor.WithTxFn = func(ctx context.Context, fn func(context.Context) error) error {
					return withTx(ctx, func(txCtx context.Context) error {
						err := fn(txCtx)
						if len(f.Logger.Entries) != 0 {
							t.Fatal("success was logged before commit")
						}
						return err
					})
				}
				id, err := mutateOrigin(f, operation)
				if commitFailure {
					if err != f.CommitErr || id != 0 || f.Committed {
						t.Fatalf("failed commit exposed result: id=%d err=%v", id, err)
					}
				} else if err != nil || !f.Committed || operation == "create" && id != 50 {
					t.Fatalf("mutation did not commit: id=%d err=%v", id, err)
				}
				if len(f.Logger.Entries) != 1 || f.Logger.Entries[0].Level != map[bool]string{false: "info", true: "error"}[commitFailure] {
					t.Fatalf("unexpected mutation logs: %+v", f.Logger.Entries)
				}
			})
		}
	}
}

func TestClientOriginMutationStepAndTransactionFailures(t *testing.T) {
	for _, operation := range []string{"create", "update", "delete"} {
		t.Run(operation, func(t *testing.T) {
			f := originFixture(t)
			source := errors.New("database unavailable")
			failure := domainmodelserror.NewError("origin persistence failed", domainmodelserror.Failure, source)
			f.ClientOrigin.CreateFn = func(context.Context, int64, domaincontractsrepository.CreateClientOriginArgs) (int64, error) {
				return 0, failure
			}
			f.ClientOrigin.UpdateByIdFn = func(context.Context, int64, int64, domaincontractsrepository.UpdateClientOriginArgs) error {
				return failure
			}
			f.ClientOrigin.DeleteByIdFn = func(context.Context, int64, int64, *int64) error { return failure }
			_, err := mutateOrigin(f, operation)
			if err != failure || !errors.Is(err, source) || f.Committed {
				t.Fatalf("persistence failure swallowed or transaction committed: %v", err)
			}
			if len(f.Logger.Entries) != 2 || f.Logger.Entries[0].Level != "error" || f.Logger.Entries[1].Level != "error" || f.Logger.Entries[0].Message == f.Logger.Entries[1].Message {
				t.Fatalf("missing distinct step and transaction logs: %+v", f.Logger.Entries)
			}
			for _, entry := range f.Logger.Entries {
				if entry.Meta["err"] != failure || entry.Meta["client_id"] != int64(30) {
					t.Fatal("log lost failure or client scope")
				}
			}
		})
	}
}

func TestClientOriginReadsAndEmptyLists(t *testing.T) {
	f := originFixture(t)
	u := client(f)
	item, err := u.ReadOriginById(context.Background(), f.Caller, 10, 30, 50)
	if err != nil || item.Id != 50 {
		t.Fatalf("id lookup failed: %v", err)
	}
	item, err = u.ReadOriginByValue(context.Background(), f.Caller, 10, 30, clientOriginValue)
	if err != nil || item.Origin != clientOriginValue {
		t.Fatalf("origin lookup failed: %v", err)
	}
	items, err := u.ReadOriginsByClientId(context.Background(), f.Caller, 10, 30)
	if err != nil || items == nil || len(items) != 0 {
		t.Fatalf("incorrect empty list: %v", err)
	}
	page, err := u.ReadOriginsByFilter(context.Background(), f.Caller, 10, 30, domainusecasesmanagement.FilterClientOriginRequest{Page: 999, Limit: 10})
	if err != nil || page.Items == nil || len(page.Items) != 0 || page.Total != 7 {
		t.Fatalf("empty page lost total: %+v %v", page, err)
	}
	if len(f.Logger.Entries) != 0 || f.Transactions != 0 {
		t.Fatal("read emitted success logs or started a transaction")
	}
}

func TestClientOriginCrossClientIsolation(t *testing.T) {
	for _, operation := range []string{"read", "update", "delete", "value"} {
		t.Run(operation, func(t *testing.T) {
			f := originFixture(t)
			f.ClientOrigin.UpdateByIdFn = nil
			f.ClientOrigin.DeleteByIdFn = nil
			f.ClientOrigin.ReadByIdFn = func(context.Context, int64, int64) (*domainmodelsentity.ClientOrigin, error) {
				return &domainmodelsentity.ClientOrigin{Id: 50, ClientId: 31}, nil
			}
			f.ClientOrigin.ReadByClientIdAndOriginFn = func(context.Context, int64, int64, string) (*domainmodelsentity.ClientOrigin, error) {
				return &domainmodelsentity.ClientOrigin{Id: 50, ClientId: 31, Origin: clientOriginValue}, nil
			}
			var err error
			if operation == "read" {
				_, err = client(f).ReadOriginById(context.Background(), f.Caller, 10, 30, 50)
			} else if operation == "value" {
				_, err = client(f).ReadOriginByValue(context.Background(), f.Caller, 10, 30, clientOriginValue)
			} else {
				_, err = mutateOrigin(f, operation)
			}
			testshared.RequireError(t, err, domainmodelserror.ClientOriginNotFound)
		})
	}
}

func TestClientOriginManagementValidation(t *testing.T) {
	for _, operation := range []string{"create invalid origin", "update invalid origin", "invalid id", "pagination overflow", "nil client", "nil origin", "nil value result", "wrong client space", "invalid created id"} {
		t.Run(operation, func(t *testing.T) {
			f := originFixture(t)
			u := client(f)
			kind := domainmodelserror.Failure
			var err error
			switch operation {
			case "create invalid origin":
				f.Client.ReadByIdFn = nil
				_, err = u.CreateOrigin(context.Background(), f.Caller, 10, 30, domainusecasesmanagement.CreateClientOriginRequest{Origin: "http://example.com"})
				kind = domainmodelserror.ClientOriginInvalid
			case "update invalid origin":
				f.Client.ReadByIdFn = nil
				value := "https://example.com/"
				err = u.UpdateOriginById(context.Background(), f.Caller, 10, 30, 50, domainusecasesmanagement.UpdateClientOriginRequest{Origin: &value})
				kind = domainmodelserror.ClientOriginInvalid
			case "invalid id":
				f.Client.ReadByIdFn = nil
				_, err = u.ReadOriginById(context.Background(), f.Caller, 10, 0, 50)
				kind = domainmodelserror.BadArgs
			case "pagination overflow":
				f.Client.ReadByIdFn = nil
				_, err = u.ReadOriginsByFilter(context.Background(), f.Caller, 10, 30, domainusecasesmanagement.FilterClientOriginRequest{Page: math.MaxInt64, Limit: 2})
				kind = domainmodelserror.BadArgs
			case "nil client":
				f.Client.ReadByIdFn = func(context.Context, int64, int64) (*domainmodelsentity.Client, error) { return nil, nil }
				_, err = u.ReadOriginsByClientId(context.Background(), f.Caller, 10, 30)
			case "nil origin":
				f.ClientOrigin.ReadByIdFn = func(context.Context, int64, int64) (*domainmodelsentity.ClientOrigin, error) { return nil, nil }
				_, err = u.ReadOriginById(context.Background(), f.Caller, 10, 30, 50)
			case "nil value result":
				f.ClientOrigin.ReadByClientIdAndOriginFn = func(context.Context, int64, int64, string) (*domainmodelsentity.ClientOrigin, error) { return nil, nil }
				_, err = u.ReadOriginByValue(context.Background(), f.Caller, 10, 30, clientOriginValue)
			case "wrong client space":
				f.Client.ReadByIdFn = func(context.Context, int64, int64) (*domainmodelsentity.Client, error) {
					return &domainmodelsentity.Client{Id: 30, SpaceId: 11}, nil
				}
				_, err = u.ReadOriginsByClientId(context.Background(), f.Caller, 10, 30)
				kind = domainmodelserror.ClientNotFound
			case "invalid created id":
				f.ClientOrigin.CreateFn = func(context.Context, int64, domaincontractsrepository.CreateClientOriginArgs) (int64, error) {
					return 0, nil
				}
				_, err = u.CreateOrigin(context.Background(), f.Caller, 10, 30, domainusecasesmanagement.CreateClientOriginRequest{Origin: clientOriginValue})
			}
			testshared.RequireError(t, err, kind)
			if f.Committed || len(f.Logger.Entries) == 0 {
				t.Fatal("failed operation committed or omitted error logs")
			}
		})
	}
}

func TestClientOriginControlSpaceAuthority(t *testing.T) {
	f := originFixture(t)
	f.ClientOrigin.CreateFn = func(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateClientOriginArgs) (int64, error) {
		testshared.RequireTx(t, ctx)
		if spaceId != 11 {
			t.Fatal("control authority lost target scope")
		}
		return 50, nil
	}
	_, err := client(f).CreateOrigin(context.Background(), f.Caller, 11, 30, domainusecasesmanagement.CreateClientOriginRequest{Origin: clientOriginValue})
	if err != nil {
		t.Fatalf("control-space authority denied: %v", err)
	}
}

func TestClientOriginAuditOnlyUpdateStaysQuiet(t *testing.T) {
	f := originFixture(t)
	err := client(f).UpdateOriginById(context.Background(), f.Caller, 10, 30, 50, domainusecasesmanagement.UpdateClientOriginRequest{})
	if err != nil || !f.Committed || len(f.Logger.Entries) != 0 {
		t.Fatalf("audit-only update behavior: %v", err)
	}
}
