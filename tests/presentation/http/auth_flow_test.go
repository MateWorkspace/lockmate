package presentationhttp_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	presentationhttphandlerauth "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/auth"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
)

type googleFlowStub struct {
	start    func(context.Context, *domainmodelsentity.ClientIdentity, domainusecasesauth.StartGoogleSignInRequest) (*domainusecasesauth.StartGoogleSignInResult, error)
	launch   func(context.Context, string) (*domainusecasesauth.LaunchGoogleSignInResult, error)
	callback func(context.Context, domainusecasesauth.GoogleSignInCallbackRequest) (*domainusecasesauth.GoogleSignInCallbackResult, error)
}

func (s *googleFlowStub) Start(ctx context.Context, client *domainmodelsentity.ClientIdentity, request domainusecasesauth.StartGoogleSignInRequest) (*domainusecasesauth.StartGoogleSignInResult, error) {
	return s.start(ctx, client, request)
}
func (s *googleFlowStub) Launch(ctx context.Context, ticket string) (*domainusecasesauth.LaunchGoogleSignInResult, error) {
	return s.launch(ctx, ticket)
}
func (s *googleFlowStub) SignInCallback(ctx context.Context, request domainusecasesauth.GoogleSignInCallbackRequest) (*domainusecasesauth.GoogleSignInCallbackResult, error) {
	return s.callback(ctx, request)
}

type handoffHTTPStub struct {
	exchange func(context.Context, *domainmodelsentity.ClientIdentity, domainusecasesauth.ExchangeAuthHandoffRequest) (*domainusecasesauth.SessionResult, error)
}

func (s *handoffHTTPStub) Exchange(ctx context.Context, client *domainmodelsentity.ClientIdentity, request domainusecasesauth.ExchangeAuthHandoffRequest) (*domainusecasesauth.SessionResult, error) {
	return s.exchange(ctx, client, request)
}

func TestGoogleHTTPBrowserFlow(t *testing.T) {
	state := strings.Repeat("s", 43)
	binding := strings.Repeat("b", 43)
	called := map[string]int{}
	google := &googleFlowStub{
		start: func(ctx context.Context, client *domainmodelsentity.ClientIdentity, request domainusecasesauth.StartGoogleSignInRequest) (*domainusecasesauth.StartGoogleSignInResult, error) {
			called["start"]++
			if client.Id != 1 || client.SpaceId != 2 || request.State != "project-state" || ctx.Value(traceMarker{}) == nil {
				t.Fatal("untrusted initiation")
			}
			return &domainusecasesauth.StartGoogleSignInResult{LaunchUrl: "https://lockmate-dev.dhonan.my.id/api/v1/auth/google/launch?ticket=ticket"}, nil
		},
		launch: func(ctx context.Context, ticket string) (*domainusecasesauth.LaunchGoogleSignInResult, error) {
			called["launch"]++
			if ticket != "ticket" {
				t.Fatal("launch ticket lost")
			}
			return &domainusecasesauth.LaunchGoogleSignInResult{AuthorizationUrl: "https://accounts.google.com/authorize", State: state, Binding: binding, CookieDuration: 600}, nil
		},
		callback: func(ctx context.Context, request domainusecasesauth.GoogleSignInCallbackRequest) (*domainusecasesauth.GoogleSignInCallbackResult, error) {
			called["callback"]++
			if request.Binding != binding || request.State != state || request.Code != "provider-code" {
				t.Fatal("callback scope/binding lost")
			}
			return &domainusecasesauth.GoogleSignInCallbackResult{RedirectUrl: "https://app.example.com/callback?code=handoff&state=project-state"}, nil
		},
	}
	handoff := &handoffHTTPStub{exchange: func(ctx context.Context, client *domainmodelsentity.ClientIdentity, request domainusecasesauth.ExchangeAuthHandoffRequest) (*domainusecasesauth.SessionResult, error) {
		called["exchange"]++
		if client.Id != 1 || client.SpaceId != 2 || request.Code != "handoff" {
			t.Fatal("exchange lost trusted client")
		}
		return &domainusecasesauth.SessionResult{AccessToken: "access", RefreshToken: "refresh"}, nil
	}}
	f := newHTTPFixtureWithFlow(presentationhttphandlerauth.NewFlowHandler(google, handoff))
	started := perform(f.router, businessRequest("POST", "/api/v1/auth/google/start", `{"state":"project-state","redirect_uri":"https://attacker.example.com"}`))
	if started.Code != 200 || started.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("start: %d %s", started.Code, started.Body.String())
	}
	launch := perform(f.router, httptest.NewRequest("GET", "/api/v1/auth/google/launch?ticket=ticket", nil))
	if launch.Code != 303 || launch.Header().Get("Location") != "https://accounts.google.com/authorize" || launch.Header().Get("Referrer-Policy") != "no-referrer" {
		t.Fatal("launch redirect")
	}
	cookies := launch.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].Secure || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteLaxMode || cookies[0].Domain != "" || cookies[0].Path != "/" || cookies[0].MaxAge != 600 {
		t.Fatal("binding cookie attributes")
	}
	callback := httptest.NewRequest("GET", "/api/v1/auth/google/callback?state="+state+"&code=provider-code", nil)
	callback.AddCookie(cookies[0])
	completed := perform(f.router, callback)
	if completed.Code != 303 || !strings.HasPrefix(completed.Header().Get("Location"), "https://app.example.com/callback?") || completed.Result().Cookies()[0].MaxAge != -1 {
		t.Fatal("callback redirect/cleanup")
	}
	exchanged := perform(f.router, businessRequest("POST", "/api/v1/auth/exchange", `{"code":"handoff"}`))
	if exchanged.Code != 200 || exchanged.Header().Get("Cache-Control") != "no-store" || !strings.Contains(exchanged.Body.String(), `"access_token":"access"`) {
		t.Fatal("exchange response")
	}
	for _, operation := range []string{"start", "launch", "callback", "exchange"} {
		if called[operation] != 1 {
			t.Fatalf("%s not dispatched", operation)
		}
	}
}

