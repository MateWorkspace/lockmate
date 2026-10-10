package profile_test

import (
	"context"
	"errors"
	applicationprofileaccount "github.com/MateWorkspace/lockmate/internal/application/profile/account"
	applicationprofilesecurity "github.com/MateWorkspace/lockmate/internal/application/profile/security"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesprofile "github.com/MateWorkspace/lockmate/internal/domain/usecases/profile"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
	"strings"
	"testing"
)

func TestEveryProtectedOperation(t *testing.T) {
	cases := []struct {
		name, permission string
		scoped           bool
		call             func(*testshared.Fixture) error
	}{
		{name: "account/GetProfile", permission: "profile:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationprofileaccount.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Validator, f.Transactor, f.Logger).GetProfile(context.Background(), f.Caller)
			return err
		}},
		{name: "account/GetRoles", permission: "profile:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationprofileaccount.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Validator, f.Transactor, f.Logger).GetRoles(context.Background(), f.Caller)
			return err
		}},
		{name: "account/GetPermissions", permission: "profile:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationprofileaccount.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Validator, f.Transactor, f.Logger).GetPermissions(context.Background(), f.Caller)
			return err
		}},
		{name: "account/UpdateProfile", permission: "profile:set", scoped: false, call: func(f *testshared.Fixture) error {
			err := applicationprofileaccount.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Validator, f.Transactor, f.Logger).UpdateProfile(context.Background(), f.Caller, domainusecasesprofile.UpdateProfileRequest{})
			return err
		}},
		{name: "account/UpdateMembershipPreferences", permission: "profile:set", scoped: false, call: func(f *testshared.Fixture) error {
			err := applicationprofileaccount.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Validator, f.Transactor, f.Logger).UpdateMembershipPreferences(context.Background(), f.Caller, domainusecasesprofile.UpdateMembershipPreferencesRequest{})
			return err
		}},
		{name: "security/GetAuthMethods", permission: "profile_security:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationprofilesecurity.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.GoogleAuth, f.Password, f.Guard, f.Validator, f.Transactor, f.Logger).GetAuthMethods(context.Background(), f.Caller)
			return err
		}},
		{name: "security/UpdateClassicAuth", permission: "profile_security:set", scoped: false, call: func(f *testshared.Fixture) error {
			err := applicationprofilesecurity.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.GoogleAuth, f.Password, f.Guard, f.Validator, f.Transactor, f.Logger).UpdateClassicAuth(context.Background(), f.Caller, domainusecasesprofile.UpdateClassicAuthRequest{})
			return err
		}},
		{name: "security/ChangePassword", permission: "profile_security:set", scoped: false, call: func(f *testshared.Fixture) error {
			err := applicationprofilesecurity.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.GoogleAuth, f.Password, f.Guard, f.Validator, f.Transactor, f.Logger).ChangePassword(context.Background(), f.Caller, domainusecasesprofile.ChangePasswordRequest{})
			return err
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Run("permission denied", func(t *testing.T) {
				f := testshared.NewFixture()
				f.User.ReadByIdFn = nil
				f.Space.ReadByIdFn = nil
				f.SpaceMember.ReadByIdFn = nil
				f.Transactor.WithTxFn = nil
				denied := domainmodelserror.NewError("do not log secret-content", domainmodelserror.Forbidden, errors.New("private-secret"))
				called := false
				f.Guard.CheckFn = func(caller *domainmodelsentity.AccessClaims, permissions []string) error {
					called = true
					if caller != f.Caller || len(permissions) != 1 || permissions[0] != test.permission {
						t.Fatalf("wrong guard arguments: %v", permissions)
					}
					return denied
				}
				err := test.call(f)
				if err != denied || !called {
					t.Fatalf("permission failure changed or guard skipped: %v", err)
				}
				if len(f.Logger.Entries) != 1 || f.Logger.Entries[0].Level != "error" {
					t.Fatalf("expected one error log: %v", f.Logger.Entries)
				}
				if f.Logger.Entries[0].Meta["err"] != denied || !strings.HasPrefix(f.Logger.Entries[0].Message, "permission denied to ") {
					t.Fatal("permission failure lacks specific message or original error")
				}
				if f.Transactions != 0 {
					t.Fatal("permission denial started a transaction")
				}
			})
			t.Run("nil caller", func(t *testing.T) {
				f := testshared.NewFixture()
				f.Caller = nil
				err := test.call(f)
				testshared.RequireError(t, err, domainmodelserror.BadArgs)
				if len(f.Logger.Entries) != 1 || f.Logger.Entries[0].Level != "error" {
					t.Fatal("missing single error log")
				}
			})

		})
	}
}
