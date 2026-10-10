package management_test

import (
	"context"
	"testing"

	applicationmanagementclassicauth "github.com/MateWorkspace/lockmate/internal/application/management/classic_auth"
	applicationmanagementclient "github.com/MateWorkspace/lockmate/internal/application/management/client"
	applicationmanagementrole "github.com/MateWorkspace/lockmate/internal/application/management/role"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
)

func classic(f *testshared.Fixture) domainusecasesmanagement.ClassicAuth {
	return applicationmanagementclassicauth.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.GoogleAuth, f.Guard, f.Validator, f.Logger, f.Transactor, f.Password, "")
}
func client(f *testshared.Fixture) domainusecasesmanagement.Client {
	return applicationmanagementclient.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Client, f.Guard, f.Validator, f.Logger, f.Transactor, f.Keygen, "")
}
func TestAdministratorPasswordResetAndLastMethod(t *testing.T) {
	f := testshared.NewFixture()
	f.Password.CompareFn = nil
	err := classic(f).ResetPassword(context.Background(), f.Caller, 1, domainusecasesmanagement.ResetUserPasswordRequest{Password: "replacement123"})
	if err != nil || f.Classics[1].PasswordHash != "hash:replacement123" {
		t.Fatalf("reset required old password: %v", err)
	}
	err = classic(f).DeleteById(context.Background(), f.Caller, 40)
	testshared.RequireError(t, err, domainmodelserror.BadState)
	if f.Classics[1] == nil {
		t.Fatal("removed last method")
	}
	f.Googles[1] = &domainmodelsentity.GoogleAuth{Id: 50, UserId: 1, Subject: "google"}
	err = classic(f).DeleteById(context.Background(), f.Caller, 40)
	if err != nil || f.Classics[1] != nil {
		t.Fatalf("remaining method deletion: %v", err)
	}
}
func TestClientSecretAndCommitFailure(t *testing.T) {
	for _, operation := range []string{"create", "rotate"} {
		for _, fail := range []bool{false, true} {
			t.Run(operation+map[bool]string{false: "/commit", true: "/failure"}[fail], func(t *testing.T) {
				f := testshared.NewFixture()
				if fail {
					f.CommitErr = testshared.NotFound(domainmodelserror.Failure)
				}
				f.Client.CreateFn = func(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateClientArgs) (int64, error) {
					testshared.RequireTx(t, ctx)
					if spaceId != 10 || args.By == nil || *args.By != 1 || args.SecretKeyHash != "hash:secret" || args.SecretKeyRedacted != "reda****cted" {
						t.Fatal("wrong client persistence")
					}
					return 30, nil
				}
				f.Client.UpdateByIdFn = func(ctx context.Context, spaceId, id int64, args domaincontractsrepository.UpdateClientArgs) error {
					testshared.RequireTx(t, ctx)
					if spaceId != 10 || id != 30 || args.By == nil || *args.By != 1 || args.SecretKeyHash == nil || args.SecretKeyRedacted == nil {
						t.Fatal("rotation scope or fields wrong")
					}
					return nil
				}
				var result *domainusecasesmanagement.ClientSecretResult
				var err error
				if operation == "create" {
					result, err = client(f).Create(context.Background(), f.Caller, 10, domainusecasesmanagement.CreateClientRequest{Slug: "web", Name: "Web client"})
				} else {
					result, err = client(f).RotateSecret(context.Background(), f.Caller, 10, 30)
				}
				if fail {
					if result != nil || err == nil || f.Logger.Entries[0].Level != "error" {
						t.Fatal("failed commit exposed secret or success")
					}
				} else if err != nil || result.Secret != "raw-secret" || !f.Committed || f.Logger.Entries[0].Level != "info" {
					t.Fatalf("secret result: %v", err)
				}
				if len(f.Logger.Entries) != 1 {
					t.Fatal("wrong log count")
				}
			})
		}
	}
}
func TestDefaultRolePreservesScopeActorAndTransaction(t *testing.T) {
	f := testshared.NewFixture()
	called := false
	f.Role.UpdateByIdFn = func(ctx context.Context, spaceId, id int64, args domaincontractsrepository.UpdateRoleArgs) error {
		testshared.RequireTx(t, ctx)
		called = true
		if spaceId != 10 || id != 60 || args.IsDefault == nil || !*args.IsDefault || args.By == nil || *args.By != 1 || len(f.Logger.Entries) != 0 {
			t.Fatal("default role mapping/commit timing")
		}
		return nil
	}
	u := applicationmanagementrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Role, f.Guard, f.Validator, f.Logger, f.Transactor, "")
	if err := u.SetDefault(context.Background(), f.Caller, 10, 60); err != nil || !called || !f.Committed {
		t.Fatalf("set default: %v", err)
	}
	if len(f.Logger.Entries) != 1 || f.Logger.Entries[0].Level != "info" {
		t.Fatal("wrong success log")
	}
}
func TestClientSlugCannotCrossSpace(t *testing.T) {
	f := testshared.NewFixture()
	f.Client.ReadBySlugFn = func(context.Context, string) (*domainmodelsentity.Client, error) {
		return &domainmodelsentity.Client{Id: 30, SpaceId: 11, Slug: "web"}, nil
	}
	result, err := client(f).ReadBySlug(context.Background(), f.Caller, 10, "web")
	testshared.RequireError(t, err, domainmodelserror.ClientNotFound)
	if result != nil {
		t.Fatal("cross-space result leaked")
	}
}
