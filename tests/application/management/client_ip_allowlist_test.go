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

const clientIpAllowlistValue = "203.0.113.0/24"

func ipAllowlistFixture(t *testing.T) *testshared.Fixture {
	t.Helper()
	f := testshared.NewFixture()
	f.ClientIpAllowlist.ReadByIdFn = func(ctx context.Context, spaceId, id int64) (*domainmodelsentity.ClientIpAllowlist, error) {
		return &domainmodelsentity.ClientIpAllowlist{Id: id, ClientId: 30, Cidr: clientIpAllowlistValue, IsActive: true}, nil
	}
	f.ClientIpAllowlist.ReadByClientIdAndCidrFn = func(ctx context.Context, spaceId, clientId int64, cidr string) (*domainmodelsentity.ClientIpAllowlist, error) {
		if spaceId != 10 || clientId != 30 || cidr != clientIpAllowlistValue {
			t.Fatal("incorrect value lookup scope")
		}
		return &domainmodelsentity.ClientIpAllowlist{Id: 50, ClientId: clientId, Cidr: cidr, IsActive: true}, nil
	}
	f.ClientIpAllowlist.ReadByClientIdFn = func(ctx context.Context, spaceId, clientId int64) ([]domainmodelsentity.ClientIpAllowlist, error) {
		if spaceId != 10 || clientId != 30 {
			t.Fatal("incorrect IP allowlist list scope")
		}
		return nil, nil
	}
	f.ClientIpAllowlist.ReadByFilterFn = func(ctx context.Context, spaceId int64, filter domaincontractsrepository.ClientIpAllowlistFilter) ([]domainmodelsentity.ClientIpAllowlist, int64, error) {
		if spaceId != 10 || filter.ClientId == nil || *filter.ClientId != 30 {
			t.Fatal("filter lost the requested client")
		}
		return nil, 7, nil
	}
	f.ClientIpAllowlist.CreateFn = func(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateClientIpAllowlistArgs) (int64, error) {
		testshared.RequireTx(t, ctx)
		if spaceId != 10 || args.ClientId != 30 || args.Cidr != clientIpAllowlistValue || args.By == nil || *args.By != f.Caller.UserId {
			t.Fatal("creation scope, CIDR, or audit actor incorrect")
		}
		return 50, nil
	}
	f.ClientIpAllowlist.UpdateByIdFn = func(ctx context.Context, spaceId, id int64, args domaincontractsrepository.UpdateClientIpAllowlistArgs) error {
		testshared.RequireTx(t, ctx)
		if spaceId != 10 || id != 50 || args.By == nil || *args.By != f.Caller.UserId {
			t.Fatal("update scope or audit actor incorrect")
		}
		return nil
	}
	f.ClientIpAllowlist.DeleteByIdFn = func(ctx context.Context, spaceId, id int64, by *int64) error {
		testshared.RequireTx(t, ctx)
		if spaceId != 10 || id != 50 || by == nil || *by != f.Caller.UserId {
			t.Fatal("deletion scope or audit actor incorrect")
		}
		return nil
	}
	return f
}

func mutateIpAllowlist(f *testshared.Fixture, operation string) (int64, error) {
	u := client(f)
	switch operation {
	case "create":
		return u.CreateIpAllowlist(context.Background(), f.Caller, 10, 30, domainusecasesmanagement.CreateClientIpAllowlistRequest{Cidr: clientIpAllowlistValue})
	case "update":
		cidr := "198.51.100.0/24"
		return 0, u.UpdateIpAllowlistById(context.Background(), f.Caller, 10, 30, 50, domainusecasesmanagement.UpdateClientIpAllowlistRequest{Cidr: &cidr})
	default:
		return 0, u.DeleteIpAllowlistById(context.Background(), f.Caller, 10, 30, 50)
	}
}

