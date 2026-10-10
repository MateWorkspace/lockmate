package auth_test

import (
	"context"
	"errors"
	"testing"

	applicationauthsession "github.com/MateWorkspace/lockmate/internal/application/auth/session"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
)

func session(f *testshared.Fixture) domainusecasesauth.Session {
	return applicationauthsession.NewUsecaseImpl(f.Client, f.Space, f.User, f.ClassicAuth, f.SpaceMember, f.SpaceMemberRole, f.Password, f.Token, f.Validator, f.Transactor, f.Logger)
}
func clientIdentity() *domainmodelsapp.ClientIdentity {
	return &domainmodelsapp.ClientIdentity{Id: 30, SpaceId: 10, Slug: "web"}
}
func registerRequest() domainusecasesauth.RegisterRequest {
	return domainusecasesauth.RegisterRequest{Name: "New user", Username: "new-user", Password: "password123"}
}

func TestRegistrationAtomicResult(t *testing.T) {
	for _, failure := range []string{"", "default role", "access token", "refresh token", "commit", "nested"} {
		t.Run(failure, func(t *testing.T) {
			f := testshared.NewFixture()
			ctx := context.Background()
			want := domainmodelserror.NewError("private-source", domainmodelserror.Failure, errors.New("secret-value"))
			switch failure {
			case "default role":
				f.MissingDefault = true
				want = testshared.NotFound(domainmodelserror.DefaultRoleNotFound)
			case "access token":
				f.TokenErr = want
			case "refresh token":
				f.Token.GenerateRefreshFn = func(*domainmodelsentity.RefreshClaims) (string, error) { return "", want }
			case "commit":
				f.CommitErr = want
			case "nested":
				ctx = testshared.InTransaction()
				want = testshared.NotFound(domainmodelserror.BadState)
			}
			create := f.User.CreateFn
			f.User.CreateFn = func(ctx context.Context, args domaincontractsrepository.CreateUserArgs) (int64, error) {
				testshared.RequireTx(t, ctx)
				if args.By != nil {
					t.Fatal("anonymous audit override")
				}
				return create(ctx, args)
			}
			result, err := session(f).Register(ctx, clientIdentity(), registerRequest())
			if failure != "" {
				testshared.RequireError(t, err, want.(*domainmodelserror.Error).Type)
				if result != nil || len(f.Users) != 1 || len(f.Classics) != 1 || len(f.Members) != 1 {
					t.Fatal("failed registration retained result or writes")
				}
				expected := 2
				if failure == "commit" || failure == "nested" {
					expected = 1
				}
				if len(f.Logger.Entries) != expected || f.Logger.Entries[expected-1].Message != "failed to complete registration transaction" {
					t.Fatalf("wrong registration failure logs: %v", f.Logger.Entries)
				}
				for _, entry := range f.Logger.Entries {
					if entry.Level != "error" || entry.Meta["err"] == nil {
						t.Fatal("missing explicit error metadata")
					}
				}
				return
			}
			if err != nil || result == nil || !result.IsNewUser || result.AccessToken == "" || result.RefreshToken == "" || result.User.Id != 2 || !f.Committed {
				t.Fatalf("registration result: %v %v", result, err)
			}
			if len(f.Users) != 2 || len(f.Classics) != 2 || len(f.Members) != 2 {
				t.Fatal("registration incomplete")
			}
			if len(f.Logger.Entries) != 1 || f.Logger.Entries[0].Level != "info" {
				t.Fatal("missing registration success")
			}
		})
	}
}

func TestSignInMembershipAndFailures(t *testing.T) {
	for _, state := range []string{"existing", "missing", "suspended", "wrong password", "unknown user", "outage", "conflict"} {
		t.Run(state, func(t *testing.T) {
			f := testshared.NewFixture()
			request := domainusecasesauth.SignInRequest{Username: "user", Password: "password"}
			switch state {
			case "missing", "conflict":
				delete(f.Members, 20)
			case "suspended":
				f.Members[20].IsActive = false
			case "wrong password":
				request.Password = "wrong"
			case "unknown user":
				request.Username = "unknown"
			case "outage":
				f.ClassicAuth.ReadByUsernameFn = func(context.Context, string) (*domainmodelsentity.ClassicAuth, error) {
					return nil, testshared.NotFound(domainmodelserror.Timeout)
				}
			}
			if state == "conflict" {
				create := f.SpaceMember.CreateFn
				attempt := 0
				f.SpaceMember.CreateFn = func(ctx context.Context, id int64, args domaincontractsrepository.CreateSpaceMemberArgs) (int64, error) {
					attempt++
					if attempt == 1 {
						return 0, testshared.NotFound(domainmodelserror.SpaceMemberConflict)
					}
					return create(ctx, id, args)
				}
			}
			result, err := session(f).SignIn(context.Background(), clientIdentity(), request)
			if state == "suspended" || state == "wrong password" || state == "unknown user" {
				testshared.RequireError(t, err, domainmodelserror.Unauthorized)
				if result != nil {
					t.Fatal("failed sign-in returned tokens")
				}
				return
			}
			if state == "outage" {
				testshared.RequireError(t, err, domainmodelserror.Timeout)
				return
			}
			if err != nil || result == nil || result.IsNewUser {
				t.Fatalf("sign-in failed: %v", err)
			}
			if len(f.Members) != 1 {
				t.Fatal("duplicate/missing membership")
			}
			expectedLogs := 0
			if state != "existing" {
				expectedLogs = 1
			}
			if state == "conflict" {
				expectedLogs += 2
				if f.Logger.Entries[0].Level != "warn" || f.Logger.Entries[0].Message != "failed to ensure space membership" || f.Logger.Entries[1].Meta["next_attempt"] != 2 {
					t.Fatal("missing retry warning")
				}
			}
			if len(f.Logger.Entries) != expectedLogs {
				t.Fatal("wrong sign-in logging")
			}
			if state == "conflict" && f.Transactions != 2 {
				t.Fatal("conflict did not retry in new transaction")
			}
		})
	}
}

func TestRefreshAndAuthenticationReloadGrants(t *testing.T) {
	f := testshared.NewFixture()
	access, err := session(f).AuthenticateAccessToken(context.Background(), clientIdentity(), "access")
	if err != nil || access.Name != "Current name" || len(access.Permissions) != 1 || access.Permissions[0] != "profile:get" {
		t.Fatalf("stale claims used: %v %v", access, err)
	}
	result, err := session(f).Refresh(context.Background(), clientIdentity(), domainusecasesauth.RefreshRequest{RefreshToken: "refresh"})
	if err != nil || result.AccessToken == "" || f.Transactions != 0 || len(f.Logger.Entries) != 0 {
		t.Fatalf("read-only refresh behavior: %v", err)
	}
	delete(f.Members, 20)
	result, err = session(f).Refresh(context.Background(), clientIdentity(), domainusecasesauth.RefreshRequest{RefreshToken: "refresh"})
	testshared.RequireError(t, err, domainmodelserror.Unauthorized)
	if result != nil || len(f.Members) != 0 {
		t.Fatal("refresh enrolled missing membership")
	}
}
