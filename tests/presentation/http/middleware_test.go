package presentationhttp_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
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
	router.Use(presentationhttpmiddleware.Recovery(), presentationhttpmiddleware.Trace(fixture.trace))
	router.GET("/protected", presentationhttpmiddleware.Client(fixture.client), presentationhttpmiddleware.IPAllowlist(fixture.ipAllowlist, nil, "X-Forwarded-For"), presentationhttpmiddleware.Auth(session), func(c *gin.Context) {
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
func TestIPAllowlist(t *testing.T) {
	fixture := newHTTPFixture()
	response := perform(fixture.router, businessRequest("GET", "/api/v1/profile", ""))
	if response.Code != 200 || fixture.ipAllowlist.calls != 1 || len(fixture.ipAllowlist.ips) != 1 || fixture.ipAllowlist.ips[0] == "" {
		t.Fatalf("allowed request rejected: %d %s", response.Code, response.Body.String())
	}
	handled := len(fixture.recorder.calls)
	fixture.ipAllowlist.allowed = false
	response = perform(fixture.router, businessRequest("GET", "/api/v1/profile", ""))
	if response.Code != 403 || len(fixture.recorder.calls) != handled {
		t.Fatal("disallowed IP reached the handler")
	}
	fixture.ipAllowlist.allowed = true
	fixture.ipAllowlist.failure = domainmodelserror.NewError("database failed", domainmodelserror.Failure, errors.New("private database details"))
	response = perform(fixture.router, businessRequest("GET", "/api/v1/profile", ""))
	if response.Code != 500 || strings.Contains(response.Body.String(), "private database") || len(fixture.recorder.calls) != handled {
		t.Fatal("allowlist lookup failure did not fail closed")
	}
}

func TestIPAllowlistForwardedHeader(t *testing.T) {
	for _, item := range []struct {
		name, peer, header, want string
		trusted                  []netip.Prefix
	}{
		{"trusted forwarded list", "127.0.0.1:1234", "203.0.113.9, 127.0.0.2", "203.0.113.9", []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")}},
		{"trusted single value", "127.0.0.1:1234", "203.0.113.9", "203.0.113.9", []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")}},
		{"untrusted peer ignores header", "198.18.0.1:1234", "203.0.113.9", "198.18.0.1", []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")}},
		{"trusted without header", "127.0.0.1:1234", "", "127.0.0.1", []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")}},
		{"no trusted proxy", "198.18.0.1:1234", "203.0.113.9", "198.18.0.1", nil},
	} {
		t.Run(item.name, func(t *testing.T) {
			captured := &ipAllowlistStub{allowed: true}
			router := gin.New()
			router.Use(func(c *gin.Context) {
				ctx := presentationhttputils.InjectClientIdentity(c.Request.Context(), &domainmodelsentity.ClientIdentity{Id: 1, SpaceId: 2, Slug: "web"})
				c.Request = c.Request.WithContext(ctx)
				c.Next()
			})
			router.Use(presentationhttpmiddleware.IPAllowlist(captured, item.trusted, "X-Forwarded-For"))
			router.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })
			request := httptest.NewRequest("GET", "/x", nil)
			request.RemoteAddr = item.peer
			if item.header != "" {
				request.Header.Set("X-Forwarded-For", item.header)
			}
			response := perform(router, request)
			if response.Code != 200 || len(captured.ips) != 1 || captured.ips[0] != item.want {
				t.Fatalf("resolved %v, want %s", captured.ips, item.want)
			}
		})
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
