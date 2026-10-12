package presentationhttp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	applicationauthclient "github.com/MateWorkspace/lockmate/internal/application/auth/client"
	applicationauthsession "github.com/MateWorkspace/lockmate/internal/application/auth/session"
	applicationmanagementclient "github.com/MateWorkspace/lockmate/internal/application/management/client"
	applicationmiscipallowlist "github.com/MateWorkspace/lockmate/internal/application/misc/ip_allowlist"
	applicationmisctracer "github.com/MateWorkspace/lockmate/internal/application/misc/tracer"
	applicationprofileaccount "github.com/MateWorkspace/lockmate/internal/application/profile/account"
	applicationprofilesecurity "github.com/MateWorkspace/lockmate/internal/application/profile/security"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
	infrastructurerepositorygoogleauth "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/google_auth"
	infrastructurerepositorypermission "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/permission"
	infrastructurerepositoryrolepermission "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/role_permission"
	infrastructureutilitykeygen "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/keygen"
	infrastructureutilitypassword "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/password"
	infrastructureutilitypermissionguard "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/permission_guard"
	infrastructureutilitytoken "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/token"
	infrastructureutilitytracer "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/tracer"
	infrastructureutilityvalidator "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/validator"
	presentationhttphandlerauth "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/auth"
	presentationhttphandlermanagement "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/management"
	presentationhttphandlerprofile "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/profile"
	presentationhttphandlerversion "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/version"
	presentationhttpresponse "github.com/MateWorkspace/lockmate/internal/presentation/http/response"
	presentationhttproute "github.com/MateWorkspace/lockmate/internal/presentation/http/route"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
	"github.com/MateWorkspace/mate-golang/pgxdt"
	"github.com/gin-gonic/gin"
)

