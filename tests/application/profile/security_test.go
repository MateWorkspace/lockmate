package profile_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	applicationprofilesecurity "github.com/MateWorkspace/lockmate/internal/application/profile/security"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesprofile "github.com/MateWorkspace/lockmate/internal/domain/usecases/profile"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
)

func security(f *testshared.Fixture) domainusecasesprofile.Security {
	return applicationprofilesecurity.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.GoogleAuth, f.Password, f.Guard, f.Validator, f.Transactor, f.Logger)
}
func TestOwnPasswordChange(t *testing.T) {
	for _, mode := range []string{"success", "wrong old password", "commit failure", "no classic"} {
		t.Run(mode, func(t *testing.T) {
			f := testshared.NewFixture()
			request := domainusecasesprofile.ChangePasswordRequest{CurrentPassword: "password", NewPassword: "newpassword123"}
			if mode == "wrong old password" {
				request.CurrentPassword = "wrong"
			}
			if mode == "commit failure" {
				f.CommitErr = testshared.NotFound(domainmodelserror.Failure)
			}
			if mode == "no classic" {
				delete(f.Classics, 1)
			}
			err := security(f).ChangePassword(context.Background(), f.Caller, request)
			switch mode {
			case "wrong old password":
				testshared.RequireError(t, err, domainmodelserror.PasswordMismatch)
			case "commit failure":
				testshared.RequireError(t, err, domainmodelserror.Failure)
			case "no classic":
				testshared.RequireError(t, err, domainmodelserror.ClassicAuthNotFound)
			default:
				if err != nil || f.Classics[1].PasswordHash != "hash:newpassword123" || !f.Committed {
					t.Fatalf("password not changed: %v", err)
				}
			}
			if mode != "success" && f.Classics[1] != nil && f.Classics[1].PasswordHash != "hash:password" {
				t.Fatal("failure replaced password")
			}
			expected := 1
			if mode == "wrong old password" || mode == "no classic" {
				expected = 2
			}
			if len(f.Logger.Entries) != expected {
				t.Fatalf("wrong password logs: %v", f.Logger.Entries)
			}
			if expected == 2 && f.Logger.Entries[0].Message == f.Logger.Entries[1].Message {
				t.Fatal("step and transaction messages must differ")
			}
			level := "error"
			if mode == "success" {
				level = "info"
			}
			if f.Logger.Entries[0].Level != level {
				t.Fatal("wrong log level")
			}
		})
	}
}
func TestAuthSummaryIsSafeAndMethodsOptional(t *testing.T) {
	f := testshared.NewFixture()
	result, err := security(f).GetAuthMethods(context.Background(), f.Caller)
	if err != nil || result.ClassicAuth == nil || result.GoogleAuth != nil {
		t.Fatalf("summary: %v", err)
	}
	data, _ := json.Marshal(result)
	if strings.Contains(string(data), "password_hash") || strings.Contains(string(data), "hash:password") {
		t.Fatal("hash exposed")
	}
	delete(f.Classics, 1)
	result, err = security(f).GetAuthMethods(context.Background(), f.Caller)
	if err != nil || result.ClassicAuth != nil || result.GoogleAuth != nil {
		t.Fatal("missing methods became errors")
	}
	if len(f.Logger.Entries) != 0 {
		t.Fatal("summary read logged success")
	}
}
