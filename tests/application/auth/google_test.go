package auth_test

import (
	"context"
	"errors"
	"fmt"
	infrastructureutilityauthflow "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/auth_flow"
	"net/url"
	"strings"
	"testing"
	"time"

	applicationauthgoogle "github.com/MateWorkspace/lockmate/internal/application/auth/google"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
)

func google(f *testshared.Fixture, oauth *testshared.GoogleOAuth) domainusecasesauth.Google {
	read := f.Client.ReadByIdFn
	f.Client.ReadByIdFn = func(ctx context.Context, spaceId, id int64) (*domainmodelsentity.Client, error) {
		value, err := read(ctx, spaceId, id)
		if value != nil {
			redirect := "https://app.example.com/auth/callback"
			value.AuthRedirectUri = &redirect
		}
		return value, err
	}
	return applicationauthgoogle.NewUsecaseImpl(f.Client, f.Space, f.User, f.SpaceMember, f.SpaceMemberRole, f.Validator, f.Transactor, f.Logger, f.GoogleAuth, oauth, &launchStore{}, &handoffStore{fixture: f}, infrastructureutilityauthflow.NewRandomImpl(), applicationauthgoogle.Config{Enabled: true, PublicUrl: "https://lockmate-dev.dhonan.my.id", LaunchDuration: 2 * time.Minute, AttemptDuration: 10 * time.Minute, HandoffDuration: time.Minute})
}
func callbackRequest() domainusecasesauth.GoogleSignInCallbackRequest {
	return domainusecasesauth.GoogleSignInCallbackRequest{Code: "private-code", State: "private-state", Binding: "private-binding"}
}
func TestGoogleSeamlessRegistrationAndSync(t *testing.T) {
	for _, mode := range []string{"new", "existing", "conflict", "token failure", "commit failure", "provider failure", "subject conflict"} {
		t.Run(mode, func(t *testing.T) {
			f := testshared.NewFixture()
			calls := 0
			oauth := &testshared.GoogleOAuth{CompleteFn: func(ctx context.Context, args domaincontractsutility.CompleteGoogleOAuthArgs) (*domaincontractsutility.CompleteGoogleOAuthResult, error) {
				calls++
				if mode == "provider failure" {
					return nil, testshared.NotFound(domainmodelserror.GoogleOAuthStateInvalid)
				}
				return &domaincontractsutility.CompleteGoogleOAuthResult{ClientId: 30, SpaceId: 10, RedirectUri: "https://app.example.com/auth/callback", ClientState: "project-state", Identity: domaincontractsutility.GoogleOAuthIdentity{Subject: "google-sub", Email: "new@example.com", IsEmailVerified: true}}, nil
			}}
			if mode == "existing" {
				f.Googles[1] = &domainmodelsentity.GoogleAuth{Id: 50, UserId: 1, Subject: "google-sub", Email: "old@example.com"}
			}
			if mode == "token failure" {
				f.TokenErr = testshared.NotFound(domainmodelserror.TokenFailed)
			}
			if mode == "commit failure" {
				f.CommitErr = testshared.NotFound(domainmodelserror.Failure)
			}
			if mode == "conflict" {
				create := f.SpaceMember.CreateFn
				attempt := 0
				f.SpaceMember.CreateFn = func(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateSpaceMemberArgs) (int64, error) {
					attempt++
					if attempt == 1 {
						return 0, testshared.NotFound(domainmodelserror.SpaceMemberConflict)
					}
					return create(ctx, spaceId, args)
				}
			}
			if mode == "subject conflict" {
				read := f.GoogleAuth.ReadBySubjectFn
				count := 0
				f.GoogleAuth.CreateFn = func(context.Context, domaincontractsrepository.CreateGoogleAuthArgs) (int64, error) {
					return 0, testshared.NotFound(domainmodelserror.GoogleAuthSubjectConflict)
				}
				f.GoogleAuth.ReadBySubjectFn = func(ctx context.Context, subject string) (*domainmodelsentity.GoogleAuth, error) {
					count++
					if count == 2 {
						f.Googles[1] = &domainmodelsentity.GoogleAuth{Id: 50, UserId: 1, Subject: subject, Email: "new@example.com", IsEmailVerified: true}
					}
					return read(ctx, subject)
				}
			}
			result, err := google(f, oauth).SignInCallback(context.Background(), callbackRequest())
			if calls != 1 {
				t.Fatal("provider exchange repeated")
			}
			if mode == "token failure" || mode == "commit failure" || mode == "provider failure" {
				if err == nil || result != nil || len(f.Users) != 1 || len(f.Googles) != 0 {
					t.Fatal("failed callback retained writes or result")
				}
				expected := 1
				if mode == "token failure" {
					expected = 2
				}
				if len(f.Logger.Entries) != expected || f.Logger.Entries[expected-1].Level != "error" {
					t.Fatalf("failed callback logging: %v", f.Logger.Entries)
				}
				return
			}
			if err != nil || result == nil {
				t.Fatalf("callback: %v", err)
			}
			if mode == "existing" || mode == "subject conflict" {
				if f.Users[1].Name != "Current name" || f.Googles[1].Email != "new@example.com" || f.Classics[1].Username != "user" {
					t.Fatal("existing identity overwritten")
				}
			} else if f.Users[2] == nil || f.Users[2].Name != "Google user" || len(f.Classics) != 1 {
				t.Fatal("google registration not separate from classic")
			}
			if mode == "conflict" || mode == "subject conflict" {
				if f.Transactions != 2 || len(f.Logger.Entries) == 0 || f.Logger.Entries[0].Level != "warn" || f.Logger.Entries[1].Meta["next_attempt"] != 2 {
					t.Fatal("missing conflict retry or warning")
				}
			}
		})
	}
}
func TestGoogleInputAndDenial(t *testing.T) {
	f := testshared.NewFixture()
	oauth := &testshared.GoogleOAuth{}
	result, err := google(f, oauth).SignInCallback(context.Background(), domainusecasesauth.GoogleSignInCallbackRequest{})
	testshared.RequireError(t, err, domainmodelserror.BadArgs)
	if result != nil {
		t.Fatal("invalid result")
	}
	request := callbackRequest()
	request.Code = ""
	request.ProviderError = "access_denied"
	oauth.CompleteFn = func(ctx context.Context, args domaincontractsutility.CompleteGoogleOAuthArgs) (*domaincontractsutility.CompleteGoogleOAuthResult, error) {
		if args.ProviderError != "access_denied" {
			t.Fatal("denial lost")
		}
		return nil, testshared.NotFound(domainmodelserror.GoogleOAuthDenied)
	}
	_, err = google(f, oauth).SignInCallback(context.Background(), request)
	testshared.RequireError(t, err, domainmodelserror.GoogleOAuthDenied)
	if f.Transactions != 0 {
		t.Fatal("provider failure began transaction")
	}
}

