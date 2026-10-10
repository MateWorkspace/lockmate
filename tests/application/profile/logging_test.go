package profile_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	applicationprofileaccount "github.com/MateWorkspace/lockmate/internal/application/profile/account"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesprofile "github.com/MateWorkspace/lockmate/internal/domain/usecases/profile"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
)

func TestCosmeticProfileUpdatesStayQuiet(t *testing.T) {
	f := testshared.NewFixture()
	f.SpaceMember.UpdateByIdFn = func(ctx context.Context, spaceId, id int64, args domaincontractsrepository.UpdateSpaceMemberArgs) error {
		testshared.RequireTx(t, ctx)
		if spaceId != 10 || id != 20 || args.Preferences == nil || len(f.Logger.Entries) != 0 {
			t.Fatal("incorrect membership update or premature log")
		}
		return nil
	}
	u := applicationprofileaccount.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Validator, f.Transactor, f.Logger)
	name := "Updated name"
	if err := u.UpdateProfile(context.Background(), f.Caller, domainusecasesprofile.UpdateProfileRequest{Name: &name}); err != nil {
		t.Fatal(err)
	}
	if err := u.UpdateMembershipPreferences(context.Background(), f.Caller, domainusecasesprofile.UpdateMembershipPreferencesRequest{Preferences: json.RawMessage(`{"theme":"dark"}`)}); err != nil {
		t.Fatal(err)
	}
	if !f.Committed || len(f.Logger.Entries) != 0 {
		t.Fatal("cosmetic updates logged success")
	}
}

func TestClassicCredentialUpdateLogsSuccess(t *testing.T) {
	f := testshared.NewFixture()
	username := "updateduser"
	if err := security(f).UpdateClassicAuth(context.Background(), f.Caller, domainusecasesprofile.UpdateClassicAuthRequest{Username: &username}); err != nil {
		t.Fatal(err)
	}
	if !f.Committed || len(f.Logger.Entries) != 1 || f.Logger.Entries[0].Level != "info" || f.Logger.Entries[0].Message != "updated classic auth" {
		t.Fatalf("wrong credential update logs: %v", f.Logger.Entries)
	}
}

func TestProfileGrantReadsRemainQuietAndNormalizeEmptyResults(t *testing.T) {
	for _, grant := range []string{"roles", "permissions"} {
		for _, fail := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/fail=%t", grant, fail), func(t *testing.T) {
				f := testshared.NewFixture()
				failure := testshared.NotFound(domainmodelserror.Failure)
				f.SpaceMemberRole.ReadEffectiveRolesBySpaceMemberIdFn = func(context.Context, int64, int64) ([]domainmodelsentity.Role, error) {
					if fail {
						return nil, failure
					}
					return nil, nil
				}
				f.SpaceMemberRole.ReadEffectivePermissionsBySpaceMemberIdFn = func(context.Context, int64, int64) ([]domainmodelsentity.Permission, error) {
					if fail {
						return nil, failure
					}
					return nil, nil
				}
				u := applicationprofileaccount.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Validator, f.Transactor, f.Logger)
				var err error
				if grant == "roles" {
					var result []domainmodelsentity.Role
					result, err = u.GetRoles(context.Background(), f.Caller)
					if !fail && result == nil {
						t.Fatal("successful roles read returned nil slice")
					}
				} else {
					var result []domainmodelsentity.Permission
					result, err = u.GetPermissions(context.Background(), f.Caller)
					if !fail && result == nil {
						t.Fatal("successful permissions read returned nil slice")
					}
				}
				if fail {
					if err != failure || len(f.Logger.Entries) != 1 || f.Logger.Entries[0].Message != "failed to read profile "+grant || f.Logger.Entries[0].Meta["err"] != failure {
						t.Fatalf("incorrect grant failure logs: %v", f.Logger.Entries)
					}
				} else if err != nil || len(f.Logger.Entries) != 0 {
					t.Fatalf("successful read failed or logged: %v", err)
				}
			})
		}
	}
}