func TestGoogleHTTPRejectsAmbiguousBrowserInputs(t *testing.T) {
	state := strings.Repeat("s", 43)
	name := presentationhttputils.GoogleBindingCookieName(state)
	for _, test := range []struct {
		path    string
		cookies int
	}{
		{"/api/v1/auth/google/launch?ticket=a&ticket=b", 0},
		{"/api/v1/auth/google/callback?state=" + state + "&code=c", 0},
		{"/api/v1/auth/google/callback?state=" + state + "&state=" + state + "&code=c", 1},
		{"/api/v1/auth/google/callback?state=" + state + "&code=c&code=d", 1},
		{"/api/v1/auth/google/callback?state=" + state + "&code=c", 2},
	} {
		t.Run(test.path+string(rune('0'+test.cookies)), func(t *testing.T) {
			f := newHTTPFixtureWithFlow(presentationhttphandlerauth.NewFlowHandler(&googleFlowStub{}, nil))
			request := httptest.NewRequest("GET", test.path, nil)
			for i := 0; i < test.cookies; i++ {
				request.AddCookie(&http.Cookie{Name: name, Value: strings.Repeat("b", 43)})
			}
			response := perform(f.router, request)
			if response.Code != 400 && response.Code != 401 {
				t.Fatalf("ambiguous input accepted: %d", response.Code)
			}
		})
	}
	if name == presentationhttputils.GoogleBindingCookieName("other-state") {
		t.Fatal("concurrent cookies collide")
	}
}

func TestGoogleHTTPDisabledAndSafeErrors(t *testing.T) {
	f := newHTTPFixture()
	for _, request := range []*http.Request{businessRequest("POST", "/api/v1/auth/google/start", `{}`), businessRequest("POST", "/api/v1/auth/exchange", `{}`), httptest.NewRequest("GET", "/api/v1/auth/google/launch", nil), httptest.NewRequest("GET", "/api/v1/auth/google/callback", nil)} {
		if response := perform(f.router, request); response.Code != 501 {
			t.Fatalf("disabled status %d", response.Code)
		}
	}
	google := &googleFlowStub{start: func(context.Context, *domainmodelsentity.ClientIdentity, domainusecasesauth.StartGoogleSignInRequest) (*domainusecasesauth.StartGoogleSignInResult, error) {
		return nil, domainmodelserror.NewError("safe", domainmodelserror.GoogleOAuthFailed, errors.New("private-provider-secret"))
	}}
	f = newHTTPFixtureWithFlow(presentationhttphandlerauth.NewFlowHandler(google, nil))
	response := perform(f.router, businessRequest("POST", "/api/v1/auth/google/start", `{"state":"x"}`))
	if response.Code != 500 || strings.Contains(response.Body.String(), "private-provider-secret") {
		t.Fatal("provider source exposed")
	}
	if response := perform(f.router, httptest.NewRequest("POST", "/api/v1/auth/google/start", strings.NewReader(`{}`))); response.Code != 401 {
		t.Fatal("start bypassed client authentication")
	}
}
