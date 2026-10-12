package auth_test

import (
	"context"
	"strings"
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
	oauth := &testshared.GoogleOAuth{BeginFn: func(ctx context.Context, args domaincontractsutility.BeginGoogleOAuthArgs) (*domaincontractsutility.BeginGoogleOAuthResult, error) {
		called = true
		if args.ClientId != 30 || args.SpaceId != 10 || len(args.Binding) != 43 || args.RedirectUri != "https://app.example.com/auth/callback" {
			t.Fatal("wrong OAuth scope/binding")
		}
		return &domaincontractsutility.BeginGoogleOAuthResult{AuthorizationUrl: "https://accounts.google.com/authorization", State: "provider-state"}, nil
	}}
	usecase := google(f, oauth)
	result, err := usecase.Start(context.Background(), clientIdentity(), domainusecasesauth.StartGoogleSignInRequest{State: "0123456789abcdef0123456789abcdef"})
	if err != nil || called || result.LaunchUrl == "" {
		t.Fatalf("initiation: %v", err)
	}
	ticket := strings.Split(result.LaunchUrl, "ticket=")[1]
	launch, err := usecase.Launch(context.Background(), ticket)
	if err != nil || !called || launch.Binding == "" {
		t.Fatalf("launch: %v", err)
	}
	_, err = usecase.Start(context.Background(), nil, domainusecasesauth.StartGoogleSignInRequest{State: "0123456789abcdef0123456789abcdef"})
	testshared.RequireError(t, err, domainmodelserror.Unauthorized)
}