func TestGoogleNameFallbackIsBestEffort(t *testing.T) {
	for _, supplied := range []bool{false, true} {
		t.Run(map[bool]string{false: "absent", true: "invalid"}[supplied], func(t *testing.T) {
			f := testshared.NewFixture()
			invalidName := strings.Repeat("private-provider-name", 100)
			identity := domaincontractsutility.GoogleOAuthIdentity{Subject: "google-sub", Email: "new@example.com", IsEmailVerified: true}
			if supplied {
				identity.Name = &invalidName
			}
			oauth := &testshared.GoogleOAuth{CompleteFn: func(context.Context, domaincontractsutility.CompleteGoogleOAuthArgs) (*domaincontractsutility.CompleteGoogleOAuthResult, error) {
				return &domaincontractsutility.CompleteGoogleOAuthResult{ClientId: 30, SpaceId: 10, RedirectUri: "https://app.example.com/auth/callback", ClientState: "project-state", Identity: identity}, nil
			}}
			result, err := google(f, oauth).SignInCallback(context.Background(), callbackRequest())
			if err != nil || result == nil || f.Users[2] == nil || f.Users[2].Name != "Google user" || !f.Committed {
				t.Fatalf("fallback registration failed: %v", err)
			}
			expected := 1
			if supplied {
				expected++
				if f.Logger.Entries[0].Level != "warn" || !errors.Is(f.Logger.Entries[0].Meta["err"].(error), domainmodelserror.UserNameTooLong) {
					t.Fatalf("missing fallback warning: %v", f.Logger.Entries)
				}
			}
			if len(f.Logger.Entries) != expected || f.Logger.Entries[expected-1].Level != "info" {
				t.Fatalf("wrong logs: %v", f.Logger.Entries)
			}
			if strings.Contains(fmt.Sprint(f.Logger.Entries), "private-provider-name") {
				t.Fatal("provider name leaked")
			}
		})
	}
}

// These doubles exercise application orchestration; PostgreSQL tests cover atomicity.
type launchStore struct {
	domaincontractsrepository.GoogleOAuthLaunch
	value *domainmodelsentity.GoogleOAuthLaunch
}

func (s *launchStore) Create(_ context.Context, value domainmodelsentity.GoogleOAuthLaunch) error {
	s.value = &value
	return nil
}
func (s *launchStore) Consume(_ context.Context, _ string) (*domainmodelsentity.GoogleOAuthLaunch, error) {
	return s.value, nil
}

type handoffStore struct {
	domaincontractsrepository.AuthHandoff
	fixture *testshared.Fixture
	value   *domainmodelsentity.AuthHandoff
}

func (s *handoffStore) Create(_ context.Context, value domainmodelsentity.AuthHandoff) error {
	if s.fixture.TokenErr != nil {
		return s.fixture.TokenErr
	}
	s.value = &value
	return nil
}
func TestGoogleCallbackRedirectHasNoTokens(t *testing.T) {
	f := testshared.NewFixture()
	oauth := &testshared.GoogleOAuth{CompleteFn: func(context.Context, domaincontractsutility.CompleteGoogleOAuthArgs) (*domaincontractsutility.CompleteGoogleOAuthResult, error) {
		return &domaincontractsutility.CompleteGoogleOAuthResult{ClientId: 30, SpaceId: 10, RedirectUri: "https://app.example.com/auth/callback", ClientState: "project-state", Identity: domaincontractsutility.GoogleOAuthIdentity{Subject: "sub", Email: "user@example.com", IsEmailVerified: true}}, nil
	}}
	result, err := google(f, oauth).SignInCallback(context.Background(), callbackRequest())
	if err != nil {
		t.Fatal(err)
	}
	uri, err := url.Parse(result.RedirectUrl)
	if err != nil || len(uri.Query().Get("code")) != 43 || uri.Query().Get("state") != "project-state" || len(uri.Query()) != 2 {
		t.Fatalf("unsafe callback redirect: %v", err)
	}
}
