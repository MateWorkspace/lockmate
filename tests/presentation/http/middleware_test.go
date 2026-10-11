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
	presentationhttpmiddleware "github.com/MateWorkspace/lockmate/internal/presentation/http/middleware"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-gonic/gin"
)

type accessAuthStub struct {
	domainusecasesauth.Session
	claims  *domainmodelsentity.AccessClaims
	failure error
	calls   int
	token   string
}

func (s *accessAuthStub) AuthenticateAccessToken(ctx context.Context, client *domainmodelsentity.ClientIdentity, token string) (*domainmodelsentity.AccessClaims, error) {
	s.calls++
	s.token = token
	if client == nil || client.Id != 1 || ctx.Value(traceMarker{}) == nil {
		panic("untrusted middleware context")
	}
	return s.claims, s.failure
}
func middlewareRouter(fixture *httpFixture, session *accessAuthStub) *gin.Engine {
	router := gin.New()
	router.Use(presentationhttpmiddleware.Recovery(), presentationhttpmiddleware.Trace(fixture.trace), presentationhttpmiddleware.CORS(fixture.origin))
	router.GET("/protected", presentationhttpmiddleware.Client(fixture.client, fixture.origin), presentationhttpmiddleware.Auth(session), func(c *gin.Context) {
		claims := presentationhttputils.AccessClaims(c.Request.Context())
		c.JSON(http.StatusOK, claims)
	})
	return router
}
func TestClientHeaders(t *testing.T) {
	for _, item := range []struct {
		name   string
		modify func(*http.Request)
		status int
	}{
		{"missing slug", func(r *http.Request) { r.Header.Del("X-Client-Slug") }, 401},
		{"missing secret", func(r *http.Request) { r.Header.Del("X-Client-Secret") }, 401},
		{"missing execution", func(r *http.Request) { r.Header.Del("X-Client-Exec-Id") }, 400},
		{"invalid execution", func(r *http.Request) { r.Header.Set("X-Client-Exec-Id", "bad") }, 400},
		{"zero execution", func(r *http.Request) { r.Header.Set("X-Client-Exec-Id", "00000000-0000-0000-0000-000000000000") }, 400},
		{"duplicate slug", func(r *http.Request) { r.Header.Add("X-Client-Slug", "other") }, 400},
		{"duplicate secret", func(r *http.Request) { r.Header.Add("X-Client-Secret", "other") }, 400},
		{"duplicate execution", func(r *http.Request) { r.Header.Add("X-Client-Exec-Id", executionId) }, 400},
	} {
		t.Run(item.name, func(t *testing.T) {
			fixture := newHTTPFixture()
			req := businessRequest("GET", "/api/v1/profile", "")
			item.modify(req)
			response := perform(fixture.router, req)
			if response.Code != item.status || fixture.client.calls != 0 || len(fixture.recorder.calls) != 0 {
				t.Fatalf("%d %s", response.Code, response.Body.String())
			}
		})
	}
	fixture := newHTTPFixture()
	response := perform(fixture.router, businessRequest("GET", "/api/v1/profile", ""))
	if response.Code != 200 || fixture.client.request.Secret != "  secret stays unchanged  " {
		t.Fatal("client credential was changed")
	}
	fixture.client.failure = domainmodelserror.NewError("authentication failed", domainmodelserror.Unauthorized, nil)
	response = perform(fixture.router, businessRequest("GET", "/api/v1/profile", ""))
	if response.Code != 401 {
		t.Fatal("client rejection ignored")
	}
}
func TestAccessAuthentication(t *testing.T) {
	fixture := newHTTPFixture()
	session := &accessAuthStub{claims: &domainmodelsentity.AccessClaims{UserId: 3, SpaceId: 2, SpaceMemberId: 4, Permissions: []string{"current:get"}}}
	router := middlewareRouter(fixture, session)
	for _, header := range []string{"", "raw-token", "Basic token", "Bearer", "Bearer token extra"} {
		req := businessRequest("GET", "/protected", "")
		req.Header.Set("Authorization", header)
		response := perform(router, req)
		if response.Code != 401 || session.calls != 0 {
			t.Fatalf("bad authorization accepted: %s", header)
		}
	}
	req := businessRequest("GET", "/protected", "")
	req.Header.Add("Authorization", "Bearer another")
	if response := perform(router, req); response.Code != 401 {
		t.Fatal("duplicate authorization accepted")
	}
	req = businessRequest("GET", "/protected", "")
	req.Header.Set("Authorization", "bEaReR trusted-token")
	response := perform(router, req)
	if response.Code != 200 || session.token != "trusted-token" || !strings.Contains(response.Body.String(), "current:get") {
		t.Fatal("session grants were not loaded")
	}
	for _, sentinel := range []error{domainmodelserror.TokenExpired, domainmodelserror.TokenInvalid, domainmodelserror.Unauthorized, domainmodelserror.Timeout, domainmodelserror.Failure} {
		session.failure = domainmodelserror.NewError("diagnostic", sentinel, errors.New("private source"))
		response = perform(router, businessRequest("GET", "/protected", ""))
		expected := 401
		if sentinel == domainmodelserror.Timeout {
			expected = 504
		}
		if sentinel == domainmodelserror.Failure {
			expected = 500
		}
		if response.Code != expected || strings.Contains(response.Body.String(), "private source") {
			t.Fatal("token error mapping failed")
		}
	}
	session.failure = nil
	session.claims.SpaceId = 99
	if response := perform(router, businessRequest("GET", "/protected", "")); response.Code != 500 {
		t.Fatal("invalid usecase identity accepted")
	}
}
func TestDynamicCORS(t *testing.T) {
	fixture := newHTTPFixture()
	preflight := func(origin string) *http.Request {
		req := httptest.NewRequest("OPTIONS", "/api/v1/profile", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Access-Control-Request-Method", "PATCH")
		req.Header.Set("Access-Control-Request-Headers", "x-client-slug,x-client-secret,x-client-exec-id,authorization,content-type")
		return req
	}
	response := perform(fixture.router, preflight("https://client.example.com"))
	if response.Code != 204 || fixture.client.calls != 0 || fixture.origin.scopedCalls != 0 {
		t.Fatal("preflight authenticated")
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "https://client.example.com" || response.Header().Get("Access-Control-Allow-Credentials") != "" || response.Header().Get("Access-Control-Max-Age") != "0" {
		t.Fatalf("bad CORS headers: %+v", response.Header())
	}
	headers := strings.ToLower(response.Header().Get("Access-Control-Allow-Headers"))
	for _, name := range []string{"x-client-slug", "x-client-secret", "x-client-exec-id", "authorization"} {
		if !strings.Contains(headers, name) {
			t.Fatal("missing allowed header")
		}
	}
	fixture.origin.global = false
	response = perform(fixture.router, preflight("https://client.example.com"))
	if response.Code != 403 || response.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("revoked origin accepted")
	}
	fixture.origin.global = true
	fixture.origin.scoped = false
	req := businessRequest("GET", "/api/v1/profile", "")
	req.Header.Set("Origin", "https://client.example.com")
	response = perform(fixture.router, req)
	if response.Code != 403 || fixture.origin.scopedCalls != 1 || len(fixture.recorder.calls) != 0 {
		t.Fatal("other client's origin accepted")
	}
	fixture.origin.scoped = true
	response = perform(fixture.router, req)
	if response.Code != 200 || fixture.origin.scopedCalls != 2 {
		t.Fatal("client origin was not rechecked")
	}
	fixture.origin.failure = domainmodelserror.NewError("database failed", domainmodelserror.Failure, errors.New("private database details"))
	response = perform(fixture.router, preflight("https://client.example.com"))
	if response.Code != 500 || strings.Contains(response.Body.String(), "private database") || response.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("origin lookup failure did not fail closed")
	}
	fixture.origin.failure = nil
	req = preflight("https://client.example.com")
	req.Header.Add("Origin", "https://other.example.com")
	if response := perform(fixture.router, req); response.Code != 403 {
		t.Fatal("ambiguous origin accepted")
	}
	// Even matching request Host must pass the persisted allowlist.
	fixture.origin.global = false
	req = preflight("http://example.com")
	if response := perform(fixture.router, req); response.Code != 403 {
		t.Fatal("same-host origin bypassed lookup")
	}
	fixture.origin.global = true
	if response := perform(fixture.router, req); response.Code != 204 {
		t.Fatal("same-host preflight not completed")
	}
	fixture.origin.globalCalls = 0
	fixture.origin.scopedCalls = 0
	if response := perform(fixture.router, businessRequest("GET", "/api/v1/profile", "")); response.Code != 200 || fixture.origin.globalCalls != 0 || fixture.origin.scopedCalls != 0 {
		t.Fatal("non-browser request checked origin")
	}
}
func TestTraceAndRecovery(t *testing.T) {
	fixture := newHTTPFixture()
	fixture.trace.failure = domainmodelserror.NewError("failed", domainmodelserror.TraceEventIdFailed, errors.New("private entropy details"))
	response := perform(fixture.router, businessRequest("GET", "/api/v1/profile", ""))
	if response.Code != 500 || fixture.client.calls != 0 || strings.Contains(response.Body.String(), "entropy") {
		t.Fatal("trace failure continued")
	}
	router := gin.New()
	router.Use(presentationhttpmiddleware.Recovery())
	router.GET("/panic", func(c *gin.Context) { panic("secret must stay private") })
	response = perform(router, httptest.NewRequest("GET", "/panic", nil))
	if response.Code != 500 || strings.Contains(response.Body.String(), "secret") {
		t.Fatal("panic was not sanitized")
	}
}
