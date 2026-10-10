package auth_test

import (
	"context"
	"testing"

	applicationauthclient "github.com/MateWorkspace/lockmate/internal/application/auth/client"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
)

func TestClientAuthentication(t *testing.T) {
	for _, state := range []string{"active", "wrong secret", "missing", "inactive client", "inactive space", "outage"} {
		t.Run(state, func(t *testing.T) {
			f := testshared.NewFixture()
			request := domainusecasesauth.AuthenticateClientRequest{Slug: "web", Secret: "secret"}
			switch state {
			case "wrong secret":
				request.Secret = "wrong"
			case "missing":
				f.Client.ReadBySlugFn = func(context.Context, string) (*domainmodelsentity.Client, error) {
					return nil, testshared.NotFound(domainmodelserror.ClientNotFound)
				}
			case "inactive client":
				f.Client.ReadBySlugFn = func(context.Context, string) (*domainmodelsentity.Client, error) {
					return &domainmodelsentity.Client{Id: 30, SpaceId: 10, Slug: "web"}, nil
				}
			case "inactive space":
				f.Space.ReadByIdFn = func(context.Context, int64) (*domainmodelsentity.Space, error) {
					return &domainmodelsentity.Space{Id: 10}, nil
				}
			case "outage":
				f.Client.ReadBySlugFn = func(context.Context, string) (*domainmodelsentity.Client, error) {
					return nil, testshared.NotFound(domainmodelserror.Timeout)
				}
			}
			result, err := applicationauthclient.NewUsecaseImpl(f.Client, f.Space, f.Keygen, f.Logger).Authenticate(context.Background(), request)
			if state == "active" {
				if err != nil || result.Id != 30 || result.SpaceId != 10 || result.Slug != "web" || len(f.Logger.Entries) != 0 {
					t.Fatalf("client authentication: %v", err)
				}
				return
			}
			kind := domainmodelserror.Unauthorized
			if state == "outage" {
				kind = domainmodelserror.Timeout
			}
			testshared.RequireError(t, err, kind)
			if result != nil || len(f.Logger.Entries) != 1 {
				t.Fatal("failure result/log count")
			}
		})
	}
}
func TestGoogleInitiationBindsAuthenticatedScope(t *testing.T) {
	f := testshared.NewFixture()
	called := false
	oauth := &testshared.GoogleOAuth{BeginFn: func(ctx context.Context, args domaincontractsutility.BeginGoogleOAuthArgs) (string, error) {
		called = true
		if args.ClientId != 30 || args.SpaceId != 10 || args.Binding != "binding" {
			t.Fatal("wrong OAuth scope/binding")
		}
		return "https://accounts.google.com/authorization", nil
	}}
	result, err := google(f, oauth).BeginSignIn(context.Background(), clientIdentity(), domainusecasesauth.BeginGoogleSignInRequest{Binding: "binding"})
	if err != nil || !called || result.AuthorizationUrl == "" || len(f.Logger.Entries) != 0 {
		t.Fatalf("initiation: %v", err)
	}
	called = false
	_, err = google(f, oauth).BeginSignIn(context.Background(), nil, domainusecasesauth.BeginGoogleSignInRequest{Binding: "binding"})
	testshared.RequireError(t, err, domainmodelserror.Unauthorized)
	if called {
		t.Fatal("provider invoked without trusted client")
	}
}