func TestClientIpAllowlistMutationsAndCommitLogs(t *testing.T) {
	for _, operation := range []string{"create", "update", "delete"} {
		for _, commitFailure := range []bool{false, true} {
			t.Run(operation+map[bool]string{false: "/success", true: "/commit failure"}[commitFailure], func(t *testing.T) {
				f := ipAllowlistFixture(t)
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
				readItem := f.ClientIpAllowlist.ReadByIdFn
				f.ClientIpAllowlist.ReadByIdFn = func(ctx context.Context, spaceId, id int64) (*domainmodelsentity.ClientIpAllowlist, error) {
					testshared.RequireTx(t, ctx)
					return readItem(ctx, spaceId, id)
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
				id, err := mutateIpAllowlist(f, operation)
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

func TestClientIpAllowlistMutationStepAndTransactionFailures(t *testing.T) {
	for _, operation := range []string{"create", "update", "delete"} {
		t.Run(operation, func(t *testing.T) {
			f := ipAllowlistFixture(t)
			source := errors.New("database unavailable")
			failure := domainmodelserror.NewError("IP allowlist persistence failed", domainmodelserror.Failure, source)
			f.ClientIpAllowlist.CreateFn = func(context.Context, int64, domaincontractsrepository.CreateClientIpAllowlistArgs) (int64, error) {
				return 0, failure
			}
			f.ClientIpAllowlist.UpdateByIdFn = func(context.Context, int64, int64, domaincontractsrepository.UpdateClientIpAllowlistArgs) error {
				return failure
			}
			f.ClientIpAllowlist.DeleteByIdFn = func(context.Context, int64, int64, *int64) error { return failure }
			_, err := mutateIpAllowlist(f, operation)
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

func TestClientIpAllowlistReadsAndEmptyLists(t *testing.T) {
	f := ipAllowlistFixture(t)
	u := client(f)
	item, err := u.ReadIpAllowlistById(context.Background(), f.Caller, 10, 30, 50)
	if err != nil || item.Id != 50 {
		t.Fatalf("id lookup failed: %v", err)
	}
	item, err = u.ReadIpAllowlistByCidr(context.Background(), f.Caller, 10, 30, clientIpAllowlistValue)
	if err != nil || item.Cidr != clientIpAllowlistValue {
		t.Fatalf("CIDR lookup failed: %v", err)
	}
	items, err := u.ReadIpAllowlistsByClientId(context.Background(), f.Caller, 10, 30)
	if err != nil || items == nil || len(items) != 0 {
		t.Fatalf("incorrect empty list: %v", err)
	}
	page, err := u.ReadIpAllowlistsByFilter(context.Background(), f.Caller, 10, 30, domainusecasesmanagement.FilterClientIpAllowlistRequest{Page: 999, Limit: 10})
	if err != nil || page.Items == nil || len(page.Items) != 0 || page.Total != 7 {
		t.Fatalf("empty page lost total: %+v %v", page, err)
	}
	if len(f.Logger.Entries) != 0 || f.Transactions != 0 {
		t.Fatal("read emitted success logs or started a transaction")
	}
}

func TestClientIpAllowlistCrossClientIsolation(t *testing.T) {
	for _, operation := range []string{"read", "update", "delete", "value"} {
		t.Run(operation, func(t *testing.T) {
			f := ipAllowlistFixture(t)
			f.ClientIpAllowlist.UpdateByIdFn = nil
			f.ClientIpAllowlist.DeleteByIdFn = nil
			f.ClientIpAllowlist.ReadByIdFn = func(context.Context, int64, int64) (*domainmodelsentity.ClientIpAllowlist, error) {
				return &domainmodelsentity.ClientIpAllowlist{Id: 50, ClientId: 31}, nil
			}
			f.ClientIpAllowlist.ReadByClientIdAndCidrFn = func(context.Context, int64, int64, string) (*domainmodelsentity.ClientIpAllowlist, error) {
				return &domainmodelsentity.ClientIpAllowlist{Id: 50, ClientId: 31, Cidr: clientIpAllowlistValue}, nil
			}
			var err error
			if operation == "read" {
				_, err = client(f).ReadIpAllowlistById(context.Background(), f.Caller, 10, 30, 50)
			} else if operation == "value" {
				_, err = client(f).ReadIpAllowlistByCidr(context.Background(), f.Caller, 10, 30, clientIpAllowlistValue)
			} else {
				_, err = mutateIpAllowlist(f, operation)
			}
			testshared.RequireError(t, err, domainmodelserror.ClientIpAllowlistNotFound)
		})
	}
}

func TestClientIpAllowlistManagementValidation(t *testing.T) {
	for _, operation := range []string{"create invalid cidr", "update invalid cidr", "invalid id", "pagination overflow", "nil client", "nil entry", "nil value result", "wrong client space", "invalid created id"} {
		t.Run(operation, func(t *testing.T) {
			f := ipAllowlistFixture(t)
			u := client(f)
			kind := domainmodelserror.Failure
			var err error
			switch operation {
			case "create invalid cidr":
				f.Client.ReadByIdFn = nil
				_, err = u.CreateIpAllowlist(context.Background(), f.Caller, 10, 30, domainusecasesmanagement.CreateClientIpAllowlistRequest{Cidr: "300.0.0.0/8"})
				kind = domainmodelserror.ClientIpAllowlistInvalid
			case "update invalid cidr":
				f.Client.ReadByIdFn = nil
				value := "203.0.113.5/24"
				err = u.UpdateIpAllowlistById(context.Background(), f.Caller, 10, 30, 50, domainusecasesmanagement.UpdateClientIpAllowlistRequest{Cidr: &value})
				kind = domainmodelserror.ClientIpAllowlistInvalid
			case "invalid id":
				f.Client.ReadByIdFn = nil
				_, err = u.ReadIpAllowlistById(context.Background(), f.Caller, 10, 0, 50)
				kind = domainmodelserror.BadArgs
			case "pagination overflow":
				f.Client.ReadByIdFn = nil
				_, err = u.ReadIpAllowlistsByFilter(context.Background(), f.Caller, 10, 30, domainusecasesmanagement.FilterClientIpAllowlistRequest{Page: math.MaxInt64, Limit: 2})
				kind = domainmodelserror.BadArgs
			case "nil client":
				f.Client.ReadByIdFn = func(context.Context, int64, int64) (*domainmodelsentity.Client, error) { return nil, nil }
				_, err = u.ReadIpAllowlistsByClientId(context.Background(), f.Caller, 10, 30)
			case "nil entry":
				f.ClientIpAllowlist.ReadByIdFn = func(context.Context, int64, int64) (*domainmodelsentity.ClientIpAllowlist, error) { return nil, nil }
				_, err = u.ReadIpAllowlistById(context.Background(), f.Caller, 10, 30, 50)
			case "nil value result":
				f.ClientIpAllowlist.ReadByClientIdAndCidrFn = func(context.Context, int64, int64, string) (*domainmodelsentity.ClientIpAllowlist, error) {
					return nil, nil
				}
				_, err = u.ReadIpAllowlistByCidr(context.Background(), f.Caller, 10, 30, clientIpAllowlistValue)
			case "wrong client space":
				f.Client.ReadByIdFn = func(context.Context, int64, int64) (*domainmodelsentity.Client, error) {
					return &domainmodelsentity.Client{Id: 30, SpaceId: 11}, nil
				}
				_, err = u.ReadIpAllowlistsByClientId(context.Background(), f.Caller, 10, 30)
				kind = domainmodelserror.ClientNotFound
			case "invalid created id":
				f.ClientIpAllowlist.CreateFn = func(context.Context, int64, domaincontractsrepository.CreateClientIpAllowlistArgs) (int64, error) {
					return 0, nil
				}
				_, err = u.CreateIpAllowlist(context.Background(), f.Caller, 10, 30, domainusecasesmanagement.CreateClientIpAllowlistRequest{Cidr: clientIpAllowlistValue})
			}
			testshared.RequireError(t, err, kind)
			if f.Committed || len(f.Logger.Entries) == 0 {
				t.Fatal("failed operation committed or omitted error logs")
			}
		})
	}
}

func TestClientIpAllowlistControlSpaceAuthority(t *testing.T) {
	f := ipAllowlistFixture(t)
	f.ClientIpAllowlist.CreateFn = func(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateClientIpAllowlistArgs) (int64, error) {
		testshared.RequireTx(t, ctx)
		if spaceId != 11 {
			t.Fatal("control authority lost target scope")
		}
		return 50, nil
	}
	_, err := client(f).CreateIpAllowlist(context.Background(), f.Caller, 11, 30, domainusecasesmanagement.CreateClientIpAllowlistRequest{Cidr: clientIpAllowlistValue})
	if err != nil {
		t.Fatalf("control-space authority denied: %v", err)
	}
}

func TestClientIpAllowlistAuditOnlyUpdateStaysQuiet(t *testing.T) {
	f := ipAllowlistFixture(t)
	err := client(f).UpdateIpAllowlistById(context.Background(), f.Caller, 10, 30, 50, domainusecasesmanagement.UpdateClientIpAllowlistRequest{})
	if err != nil || !f.Committed || len(f.Logger.Entries) != 0 {
		t.Fatalf("audit-only update behavior: %v", err)
	}
}

func TestClientCreateAddsLoopbackAllowlist(t *testing.T) {
	f := ipAllowlistFixture(t)
	var created []string
	f.Client.CreateFn = func(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateClientArgs) (int64, error) {
		testshared.RequireTx(t, ctx)
		return 40, nil
	}
	f.ClientIpAllowlist.CreateFn = func(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateClientIpAllowlistArgs) (int64, error) {
		testshared.RequireTx(t, ctx)
		if spaceId != 10 || args.ClientId != 40 || args.By == nil || *args.By != f.Caller.UserId {
			t.Fatalf("unexpected loopback creation scope: %d %d", spaceId, args.ClientId)
		}
		created = append(created, args.Cidr)
		return int64(90 + len(created)), nil
	}
	result, err := client(f).Create(context.Background(), f.Caller, 10, domainusecasesmanagement.CreateClientRequest{Slug: "new", Name: "New"})
	if err != nil || result == nil || result.Secret == "" {
		t.Fatalf("client create failed: %v", err)
	}
	if len(created) != 2 || created[0] != "127.0.0.1/32" || created[1] != "::1/128" {
		t.Fatalf("loopback entries not created: %v", created)
	}
}
