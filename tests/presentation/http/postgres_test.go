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
	applicationmiscallowedorigin "github.com/MateWorkspace/lockmate/internal/application/misc/allowed_origin"
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

func TestPostgresHTTPAuthenticationAndDynamicOrigins(t *testing.T) {
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
	for _, slug := range []string{"profile:get", "profile:set", "profile_security:get", "profile_security:set", "client:get", "client:add", "client_secret:set", "client_origin:get", "client_origin:add", "client_origin:set", "client_origin:remove"} {
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
	originId, err := fixture.ClientOrigin.Create(ctx, spaceId, domaincontractsrepository.CreateClientOriginArgs{ClientId: clientId, Origin: "https://app.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if otherClientId == clientId {
		t.Fatal("client identity collision")
	}
	validator := infrastructureutilityvalidator.NewValidatorImpl()
	password := infrastructureutilitypassword.NewBcryptImpl(4)
	guard := infrastructureutilitypermissionguard.NewClaimsImpl()
	token := infrastructureutilitytoken.NewJwtImpl("test-access-secret", "test-refresh-secret", 15*time.Minute, 24*time.Hour)
	tracer := infrastructureutilitytracer.NewContextImpl()
	logger := &testshared.Logger{}
	clientAuth := applicationauthclient.NewUsecaseImpl(fixture.Client, fixture.Space, keygen, logger)
	session := applicationauthsession.NewUsecaseImpl(fixture.Client, fixture.Space, fixture.User, fixture.ClassicAuth, fixture.SpaceMember, fixture.SpaceMemberRole, password, token, validator, fixture.Transactor, logger)
	account := applicationprofileaccount.NewUsecaseImpl(fixture.User, fixture.Space, fixture.SpaceMember, fixture.SpaceMemberRole, guard, validator, fixture.Transactor, logger)
	security := applicationprofilesecurity.NewUsecaseImpl(fixture.User, fixture.Space, fixture.SpaceMember, fixture.ClassicAuth, googleAuth, password, guard, validator, fixture.Transactor, logger)
	clientManagement := applicationmanagementclient.NewUsecaseImpl(fixture.User, fixture.Space, fixture.SpaceMember, fixture.Client, fixture.ClientOrigin, guard, validator, logger, fixture.Transactor, keygen, "lockmate")
	r := &recorder{}
	router := gin.New()
	presentationhttproute.Route(router, presentationhttproute.Args{
		Auth:       presentationhttphandlerauth.NewHandler(session),
		Profile:    presentationhttphandlerprofile.NewHandler(account, security),
		Management: presentationhttphandlermanagement.NewHandler(&managementSpaceStub{r}, &managementPermissionStub{r}, &managementRoleStub{r}, &managementUserStub{r}, &managementClassicAuthStub{r}, &managementGoogleAuthStub{r}, clientManagement, &managementSpaceMemberStub{r}, &managementRolePermissionStub{r}, &managementSpaceMemberRoleStub{r}),
		Version:    presentationhttphandlerversion.NewHandler(domainmodelsapp.Info{Name: "lockmate", Version: "test"}),
		Client:     clientAuth, Session: session, Tracer: applicationmisctracer.NewUsecaseImpl(tracer, logger), AllowedOrigin: applicationmiscallowedorigin.NewUsecaseImpl(fixture.ClientOrigin, validator, logger),
	})
	access := ""
	makeRequest := func(method, path, body string) *http.Request {
		req := businessRequest(method, path, body)
		req.Header.Set("X-Client-Slug", "web")
		req.Header.Set("X-Client-Secret", secret)
		req.Header.Set("Origin", "https://app.example.com")
		req.Header.Set("Authorization", "Bearer "+access)
		return req
	}
	response := perform(router, makeRequest("POST", "/api/v1/auth/register", `{"name":"Alex Doe","username":"alex","password":"SecurePass123!","email":"alex@example.com"}`))
	if response.Code != 201 {
		t.Fatalf("register: %d %s", response.Code, response.Body.String())
	}
	var registered presentationhttpresponse.SessionResponse
	if err := json.Unmarshal(response.Body.Bytes(), &registered); err != nil {
		t.Fatal(err)
	}
	if registered.AccessToken == "" || registered.RefreshToken == "" || !registered.IsNewUser || registered.Space.Id != spaceId {
		t.Fatal("registration did not issue a scoped session")
	}
	access = registered.AccessToken
	response = perform(router, makeRequest("GET", "/api/v1/profile", ""))
	if response.Code != 200 {
		t.Fatalf("profile: %d %s", response.Code, response.Body.String())
	}
	req := makeRequest("GET", "/api/v1/profile", "")
	req.Header.Set("X-Client-Slug", "other-web")
	req.Header.Set("X-Client-Secret", otherSecret)
	if response := perform(router, req); response.Code != 403 {
		t.Fatal("another client's origin accepted")
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
	// Client-origin mutation is guarded by independent permission slugs.
	originsPath := fmt.Sprintf("/api/v1/management/spaces/%d/clients/%d/origins", spaceId, clientId)
	response = perform(router, makeRequest("POST", originsPath, `{"origin":"https://preview.example.com"}`))
	if response.Code != 201 {
		t.Fatalf("origin create: %d %s", response.Code, response.Body.String())
	}
	var created presentationhttpresponse.IdResponse
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	response = perform(router, makeRequest("PATCH", fmt.Sprintf("%s/%d", originsPath, created.Id), `{"is_active":false}`))
	if response.Code != 204 {
		t.Fatalf("origin update: %s", response.Body.String())
	}
	response = perform(router, makeRequest("DELETE", fmt.Sprintf("%s/%d", originsPath, created.Id), ""))
	if response.Code != 204 {
		t.Fatalf("origin delete: %s", response.Body.String())
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
	// Persisted origin changes take effect on the next actual request.
	no := false
	if err := fixture.ClientOrigin.UpdateById(ctx, spaceId, originId, domaincontractsrepository.UpdateClientOriginArgs{IsActive: &no}); err != nil {
		t.Fatal(err)
	}
	if response := perform(router, makeRequest("POST", "/api/v1/auth/refresh", fmt.Sprintf(`{"refresh_token":%q}`, registered.RefreshToken))); response.Code != 403 {
		t.Fatal("revoked origin remained valid")
	}
}