func TestPostgresHTTPAuthenticationAndIpAllowlist(t *testing.T) {
	fixture := testshared.NewPostgresFixture(t)
	ctx := context.Background()
	yes := true
	spaceId, err := fixture.Space.Create(ctx, domaincontractsrepository.CreateSpaceArgs{Slug: "lockmate", Name: "Lockmate"})
	if err != nil {
		t.Fatal(err)
	}
	roleId, err := fixture.Role.Create(ctx, spaceId, domaincontractsrepository.CreateRoleArgs{Slug: "member", Name: "Member", IsDefault: &yes})
	if err != nil {
		t.Fatal(err)
	}
	dt := pgxdt.NewPgxdt(fixture.Pool)
	question := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)
	dollar := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	permission := infrastructurerepositorypermission.NewPostgresImpl(dt, &question, &dollar)
	rolePermission := infrastructurerepositoryrolepermission.NewPostgresImpl(dt, &question, &dollar)
	googleAuth := infrastructurerepositorygoogleauth.NewPostgresImpl(dt, &question, &dollar)
	var profileGrantId int64
	for _, slug := range []string{"profile:get", "profile:set", "profile_security:get", "profile_security:set", "client:get", "client:add", "client_secret:set", "client_ip_allowlist:get", "client_ip_allowlist:add", "client_ip_allowlist:set", "client_ip_allowlist:remove"} {
		permissionId, err := permission.Create(ctx, spaceId, domaincontractsrepository.CreatePermissionArgs{Slug: slug, Name: slug})
		if err != nil {
			t.Fatal(err)
		}
		id, err := rolePermission.Create(ctx, spaceId, domaincontractsrepository.CreateRolePermissionArgs{RoleId: roleId, PermissionId: permissionId})
		if err != nil {
			t.Fatal(err)
		}
		if slug == "profile:get" {
			profileGrantId = id
		}
	}
	keygen := infrastructureutilitykeygen.NewGeneratorImpl()
	secret, hash, redacted, err := keygen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	clientId, err := fixture.Client.Create(ctx, spaceId, domaincontractsrepository.CreateClientArgs{Slug: "web", Name: "Web", SecretKeyHash: hash, SecretKeyRedacted: redacted})
	if err != nil {
		t.Fatal(err)
	}
	otherSecret, otherHash, otherRedacted, err := keygen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	otherClientId, err := fixture.Client.Create(ctx, spaceId, domaincontractsrepository.CreateClientArgs{Slug: "other-web", Name: "Other Web", SecretKeyHash: otherHash, SecretKeyRedacted: otherRedacted})
	if err != nil {
		t.Fatal(err)
	}
	// httptest requests originate from the documentation address 192.0.2.1.
	allowlistId, err := fixture.ClientIpAllowlist.Create(ctx, spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: clientId, Cidr: "192.0.2.1/32"})
	if err != nil {
		t.Fatal(err)
	}
	if otherClientId == clientId {
		t.Fatal("client identity collision")
	}
	validator := infrastructureutilityvalidator.NewValidatorImpl(true)
	password := infrastructureutilitypassword.NewBcryptImpl(4)
	guard := infrastructureutilitypermissionguard.NewClaimsImpl()
	token := infrastructureutilitytoken.NewJwtImpl("test-refresh-secret", 15*time.Minute, 24*time.Hour)
	tracer := infrastructureutilitytracer.NewContextImpl()
	logger := &testshared.Logger{}
	clientAuth := applicationauthclient.NewUsecaseImpl(fixture.Client, fixture.Space, keygen, logger)
	session := applicationauthsession.NewUsecaseImpl(fixture.Client, fixture.Space, fixture.User, fixture.ClassicAuth, fixture.SpaceMember, fixture.SpaceMemberRole, fixture.SigningKey, password, token, fixture.Secret, validator, fixture.Transactor, logger)
	account := applicationprofileaccount.NewUsecaseImpl(fixture.User, fixture.Space, fixture.SpaceMember, fixture.SpaceMemberRole, guard, validator, fixture.Transactor, logger)
	security := applicationprofilesecurity.NewUsecaseImpl(fixture.User, fixture.Space, fixture.SpaceMember, fixture.ClassicAuth, googleAuth, password, guard, validator, fixture.Transactor, logger)
	clientManagement := applicationmanagementclient.NewUsecaseImpl(fixture.User, fixture.Space, fixture.SpaceMember, fixture.Client, fixture.ClientIpAllowlist, guard, validator, logger, fixture.Transactor, keygen, "lockmate")
	r := &recorder{}
	router := gin.New()
	presentationhttproute.Route(router, presentationhttproute.Args{
		Auth:       presentationhttphandlerauth.NewHandler(session),
		Profile:    presentationhttphandlerprofile.NewHandler(account, security),
		Management: presentationhttphandlermanagement.NewHandler(&managementSpaceStub{r}, &managementPermissionStub{r}, &managementRoleStub{r}, &managementUserStub{r}, &managementClassicAuthStub{r}, &managementGoogleAuthStub{r}, clientManagement, &managementSpaceMemberStub{r}, &managementRolePermissionStub{r}, &managementSpaceMemberRoleStub{r}),
		Version:    presentationhttphandlerversion.NewHandler(domainmodelsapp.Info{Name: "lockmate", Version: "test"}),
		Client:     clientAuth, Session: session, Tracer: applicationmisctracer.NewUsecaseImpl(tracer, logger), IpAllowlist: applicationmiscipallowlist.NewUsecaseImpl(fixture.ClientIpAllowlist, logger),
	})
	access := ""
	makeRequest := func(method, path, body string) *http.Request {
		req := businessRequest(method, path, body)
		req.Header.Set("X-Client-Slug", "web")
		req.Header.Set("X-Client-Secret", secret)
		req.Header.Set("Authorization", "Bearer "+access)
		return req
	}
	response := perform(router, makeRequest("POST", "/api/v1/auth/register", `{"name":"Alex Doe","username":"alex","password":"SecurePass123!","email":"alex@example.com"}`))
	if response.Code != 201 {
		t.Fatalf("register: %d %s", response.Code, response.Body.String())
	}
	var registered presentationhttpresponse.TokenResponse
	if err := json.Unmarshal(response.Body.Bytes(), &registered); err != nil {
		t.Fatal(err)
	}
	if registered.AccessToken == "" || registered.RefreshToken == "" || !registered.IsNewUser {
		t.Fatal("registration did not issue a scoped session")
	}
	access = registered.AccessToken
	response = perform(router, makeRequest("GET", "/api/v1/profile", ""))
	if response.Code != 200 {
		t.Fatalf("profile: %d %s", response.Code, response.Body.String())
	}
	// A client without an allowlist entry is rejected before the handler.
	req := makeRequest("GET", "/api/v1/profile", "")
	req.Header.Set("X-Client-Slug", "other-web")
	req.Header.Set("X-Client-Secret", otherSecret)
	if response := perform(router, req); response.Code != 403 {
		t.Fatal("client without allowlist entry accepted")
	}
	response = perform(router, makeRequest("POST", "/api/v1/auth/sign-in", `{"username":"alex","password":"SecurePass123!"}`))
	if response.Code != 200 {
		t.Fatalf("sign-in: %s", response.Body.String())
	}
	response = perform(router, makeRequest("POST", "/api/v1/auth/refresh", fmt.Sprintf(`{"refresh_token":%q}`, registered.RefreshToken)))
	if response.Code != 200 {
		t.Fatalf("refresh: %s", response.Body.String())
	}
	// A signed token cannot preserve a grant deleted after it was issued.
	if err := rolePermission.DeleteById(ctx, spaceId, profileGrantId); err != nil {
		t.Fatal(err)
	}
	if response := perform(router, makeRequest("GET", "/api/v1/profile", "")); response.Code != 403 {
		t.Fatal("stale signed grant was trusted")
	}
	// IP allowlist mutation is guarded by independent permission slugs.
	allowlistPath := fmt.Sprintf("/api/v1/management/spaces/%d/clients/%d/ip-allowlist", spaceId, clientId)
	response = perform(router, makeRequest("POST", allowlistPath, `{"cidr":"203.0.113.0/24"}`))
	if response.Code != 201 {
		t.Fatalf("allowlist create: %d %s", response.Code, response.Body.String())
	}
	var created presentationhttpresponse.IdResponse
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	response = perform(router, makeRequest("PATCH", fmt.Sprintf("%s/%d", allowlistPath, created.Id), `{"is_active":false}`))
	if response.Code != 204 {
		t.Fatalf("allowlist update: %s", response.Body.String())
	}
	response = perform(router, makeRequest("DELETE", fmt.Sprintf("%s/%d", allowlistPath, created.Id), ""))
	if response.Code != 204 {
		t.Fatalf("allowlist delete: %s", response.Body.String())
	}
	// Committed secret rotation immediately rejects the previous secret.
	rotatePath := fmt.Sprintf("/api/v1/management/spaces/%d/clients/%d/secret", spaceId, clientId)
	response = perform(router, makeRequest("PATCH", rotatePath, ""))
	if response.Code != 200 {
		t.Fatalf("rotate: %s", response.Body.String())
	}
	var rotated presentationhttpresponse.ClientSecretResponse
	if err := json.Unmarshal(response.Body.Bytes(), &rotated); err != nil {
		t.Fatal(err)
	}
	if response := perform(router, makeRequest("POST", "/api/v1/auth/refresh", fmt.Sprintf(`{"refresh_token":%q}`, registered.RefreshToken))); response.Code != 401 {
		t.Fatal("old client secret remained valid")
	}
	secret = rotated.Secret
	response = perform(router, makeRequest("POST", "/api/v1/auth/refresh", fmt.Sprintf(`{"refresh_token":%q}`, registered.RefreshToken)))
	if response.Code != 200 {
		t.Fatalf("new client secret rejected: %s", response.Body.String())
	}
	// Deactivating the caller's allowlist entry rejects the next request.
	no := false
	if err := fixture.ClientIpAllowlist.UpdateById(ctx, spaceId, allowlistId, domaincontractsrepository.UpdateClientIpAllowlistArgs{IsActive: &no}); err != nil {
		t.Fatal(err)
	}
	if response := perform(router, makeRequest("POST", "/api/v1/auth/refresh", fmt.Sprintf(`{"refresh_token":%q}`, registered.RefreshToken))); response.Code != 403 {
		t.Fatal("revoked allowlist entry remained valid")
	}
}
