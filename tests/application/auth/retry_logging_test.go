package auth_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
)

func TestExhaustedConflictRetriesLogWarnThenError(t *testing.T) {
	for _, flow := range []string{"classic membership", "google membership", "google subject"} {
		t.Run(flow, func(t *testing.T) {
			f := testshared.NewFixture()
			kind := domainmodelserror.SpaceMemberConflict
			if flow == "google subject" {
				kind = domainmodelserror.GoogleAuthSubjectConflict
			}
			failure := domainmodelserror.NewError("private-message", kind, errors.New("private-source"))
			f.SpaceMember.CreateFn = func(context.Context, int64, domaincontractsrepository.CreateSpaceMemberArgs) (int64, error) {
				return 0, failure
			}
			if flow == "google subject" {
				f.GoogleAuth.CreateFn = func(context.Context, domaincontractsrepository.CreateGoogleAuthArgs) (int64, error) {
					return 0, failure
				}
			}
			var err error
			if flow == "classic membership" {
				delete(f.Members, 20)
				_, err = session(f).SignIn(context.Background(), clientIdentity(), domainusecasesauth.SignInRequest{Username: "user", Password: "password"})
			} else {
				calls := 0
				oauth := &testshared.GoogleOAuth{CompleteFn: func(context.Context, domaincontractsutility.CompleteGoogleOAuthArgs) (*domaincontractsutility.CompleteGoogleOAuthResult, error) {
					calls++
					return &domaincontractsutility.CompleteGoogleOAuthResult{ClientId: 30, SpaceId: 10, RedirectUri: "https://app.example.com/auth/callback", ClientState: "project-state", Identity: domaincontractsutility.GoogleOAuthIdentity{Subject: "google-sub", Email: "new@example.com", IsEmailVerified: true}}, nil
				}}
				_, err = google(f, oauth).SignInCallback(context.Background(), callbackRequest())
				if calls != 1 {
					t.Fatal("provider exchange repeated")
				}
			}
			if err != failure || f.Transactions != 2 || f.Committed {
				t.Fatalf("incorrect retry outcome: %v", err)
			}
			if len(f.Logger.Entries) != 4 || f.Logger.Entries[0].Level != "warn" || f.Logger.Entries[1].Level != "warn" || f.Logger.Entries[2].Level != "error" || f.Logger.Entries[3].Level != "error" {
				t.Fatalf("wrong retry logs: %v", f.Logger.Entries)
			}
			if f.Logger.Entries[1].Meta["next_attempt"] != 2 || f.Logger.Entries[0].Meta["err"] != failure {
				t.Fatal("missing retry metadata")
			}
			if !strings.Contains(fmt.Sprint(f.Logger.Entries), "private-source") {
				t.Fatal("native error details missing")
			}
			if f.Logger.Entries[0].Message == f.Logger.Entries[1].Message || f.Logger.Entries[2].Message == f.Logger.Entries[3].Message {
				t.Fatal("failure and transaction messages are indistinguishable")
			}
		})
	}
}
