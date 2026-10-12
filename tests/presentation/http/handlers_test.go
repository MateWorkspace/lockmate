package presentationhttp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	domainusecasesprofile "github.com/MateWorkspace/lockmate/internal/domain/usecases/profile"
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
	presentationhttphandlerauth "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/auth"
	presentationhttphandlermanagement "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/management"
	presentationhttphandlerprofile "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/profile"
	presentationhttphandlerversion "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/version"
	presentationhttproute "github.com/MateWorkspace/lockmate/internal/presentation/http/route"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() { gin.SetMode(gin.TestMode) }

type invocation struct {
	name string
	args []any
	ctx  context.Context
}
type recorder struct {
	calls     []invocation
	failure   error
	nilResult bool
}

func (r *recorder) record(ctx context.Context, name string, args ...any) {
	r.calls = append(r.calls, invocation{name: name, args: args, ctx: ctx})
}

type authSessionStub struct{ recorder *recorder }

func (s *authSessionStub) SignIn(ctx context.Context, client *domainmodelsentity.ClientIdentity, request domainusecasesauth.SignInRequest) (*domainusecasesauth.TokenResult, error) {
	s.recorder.record(ctx, "auth.Session.SignIn", client, request)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainusecasesauth.TokenResult{AccessToken: "access-once", RefreshToken: "refresh-once"}, nil
}
func (s *authSessionStub) Register(ctx context.Context, client *domainmodelsentity.ClientIdentity, request domainusecasesauth.RegisterRequest) (*domainusecasesauth.TokenResult, error) {
	s.recorder.record(ctx, "auth.Session.Register", client, request)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainusecasesauth.TokenResult{AccessToken: "access-once", RefreshToken: "refresh-once"}, nil
}
func (s *authSessionStub) Refresh(ctx context.Context, client *domainmodelsentity.ClientIdentity, request domainusecasesauth.RefreshRequest) (*domainusecasesauth.TokenResult, error) {
	s.recorder.record(ctx, "auth.Session.Refresh", client, request)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainusecasesauth.TokenResult{AccessToken: "access-once", RefreshToken: "refresh-once"}, nil
}
func (s *authSessionStub) AuthenticateAccessToken(ctx context.Context, client *domainmodelsentity.ClientIdentity, accessToken string) (*domainmodelsentity.AccessClaims, error) {
	return &domainmodelsentity.AccessClaims{UserId: 3, SpaceId: 2, SpaceMemberId: 4, Permissions: []string{"fresh:get"}}, nil
}

type managementClassicAuthStub struct{ recorder *recorder }

func (s *managementClassicAuthStub) Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesmanagement.CreateClassicAuthRequest) (id int64, err error) {
	s.recorder.record(ctx, "management.ClassicAuth.Create", caller, request)
	if s.recorder.failure != nil {
		return 0, s.recorder.failure
	}
	return 11, nil
}
func (s *managementClassicAuthStub) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) (*domainusecasesshared.ClassicAuthView, error) {
	s.recorder.record(ctx, "management.ClassicAuth.ReadById", caller, id)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainusecasesshared.ClassicAuthView{}, nil
}
func (s *managementClassicAuthStub) ReadByUserId(ctx context.Context, caller *domainmodelsentity.AccessClaims, value int64) (*domainusecasesshared.ClassicAuthView, error) {
	s.recorder.record(ctx, "management.ClassicAuth.ReadByUserId", caller, value)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainusecasesshared.ClassicAuthView{}, nil
}
func (s *managementClassicAuthStub) ReadByUsername(ctx context.Context, caller *domainmodelsentity.AccessClaims, value string) (*domainusecasesshared.ClassicAuthView, error) {
	s.recorder.record(ctx, "management.ClassicAuth.ReadByUsername", caller, value)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainusecasesshared.ClassicAuthView{}, nil
}
func (s *managementClassicAuthStub) ReadByEmail(ctx context.Context, caller *domainmodelsentity.AccessClaims, value string) (*domainusecasesshared.ClassicAuthView, error) {
	s.recorder.record(ctx, "management.ClassicAuth.ReadByEmail", caller, value)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainusecasesshared.ClassicAuthView{}, nil
}
func (s *managementClassicAuthStub) ReadByPhone(ctx context.Context, caller *domainmodelsentity.AccessClaims, value string) (*domainusecasesshared.ClassicAuthView, error) {
	s.recorder.record(ctx, "management.ClassicAuth.ReadByPhone", caller, value)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainusecasesshared.ClassicAuthView{}, nil
}
func (s *managementClassicAuthStub) ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesmanagement.FilterClassicAuthRequest) (domainusecasesshared.Page[domainusecasesshared.ClassicAuthView], error) {
	s.recorder.record(ctx, "management.ClassicAuth.ReadByFilter", caller, request)
	return domainusecasesshared.Page[domainusecasesshared.ClassicAuthView]{Total: 19}, s.recorder.failure
}
func (s *managementClassicAuthStub) UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64, request domainusecasesmanagement.UpdateClassicAuthRequest) error {
	s.recorder.record(ctx, "management.ClassicAuth.UpdateById", caller, id, request)
	return s.recorder.failure
}
func (s *managementClassicAuthStub) ResetPassword(ctx context.Context, caller *domainmodelsentity.AccessClaims, userId int64, request domainusecasesmanagement.ResetUserPasswordRequest) error {
	s.recorder.record(ctx, "management.ClassicAuth.ResetPassword", caller, userId, request)
	return s.recorder.failure
}
func (s *managementClassicAuthStub) DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) error {
	s.recorder.record(ctx, "management.ClassicAuth.DeleteById", caller, id)
	return s.recorder.failure
}

type managementClientStub struct{ recorder *recorder }

func (s *managementClientStub) Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.CreateClientRequest) (*domainusecasesmanagement.ClientSecretResult, error) {
	s.recorder.record(ctx, "management.Client.Create", caller, spaceId, request)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainusecasesmanagement.ClientSecretResult{Secret: "secret-once"}, nil
}
func (s *managementClientStub) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) (*domainusecasesshared.ClientView, error) {
	s.recorder.record(ctx, "management.Client.ReadById", caller, spaceId, id)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainusecasesshared.ClientView{}, nil
}
func (s *managementClientStub) ReadBySlug(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, value string) (*domainusecasesshared.ClientView, error) {
	s.recorder.record(ctx, "management.Client.ReadBySlug", caller, spaceId, value)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainusecasesshared.ClientView{}, nil
}
func (s *managementClientStub) ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.FilterClientRequest) (domainusecasesshared.Page[domainusecasesshared.ClientView], error) {
	s.recorder.record(ctx, "management.Client.ReadByFilter", caller, spaceId, request)
	return domainusecasesshared.Page[domainusecasesshared.ClientView]{Total: 19}, s.recorder.failure
}
func (s *managementClientStub) UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64, request domainusecasesmanagement.UpdateClientRequest) error {
	s.recorder.record(ctx, "management.Client.UpdateById", caller, spaceId, id, request)
	return s.recorder.failure
}
func (s *managementClientStub) RotateSecret(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) (*domainusecasesmanagement.ClientSecretResult, error) {
	s.recorder.record(ctx, "management.Client.RotateSecret", caller, spaceId, id)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainusecasesmanagement.ClientSecretResult{Secret: "secret-once"}, nil
}
func (s *managementClientStub) DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) error {
	s.recorder.record(ctx, "management.Client.DeleteById", caller, spaceId, id)
	return s.recorder.failure
}
func (s *managementClientStub) CreateIpAllowlist(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64, request domainusecasesmanagement.CreateClientIpAllowlistRequest) (int64, error) {
	s.recorder.record(ctx, "management.Client.CreateIpAllowlist", caller, spaceId, clientId, request)
	if s.recorder.failure != nil {
		return 0, s.recorder.failure
	}
	return 11, nil
}
func (s *managementClientStub) ReadIpAllowlistById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId, id int64) (*domainmodelsentity.ClientIpAllowlist, error) {
	s.recorder.record(ctx, "management.Client.ReadIpAllowlistById", caller, spaceId, clientId, id)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainmodelsentity.ClientIpAllowlist{}, nil
}
func (s *managementClientStub) ReadIpAllowlistByCidr(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64, cidr string) (*domainmodelsentity.ClientIpAllowlist, error) {
	s.recorder.record(ctx, "management.Client.ReadIpAllowlistByCidr", caller, spaceId, clientId, cidr)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainmodelsentity.ClientIpAllowlist{}, nil
}
func (s *managementClientStub) ReadIpAllowlistsByClientId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64) ([]domainmodelsentity.ClientIpAllowlist, error) {
	s.recorder.record(ctx, "management.Client.ReadIpAllowlistsByClientId", caller, spaceId, clientId)
	return nil, s.recorder.failure
}
func (s *managementClientStub) ReadIpAllowlistsByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64, request domainusecasesmanagement.FilterClientIpAllowlistRequest) (domainusecasesshared.Page[domainmodelsentity.ClientIpAllowlist], error) {
	s.recorder.record(ctx, "management.Client.ReadIpAllowlistsByFilter", caller, spaceId, clientId, request)
	return domainusecasesshared.Page[domainmodelsentity.ClientIpAllowlist]{Total: 19}, s.recorder.failure
}
func (s *managementClientStub) UpdateIpAllowlistById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId, id int64, request domainusecasesmanagement.UpdateClientIpAllowlistRequest) error {
	s.recorder.record(ctx, "management.Client.UpdateIpAllowlistById", caller, spaceId, clientId, id, request)
	return s.recorder.failure
}
func (s *managementClientStub) DeleteIpAllowlistById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId, id int64) error {
	s.recorder.record(ctx, "management.Client.DeleteIpAllowlistById", caller, spaceId, clientId, id)
	return s.recorder.failure
}

type managementGoogleAuthStub struct{ recorder *recorder }

func (s *managementGoogleAuthStub) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) (*domainmodelsentity.GoogleAuth, error) {
	s.recorder.record(ctx, "management.GoogleAuth.ReadById", caller, id)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainmodelsentity.GoogleAuth{}, nil
}
func (s *managementGoogleAuthStub) ReadByUserId(ctx context.Context, caller *domainmodelsentity.AccessClaims, value int64) (*domainmodelsentity.GoogleAuth, error) {
	s.recorder.record(ctx, "management.GoogleAuth.ReadByUserId", caller, value)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainmodelsentity.GoogleAuth{}, nil
}
func (s *managementGoogleAuthStub) ReadBySubject(ctx context.Context, caller *domainmodelsentity.AccessClaims, value string) (*domainmodelsentity.GoogleAuth, error) {
	s.recorder.record(ctx, "management.GoogleAuth.ReadBySubject", caller, value)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainmodelsentity.GoogleAuth{}, nil
}
func (s *managementGoogleAuthStub) ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesmanagement.FilterGoogleAuthRequest) (domainusecasesshared.Page[domainmodelsentity.GoogleAuth], error) {
	s.recorder.record(ctx, "management.GoogleAuth.ReadByFilter", caller, request)
	return domainusecasesshared.Page[domainmodelsentity.GoogleAuth]{Total: 19}, s.recorder.failure
}

type managementPermissionStub struct{ recorder *recorder }

func (s *managementPermissionStub) Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.CreatePermissionRequest) (id int64, err error) {
	s.recorder.record(ctx, "management.Permission.Create", caller, spaceId, request)
	if s.recorder.failure != nil {
		return 0, s.recorder.failure
	}
	return 11, nil
}
func (s *managementPermissionStub) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) (*domainmodelsentity.Permission, error) {
	s.recorder.record(ctx, "management.Permission.ReadById", caller, spaceId, id)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainmodelsentity.Permission{}, nil
}
func (s *managementPermissionStub) ReadBySlug(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, value string) (*domainmodelsentity.Permission, error) {
	s.recorder.record(ctx, "management.Permission.ReadBySlug", caller, spaceId, value)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainmodelsentity.Permission{}, nil
}
func (s *managementPermissionStub) ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.FilterPermissionRequest) (domainusecasesshared.Page[domainmodelsentity.Permission], error) {
	s.recorder.record(ctx, "management.Permission.ReadByFilter", caller, spaceId, request)
	return domainusecasesshared.Page[domainmodelsentity.Permission]{Total: 19}, s.recorder.failure
}
func (s *managementPermissionStub) UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64, request domainusecasesmanagement.UpdatePermissionRequest) error {
	s.recorder.record(ctx, "management.Permission.UpdateById", caller, spaceId, id, request)
	return s.recorder.failure
}
func (s *managementPermissionStub) DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) error {
	s.recorder.record(ctx, "management.Permission.DeleteById", caller, spaceId, id)
	return s.recorder.failure
}

type managementRoleStub struct{ recorder *recorder }

func (s *managementRoleStub) Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.CreateRoleRequest) (id int64, err error) {
	s.recorder.record(ctx, "management.Role.Create", caller, spaceId, request)
	if s.recorder.failure != nil {
		return 0, s.recorder.failure
	}
	return 11, nil
}
func (s *managementRoleStub) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) (*domainmodelsentity.Role, error) {
	s.recorder.record(ctx, "management.Role.ReadById", caller, spaceId, id)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainmodelsentity.Role{}, nil
}
func (s *managementRoleStub) ReadBySlug(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, value string) (*domainmodelsentity.Role, error) {
	s.recorder.record(ctx, "management.Role.ReadBySlug", caller, spaceId, value)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainmodelsentity.Role{}, nil
}
func (s *managementRoleStub) ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.FilterRoleRequest) (domainusecasesshared.Page[domainmodelsentity.Role], error) {
	s.recorder.record(ctx, "management.Role.ReadByFilter", caller, spaceId, request)
	return domainusecasesshared.Page[domainmodelsentity.Role]{Total: 19}, s.recorder.failure
}
func (s *managementRoleStub) ReadDefault(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64) (*domainmodelsentity.Role, error) {
	s.recorder.record(ctx, "management.Role.ReadDefault", caller, spaceId)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainmodelsentity.Role{}, nil
}
func (s *managementRoleStub) UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64, request domainusecasesmanagement.UpdateRoleRequest) error {
	s.recorder.record(ctx, "management.Role.UpdateById", caller, spaceId, id, request)
	return s.recorder.failure
}
func (s *managementRoleStub) SetDefault(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) error {
	s.recorder.record(ctx, "management.Role.SetDefault", caller, spaceId, id)
	return s.recorder.failure
}
func (s *managementRoleStub) DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) error {
	s.recorder.record(ctx, "management.Role.DeleteById", caller, spaceId, id)
	return s.recorder.failure
}

type managementRolePermissionStub struct{ recorder *recorder }

func (s *managementRolePermissionStub) Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.CreateRolePermissionRequest) (id int64, err error) {
	s.recorder.record(ctx, "management.RolePermission.Create", caller, spaceId, request)
	if s.recorder.failure != nil {
		return 0, s.recorder.failure
	}
	return 11, nil
}
func (s *managementRolePermissionStub) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) (*domaincontractsrepository.RolePermissionDetails, error) {
	s.recorder.record(ctx, "management.RolePermission.ReadById", caller, spaceId, id)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domaincontractsrepository.RolePermissionDetails{}, nil
}
func (s *managementRolePermissionStub) ReadByRoleIdAndPermissionId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, roleId, permissionId int64) (*domaincontractsrepository.RolePermissionDetails, error) {
	s.recorder.record(ctx, "management.RolePermission.ReadByRoleIdAndPermissionId", caller, spaceId, roleId, permissionId)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domaincontractsrepository.RolePermissionDetails{}, nil
}
func (s *managementRolePermissionStub) ReadByRoleId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, value int64) ([]domaincontractsrepository.RolePermissionWithPermission, error) {
	s.recorder.record(ctx, "management.RolePermission.ReadByRoleId", caller, spaceId, value)
	return nil, s.recorder.failure
}
func (s *managementRolePermissionStub) ReadByPermissionId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, value int64) ([]domaincontractsrepository.RolePermissionWithRole, error) {
	s.recorder.record(ctx, "management.RolePermission.ReadByPermissionId", caller, spaceId, value)
	return nil, s.recorder.failure
}
func (s *managementRolePermissionStub) DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) error {
	s.recorder.record(ctx, "management.RolePermission.DeleteById", caller, spaceId, id)
	return s.recorder.failure
}
func (s *managementRolePermissionStub) DeleteByRoleIdOrPermissionId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.DeleteRolePermissionRequest) error {
	s.recorder.record(ctx, "management.RolePermission.DeleteByRoleIdOrPermissionId", caller, spaceId, request)
	return s.recorder.failure
}

type managementSpaceStub struct{ recorder *recorder }

func (s *managementSpaceStub) Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesmanagement.CreateSpaceRequest) (id int64, err error) {
	s.recorder.record(ctx, "management.Space.Create", caller, request)
	if s.recorder.failure != nil {
		return 0, s.recorder.failure
	}
	return 11, nil
}
func (s *managementSpaceStub) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) (*domainmodelsentity.Space, error) {
	s.recorder.record(ctx, "management.Space.ReadById", caller, id)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainmodelsentity.Space{}, nil
}
func (s *managementSpaceStub) ReadBySlug(ctx context.Context, caller *domainmodelsentity.AccessClaims, value string) (*domainmodelsentity.Space, error) {
	s.recorder.record(ctx, "management.Space.ReadBySlug", caller, value)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainmodelsentity.Space{}, nil
}
func (s *managementSpaceStub) ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesmanagement.FilterSpaceRequest) (domainusecasesshared.Page[domainmodelsentity.Space], error) {
	s.recorder.record(ctx, "management.Space.ReadByFilter", caller, request)
	return domainusecasesshared.Page[domainmodelsentity.Space]{Total: 19}, s.recorder.failure
}
func (s *managementSpaceStub) UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64, request domainusecasesmanagement.UpdateSpaceRequest) error {
	s.recorder.record(ctx, "management.Space.UpdateById", caller, id, request)
	return s.recorder.failure
}
func (s *managementSpaceStub) DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) error {
	s.recorder.record(ctx, "management.Space.DeleteById", caller, id)
	return s.recorder.failure
}

type managementSpaceMemberStub struct{ recorder *recorder }

func (s *managementSpaceMemberStub) Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.CreateSpaceMemberRequest) (id int64, err error) {
	s.recorder.record(ctx, "management.SpaceMember.Create", caller, spaceId, request)
	if s.recorder.failure != nil {
		return 0, s.recorder.failure
	}
	return 11, nil
}
func (s *managementSpaceMemberStub) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) (*domainmodelsentity.SpaceMember, error) {
	s.recorder.record(ctx, "management.SpaceMember.ReadById", caller, spaceId, id)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainmodelsentity.SpaceMember{}, nil
}
func (s *managementSpaceMemberStub) ReadByUserId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, value int64) (*domainmodelsentity.SpaceMember, error) {
	s.recorder.record(ctx, "management.SpaceMember.ReadByUserId", caller, spaceId, value)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainmodelsentity.SpaceMember{}, nil
}
func (s *managementSpaceMemberStub) ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.FilterSpaceMemberRequest) (domainusecasesshared.Page[domaincontractsrepository.SpaceMemberWithUser], error) {
	s.recorder.record(ctx, "management.SpaceMember.ReadByFilter", caller, spaceId, request)
	return domainusecasesshared.Page[domaincontractsrepository.SpaceMemberWithUser]{Total: 19}, s.recorder.failure
}
func (s *managementSpaceMemberStub) UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64, request domainusecasesmanagement.UpdateSpaceMemberRequest) error {
	s.recorder.record(ctx, "management.SpaceMember.UpdateById", caller, spaceId, id, request)
	return s.recorder.failure
}
func (s *managementSpaceMemberStub) ReadEffectiveRolesBySpaceMemberId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, spaceMemberId int64) ([]domainmodelsentity.Role, error) {
	s.recorder.record(ctx, "management.SpaceMember.ReadEffectiveRolesBySpaceMemberId", caller, spaceId, spaceMemberId)
	return nil, s.recorder.failure
}
func (s *managementSpaceMemberStub) ReadEffectivePermissionsBySpaceMemberId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, spaceMemberId int64) ([]domainmodelsentity.Permission, error) {
	s.recorder.record(ctx, "management.SpaceMember.ReadEffectivePermissionsBySpaceMemberId", caller, spaceId, spaceMemberId)
	return nil, s.recorder.failure
}
func (s *managementSpaceMemberStub) DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) error {
	s.recorder.record(ctx, "management.SpaceMember.DeleteById", caller, spaceId, id)
	return s.recorder.failure
}

type managementSpaceMemberRoleStub struct{ recorder *recorder }

func (s *managementSpaceMemberRoleStub) Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.CreateSpaceMemberRoleRequest) (id int64, err error) {
	s.recorder.record(ctx, "management.SpaceMemberRole.Create", caller, spaceId, request)
	if s.recorder.failure != nil {
		return 0, s.recorder.failure
	}
	return 11, nil
}
func (s *managementSpaceMemberRoleStub) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) (*domaincontractsrepository.SpaceMemberRoleDetails, error) {
	s.recorder.record(ctx, "management.SpaceMemberRole.ReadById", caller, spaceId, id)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domaincontractsrepository.SpaceMemberRoleDetails{}, nil
}
func (s *managementSpaceMemberRoleStub) ReadBySpaceMemberIdAndRoleId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, spaceMemberId, roleId int64) (*domaincontractsrepository.SpaceMemberRoleDetails, error) {
	s.recorder.record(ctx, "management.SpaceMemberRole.ReadBySpaceMemberIdAndRoleId", caller, spaceId, spaceMemberId, roleId)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domaincontractsrepository.SpaceMemberRoleDetails{}, nil
}
func (s *managementSpaceMemberRoleStub) ReadBySpaceMemberId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, value int64) ([]domaincontractsrepository.SpaceMemberRoleWithRole, error) {
	s.recorder.record(ctx, "management.SpaceMemberRole.ReadBySpaceMemberId", caller, spaceId, value)
	return nil, s.recorder.failure
}
func (s *managementSpaceMemberRoleStub) ReadByRoleId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, value int64) ([]domaincontractsrepository.SpaceMemberRoleWithSpaceMember, error) {
	s.recorder.record(ctx, "management.SpaceMemberRole.ReadByRoleId", caller, spaceId, value)
	return nil, s.recorder.failure
}
func (s *managementSpaceMemberRoleStub) DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) error {
	s.recorder.record(ctx, "management.SpaceMemberRole.DeleteById", caller, spaceId, id)
	return s.recorder.failure
}
func (s *managementSpaceMemberRoleStub) DeleteBySpaceMemberIdOrRoleId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request domainusecasesmanagement.DeleteSpaceMemberRoleRequest) error {
	s.recorder.record(ctx, "management.SpaceMemberRole.DeleteBySpaceMemberIdOrRoleId", caller, spaceId, request)
	return s.recorder.failure
}

type managementUserStub struct{ recorder *recorder }

func (s *managementUserStub) Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesmanagement.CreateUserRequest) (id int64, err error) {
	s.recorder.record(ctx, "management.User.Create", caller, request)
	if s.recorder.failure != nil {
		return 0, s.recorder.failure
	}
	return 11, nil
}
func (s *managementUserStub) ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) (*domainmodelsentity.User, error) {
	s.recorder.record(ctx, "management.User.ReadById", caller, id)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainmodelsentity.User{}, nil
}
func (s *managementUserStub) ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesmanagement.FilterUserRequest) (domainusecasesshared.Page[domainmodelsentity.User], error) {
	s.recorder.record(ctx, "management.User.ReadByFilter", caller, request)
	return domainusecasesshared.Page[domainmodelsentity.User]{Total: 19}, s.recorder.failure
}
func (s *managementUserStub) UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64, request domainusecasesmanagement.UpdateUserRequest) error {
	s.recorder.record(ctx, "management.User.UpdateById", caller, id, request)
	return s.recorder.failure
}
func (s *managementUserStub) DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) error {
	s.recorder.record(ctx, "management.User.DeleteById", caller, id)
	return s.recorder.failure
}

type profileAccountStub struct{ recorder *recorder }

func (s *profileAccountStub) GetProfile(ctx context.Context, caller *domainmodelsentity.AccessClaims) (*domainusecasesprofile.ProfileResult, error) {
	s.recorder.record(ctx, "profile.Account.GetProfile", caller)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainusecasesprofile.ProfileResult{}, nil
}
func (s *profileAccountStub) GetRoles(ctx context.Context, caller *domainmodelsentity.AccessClaims) ([]domainmodelsentity.Role, error) {
	s.recorder.record(ctx, "profile.Account.GetRoles", caller)
	return nil, s.recorder.failure
}
func (s *profileAccountStub) GetPermissions(ctx context.Context, caller *domainmodelsentity.AccessClaims) ([]domainmodelsentity.Permission, error) {
	s.recorder.record(ctx, "profile.Account.GetPermissions", caller)
	return nil, s.recorder.failure
}
func (s *profileAccountStub) UpdateProfile(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesprofile.UpdateProfileRequest) error {
	s.recorder.record(ctx, "profile.Account.UpdateProfile", caller, request)
	return s.recorder.failure
}
func (s *profileAccountStub) UpdateMembershipPreferences(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesprofile.UpdateMembershipPreferencesRequest) error {
	s.recorder.record(ctx, "profile.Account.UpdateMembershipPreferences", caller, request)
	return s.recorder.failure
}

type profileSecurityStub struct{ recorder *recorder }

func (s *profileSecurityStub) GetAuthMethods(ctx context.Context, caller *domainmodelsentity.AccessClaims) (*domainusecasesprofile.AuthMethodsResult, error) {
	s.recorder.record(ctx, "profile.Security.GetAuthMethods", caller)
	if s.recorder.failure != nil {
		return nil, s.recorder.failure
	}
	if s.recorder.nilResult {
		return nil, nil
	}
	return &domainusecasesprofile.AuthMethodsResult{}, nil
}
func (s *profileSecurityStub) UpdateClassicAuth(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesprofile.UpdateClassicAuthRequest) error {
	s.recorder.record(ctx, "profile.Security.UpdateClassicAuth", caller, request)
	return s.recorder.failure
}
func (s *profileSecurityStub) ChangePassword(ctx context.Context, caller *domainmodelsentity.AccessClaims, request domainusecasesprofile.ChangePasswordRequest) error {
	s.recorder.record(ctx, "profile.Security.ChangePassword", caller, request)
	return s.recorder.failure
}

type clientAuthStub struct {
	request domainusecasesauth.AuthenticateClientRequest
	failure error
	calls   int
}

func (s *clientAuthStub) Authenticate(ctx context.Context, request domainusecasesauth.AuthenticateClientRequest) (*domainmodelsentity.ClientIdentity, error) {
	s.calls++
	s.request = request
	if s.failure != nil {
		return nil, s.failure
	}
	return &domainmodelsentity.ClientIdentity{Id: 1, SpaceId: 2, Slug: request.Slug}, nil
}

type ipAllowlistStub struct {
	allowed bool
	failure error
	calls   int
	ips     []string
}

func (s *ipAllowlistStub) IsAllowedIpForClient(ctx context.Context, spaceId, clientId int64, ip string) (bool, error) {
	s.calls++
	s.ips = append(s.ips, ip)
	if spaceId != 2 || clientId != 1 {
		panic("untrusted allowlist scope")
	}
	return s.allowed, s.failure
}

type traceStub struct {
	failure error
	events  int
}
type traceMarker struct{}

func (s *traceStub) Inject(ctx context.Context, slug string, execId uuid.UUID) (context.Context, error) {
	s.events++
	if s.failure != nil {
		return nil, s.failure
	}
	return context.WithValue(ctx, traceMarker{}, s.events), nil
}

type httpFixture struct {
	router      *gin.Engine
	recorder    *recorder
	client      *clientAuthStub
	ipAllowlist *ipAllowlistStub
	trace       *traceStub
}

func newHTTPFixture() *httpFixture {
	return newHTTPFixtureWithFlow(nil)
}

func newHTTPFixtureWithFlow(flow presentationhttproute.AuthFlowHandler) *httpFixture {
	r := &recorder{}
	fixture := &httpFixture{recorder: r, client: &clientAuthStub{}, ipAllowlist: &ipAllowlistStub{allowed: true}, trace: &traceStub{}}
	fixture.router = gin.New()
	presentationhttproute.Route(fixture.router, presentationhttproute.Args{
		Flow:    flow,
		Auth:    presentationhttphandlerauth.NewHandler(&authSessionStub{r}),
		Jwks:    presentationhttphandlerauth.NewJwksHandler(jwksStub{}),
		Profile: presentationhttphandlerprofile.NewHandler(&profileAccountStub{r}, &profileSecurityStub{r}),
		Management: presentationhttphandlermanagement.NewHandler(
			&managementSpaceStub{r}, &managementPermissionStub{r}, &managementRoleStub{r}, &managementUserStub{r},
			&managementClassicAuthStub{r}, &managementGoogleAuthStub{r}, &managementClientStub{r}, &managementSpaceMemberStub{r},
			&managementRolePermissionStub{r}, &managementSpaceMemberRoleStub{r}),
		Version: presentationhttphandlerversion.NewHandler(domainmodelsapp.Info{Name: "lockmate", Version: "test"}),
		Client:  fixture.client, Session: &authSessionStub{r}, Tracer: fixture.trace,
		IpAllowlist: fixture.ipAllowlist, ClientIpHeader: "X-Forwarded-For",
	})
	return fixture
}

const executionId = "60984d94-12ce-4d38-bd60-e01f25994119"

func businessRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-Slug", "trusted-client")
	req.Header.Set("X-Client-Secret", "  secret stays unchanged  ")
	req.Header.Set("X-Client-Exec-Id", executionId)
	req.Header.Set("Authorization", "Bearer token")
	return req
}
func perform(router http.Handler, req *http.Request) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

type operationCase struct {
	method    string
	path      string
	request   string
	operation string
	status    int
}

var operationCases = []operationCase{
	{"POST", "/api/v1/management/spaces", `{"slug": "value-slug", "name": "value-name", "description": "value-description", "is_active": false, "preferences": {"theme": "dark"}}`, "management.Space.Create", 201},
	{"GET", "/api/v1/management/spaces/7", `{}`, "management.Space.ReadById", 200},
	{"GET", "/api/v1/management/spaces/by-slug/lookup-slug", `{}`, "management.Space.ReadBySlug", 200},
	{"GET", "/api/v1/management/spaces?page=3&limit=25&search=value-search&is_active=false", `{}`, "management.Space.ReadByFilter", 200},
	{"PATCH", "/api/v1/management/spaces/7", `{"name": "value-name", "description": "value-description", "is_active": false, "preferences": {"theme": "dark"}}`, "management.Space.UpdateById", 204},
	{"DELETE", "/api/v1/management/spaces/7", `{}`, "management.Space.DeleteById", 204},
	{"POST", "/api/v1/management/spaces/7/permissions", `{"slug": "value-slug", "name": "value-name", "description": "value-description", "preferences": {"theme": "dark"}}`, "management.Permission.Create", 201},
	{"GET", "/api/v1/management/spaces/7/permissions/19", `{}`, "management.Permission.ReadById", 200},
	{"GET", "/api/v1/management/spaces/7/permissions/by-slug/lookup-slug", `{}`, "management.Permission.ReadBySlug", 200},
	{"GET", "/api/v1/management/spaces/7/permissions?page=3&limit=25&search=value-search", `{}`, "management.Permission.ReadByFilter", 200},
	{"PATCH", "/api/v1/management/spaces/7/permissions/19", `{"name": "value-name", "description": "value-description", "preferences": {"theme": "dark"}}`, "management.Permission.UpdateById", 204},
	{"DELETE", "/api/v1/management/spaces/7/permissions/19", `{}`, "management.Permission.DeleteById", 204},
	{"POST", "/api/v1/management/spaces/7/roles", `{"slug": "value-slug", "name": "value-name", "description": "value-description", "is_default": false, "preferences": {"theme": "dark"}}`, "management.Role.Create", 201},
	{"GET", "/api/v1/management/spaces/7/roles/17", `{}`, "management.Role.ReadById", 200},
	{"GET", "/api/v1/management/spaces/7/roles/by-slug/lookup-slug", `{}`, "management.Role.ReadBySlug", 200},
	{"GET", "/api/v1/management/spaces/7/roles?page=3&limit=25&search=value-search&is_default=false", `{}`, "management.Role.ReadByFilter", 200},
	{"GET", "/api/v1/management/spaces/7/roles/default", `{}`, "management.Role.ReadDefault", 200},
	{"PATCH", "/api/v1/management/spaces/7/roles/17", `{"name": "value-name", "description": "value-description", "is_default": false, "preferences": {"theme": "dark"}}`, "management.Role.UpdateById", 204},
	{"PATCH", "/api/v1/management/spaces/7/roles/17/default", `{}`, "management.Role.SetDefault", 204},
	{"DELETE", "/api/v1/management/spaces/7/roles/17", `{}`, "management.Role.DeleteById", 204},
	{"POST", "/api/v1/management/users", `{"name": "value-name", "bio": "value-bio", "avatar_path": "value-avatar_path", "preferences": {"theme": "dark"}, "username": "value-username", "password": "value-password", "email": "value-email", "phone": "value-phone", "is_email_verified": false, "is_phone_verified": false}`, "management.User.Create", 201},
	{"GET", "/api/v1/management/users/13", `{}`, "management.User.ReadById", 200},
	{"GET", "/api/v1/management/users?page=3&limit=25&search=value-search", `{}`, "management.User.ReadByFilter", 200},
	{"PATCH", "/api/v1/management/users/13", `{"name": "value-name", "bio": "value-bio", "avatar_path": "value-avatar_path", "avatar_path_set_null": false, "preferences": {"theme": "dark"}}`, "management.User.UpdateById", 204},
	{"DELETE", "/api/v1/management/users/13", `{}`, "management.User.DeleteById", 204},
	{"POST", "/api/v1/management/classic-auths", `{"user_id": 47, "username": "value-username", "email": "value-email", "phone": "value-phone", "password": "value-password", "is_email_verified": false, "is_phone_verified": false, "preferences": {"theme": "dark"}}`, "management.ClassicAuth.Create", 201},
	{"GET", "/api/v1/management/classic-auths/43", `{}`, "management.ClassicAuth.ReadById", 200},
	{"GET", "/api/v1/management/classic-auths/by-user/13", `{}`, "management.ClassicAuth.ReadByUserId", 200},
	{"GET", "/api/v1/management/classic-auths/by-username/lookup-user", `{}`, "management.ClassicAuth.ReadByUsername", 200},
	{"GET", "/api/v1/management/classic-auths/by-email?email=lookup-value", `{}`, "management.ClassicAuth.ReadByEmail", 200},
	{"GET", "/api/v1/management/classic-auths/by-phone?phone=lookup-value", `{}`, "management.ClassicAuth.ReadByPhone", 200},
	{"GET", "/api/v1/management/classic-auths?page=3&limit=25&search=value-search&user_id=47&is_email_verified=false&is_phone_verified=false", `{}`, "management.ClassicAuth.ReadByFilter", 200},
	{"PATCH", "/api/v1/management/classic-auths/43", `{"username": "value-username", "email": "value-email", "email_set_null": false, "phone": "value-phone", "phone_set_null": false, "is_email_verified": false, "is_phone_verified": false, "preferences": {"theme": "dark"}}`, "management.ClassicAuth.UpdateById", 204},
	{"PATCH", "/api/v1/management/users/13/password", `{"password": "value-password"}`, "management.ClassicAuth.ResetPassword", 204},
	{"DELETE", "/api/v1/management/classic-auths/43", `{}`, "management.ClassicAuth.DeleteById", 204},
	{"GET", "/api/v1/management/google-auths/43", `{}`, "management.GoogleAuth.ReadById", 200},
	{"GET", "/api/v1/management/google-auths/by-user/13", `{}`, "management.GoogleAuth.ReadByUserId", 200},
	{"GET", "/api/v1/management/google-auths/by-subject?subject=lookup-value", `{}`, "management.GoogleAuth.ReadBySubject", 200},
	{"GET", "/api/v1/management/google-auths?page=3&limit=25&search=value-search&user_id=47&is_email_verified=false", `{}`, "management.GoogleAuth.ReadByFilter", 200},
	{"POST", "/api/v1/management/spaces/7/clients", `{"slug": "value-slug", "name": "value-name", "description": "value-description", "is_active": false, "preferences": {"theme": "dark"}}`, "management.Client.Create", 201},
	{"GET", "/api/v1/management/spaces/7/clients/23", `{}`, "management.Client.ReadById", 200},
	{"GET", "/api/v1/management/spaces/7/clients/by-slug/lookup-slug", `{}`, "management.Client.ReadBySlug", 200},
	{"GET", "/api/v1/management/spaces/7/clients?page=3&limit=25&search=value-search&is_active=false", `{}`, "management.Client.ReadByFilter", 200},
	{"PATCH", "/api/v1/management/spaces/7/clients/23", `{"name": "value-name", "description": "value-description", "is_active": false, "preferences": {"theme": "dark"}}`, "management.Client.UpdateById", 204},
	{"PATCH", "/api/v1/management/spaces/7/clients/23/secret", `{}`, "management.Client.RotateSecret", 200},
	{"DELETE", "/api/v1/management/spaces/7/clients/23", `{}`, "management.Client.DeleteById", 204},
	{"POST", "/api/v1/management/spaces/7/clients/23/ip-allowlist", `{"cidr": "203.0.113.0/24", "is_active": false}`, "management.Client.CreateIpAllowlist", 201},
	{"GET", "/api/v1/management/spaces/7/clients/23/ip-allowlist/41", `{}`, "management.Client.ReadIpAllowlistById", 200},
	{"GET", "/api/v1/management/spaces/7/clients/23/ip-allowlist/by-cidr?cidr=203.0.113.0%2F24", `{}`, "management.Client.ReadIpAllowlistByCidr", 200},
	{"GET", "/api/v1/management/spaces/7/clients/23/ip-allowlist/all", `{}`, "management.Client.ReadIpAllowlistsByClientId", 200},
	{"GET", "/api/v1/management/spaces/7/clients/23/ip-allowlist?page=3&limit=25&search=value-search&is_active=false", `{}`, "management.Client.ReadIpAllowlistsByFilter", 200},
	{"PATCH", "/api/v1/management/spaces/7/clients/23/ip-allowlist/41", `{"cidr": "203.0.113.0/24", "is_active": false}`, "management.Client.UpdateIpAllowlistById", 204},
	{"DELETE", "/api/v1/management/spaces/7/clients/23/ip-allowlist/41", `{}`, "management.Client.DeleteIpAllowlistById", 204},
	{"POST", "/api/v1/management/spaces/7/space-members", `{"user_id": 47, "is_active": false, "preferences": {"theme": "dark"}}`, "management.SpaceMember.Create", 201},
	{"GET", "/api/v1/management/spaces/7/space-members/29", `{}`, "management.SpaceMember.ReadById", 200},
	{"GET", "/api/v1/management/spaces/7/space-members/by-user/13", `{}`, "management.SpaceMember.ReadByUserId", 200},
	{"GET", "/api/v1/management/spaces/7/space-members?page=3&limit=25&search=value-search&user_id=47&role_id=47&is_active=false", `{}`, "management.SpaceMember.ReadByFilter", 200},
	{"PATCH", "/api/v1/management/spaces/7/space-members/29", `{"is_active": false, "preferences": {"theme": "dark"}}`, "management.SpaceMember.UpdateById", 204},
	{"GET", "/api/v1/management/spaces/7/space-members/29/roles", `{}`, "management.SpaceMember.ReadEffectiveRolesBySpaceMemberId", 200},
	{"GET", "/api/v1/management/spaces/7/space-members/29/permissions", `{}`, "management.SpaceMember.ReadEffectivePermissionsBySpaceMemberId", 200},
	{"DELETE", "/api/v1/management/spaces/7/space-members/29", `{}`, "management.SpaceMember.DeleteById", 204},
	{"POST", "/api/v1/management/spaces/7/role-permissions", `{"role_id": 47, "permission_id": 47}`, "management.RolePermission.Create", 201},
	{"GET", "/api/v1/management/spaces/7/role-permissions/31", `{}`, "management.RolePermission.ReadById", 200},
	{"GET", "/api/v1/management/spaces/7/role-permissions/by-pair?role_id=17&permission_id=19", `{}`, "management.RolePermission.ReadByRoleIdAndPermissionId", 200},
	{"GET", "/api/v1/management/spaces/7/role-permissions/by-role/17", `{}`, "management.RolePermission.ReadByRoleId", 200},
	{"GET", "/api/v1/management/spaces/7/role-permissions/by-permission/19", `{}`, "management.RolePermission.ReadByPermissionId", 200},
	{"DELETE", "/api/v1/management/spaces/7/role-permissions/31", `{}`, "management.RolePermission.DeleteById", 204},
	{"DELETE", "/api/v1/management/spaces/7/role-permissions?role_id=47&permission_id=47", `{}`, "management.RolePermission.DeleteByRoleIdOrPermissionId", 204},
	{"POST", "/api/v1/management/spaces/7/space-member-roles", `{"space_member_id": 47, "role_id": 47}`, "management.SpaceMemberRole.Create", 201},
	{"GET", "/api/v1/management/spaces/7/space-member-roles/37", `{}`, "management.SpaceMemberRole.ReadById", 200},
	{"GET", "/api/v1/management/spaces/7/space-member-roles/by-pair?space_member_id=29&role_id=17", `{}`, "management.SpaceMemberRole.ReadBySpaceMemberIdAndRoleId", 200},
	{"GET", "/api/v1/management/spaces/7/space-member-roles/by-space-member/29", `{}`, "management.SpaceMemberRole.ReadBySpaceMemberId", 200},
	{"GET", "/api/v1/management/spaces/7/space-member-roles/by-role/17", `{}`, "management.SpaceMemberRole.ReadByRoleId", 200},
	{"DELETE", "/api/v1/management/spaces/7/space-member-roles/37", `{}`, "management.SpaceMemberRole.DeleteById", 204},
	{"DELETE", "/api/v1/management/spaces/7/space-member-roles?space_member_id=47&role_id=47", `{}`, "management.SpaceMemberRole.DeleteBySpaceMemberIdOrRoleId", 204},
	{"POST", "/api/v1/auth/sign-in", `{"username": "value-username", "password": "value-password"}`, "auth.Session.SignIn", 200},
	{"POST", "/api/v1/auth/register", `{"name": "value-name", "username": "value-username", "password": "value-password", "bio": "value-bio", "email": "value-email", "phone": "value-phone"}`, "auth.Session.Register", 201},
	{"POST", "/api/v1/auth/refresh", `{"refresh_token": "value-refresh_token"}`, "auth.Session.Refresh", 200},
	{"GET", "/api/v1/profile", `{}`, "profile.Account.GetProfile", 200},
	{"GET", "/api/v1/profile/roles", `{}`, "profile.Account.GetRoles", 200},
	{"GET", "/api/v1/profile/permissions", `{}`, "profile.Account.GetPermissions", 200},
	{"PATCH", "/api/v1/profile", `{"name": "value-name", "bio": "value-bio", "avatar_path": "value-avatar_path", "avatar_path_set_null": false, "preferences": {"theme": "dark"}}`, "profile.Account.UpdateProfile", 204},
	{"PATCH", "/api/v1/profile/membership/preferences", `{"preferences": {"theme": "dark"}}`, "profile.Account.UpdateMembershipPreferences", 204},
	{"GET", "/api/v1/profile/auth-methods", `{}`, "profile.Security.GetAuthMethods", 200},
	{"PATCH", "/api/v1/profile/classic-auth", `{"username": "value-username", "email": "value-email", "email_set_null": false, "phone": "value-phone", "phone_set_null": false}`, "profile.Security.UpdateClassicAuth", 204},
	{"PATCH", "/api/v1/profile/password", `{"current_password": "value-current_password", "new_password": "value-new_password"}`, "profile.Security.ChangePassword", 204},
}

func TestAllOperationsDispatchAndErrors(t *testing.T) {
	for _, item := range operationCases {
		t.Run(item.operation, func(t *testing.T) {
			fixture := newHTTPFixture()
			response := perform(fixture.router, businessRequest(item.method, item.path, item.request))
			if response.Code != item.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if len(fixture.recorder.calls) != 1 || fixture.recorder.calls[0].name != item.operation {
				t.Fatalf("wrong operation: %+v", fixture.recorder.calls)
			}
			call := fixture.recorder.calls[0]
			if call.ctx.Value(traceMarker{}) == nil {
				t.Fatal("trace context was discarded")
			}
			if strings.HasPrefix(item.operation, "auth.") {
				if client := call.args[0].(*domainmodelsentity.ClientIdentity); client.Id != 1 || client.SpaceId != 2 {
					t.Fatal("client identity was not trusted")
				}
			} else {
				if caller := call.args[0].(*domainmodelsentity.AccessClaims); caller.UserId != 3 || caller.Permissions[0] != "fresh:get" {
					t.Fatal("caller claims were not reloaded")
				}
			}
			if strings.Contains(response.Body.String(), "hash") {
				t.Fatal("credential hash leaked")
			}
			if item.status == 204 && response.Body.Len() != 0 {
				t.Fatal("204 response has body")
			}
			if strings.Contains(item.operation, "ReadByFilter") || strings.Contains(item.operation, "ReadIpAllowlistsByFilter") {
				var page struct {
					Data []any
					Page struct{ Page, Limit, TotalItems int64 }
				}
				if err := json.Unmarshal(response.Body.Bytes(), &page); err != nil {
					t.Fatal(err)
				}
				// Check the actual snake_case wire names separately.
				if !strings.Contains(response.Body.String(), `"total_items":19`) || page.Page.Page != 3 || page.Page.Limit != 25 || page.Data == nil {
					t.Fatalf("bad page response: %s", response.Body.String())
				}
			}
			if strings.HasPrefix(item.operation, "auth.") || strings.HasSuffix(item.operation, ".RotateSecret") || item.operation == "management.Client.Create" {
				if response.Header().Get("Cache-Control") != "no-store" {
					t.Fatal("credentials are cacheable")
				}
			}
			fixture.recorder.calls = nil
			fixture.recorder.failure = domainmodelserror.NewError("operation conflict", domainmodelserror.Conflict, fmt.Errorf("private diagnostic"))
			response = perform(fixture.router, businessRequest(item.method, item.path, item.request))
			if response.Code != 409 || strings.Contains(response.Body.String(), "private diagnostic") {
				t.Fatalf("bad error response: %d %s", response.Code, response.Body.String())
			}
		})
	}
}
func TestHandlerArgumentsAndClearFlags(t *testing.T) {
	fixture := newHTTPFixture()
	req := businessRequest("PATCH", "/api/v1/management/spaces/7/clients/23/ip-allowlist/41", `{"cidr":"198.51.100.0/24","is_active":false,"created_by":900,"client_id":900}`)
	response := perform(fixture.router, req)
	if response.Code != 204 {
		t.Fatalf("%d: %s", response.Code, response.Body.String())
	}
	call := fixture.recorder.calls[0]
	if call.args[1] != int64(7) || call.args[2] != int64(23) || call.args[3] != int64(41) {
		t.Fatalf("wrong ownership: %+v", call.args)
	}
	arg := call.args[4].(domainusecasesmanagement.UpdateClientIpAllowlistRequest)
	if arg.Cidr == nil || *arg.Cidr != "198.51.100.0/24" || arg.IsActive == nil || *arg.IsActive {
		t.Fatal("optional update fields lost")
	}
	fixture.recorder.calls = nil
	response = perform(fixture.router, businessRequest("PATCH", "/api/v1/profile", `{"avatar_path_set_null":true}`))
	if response.Code != 204 {
		t.Fatal(response.Body.String())
	}
	profile := fixture.recorder.calls[0].args[1].(domainusecasesprofile.UpdateProfileRequest)
	if !profile.AvatarPathSetNull || profile.Name != nil || profile.AvatarPath != nil {
		t.Fatal("clear/omitted fields changed")
	}
	fixture.recorder.calls = nil
	response = perform(fixture.router, businessRequest("DELETE", "/api/v1/management/spaces/7/role-permissions?role_id=17&permission_id=19", ""))
	if response.Code != 204 {
		t.Fatal(response.Body.String())
	}
	deletion := fixture.recorder.calls[0].args[2].(domainusecasesmanagement.DeleteRolePermissionRequest)
	if deletion.RoleId == nil || *deletion.RoleId != 17 || deletion.PermissionId == nil || *deletion.PermissionId != 19 {
		t.Fatal("bulk selectors changed")
	}
}
func TestPublicDocsVersionAndMissingRoutes(t *testing.T) {
	fixture := newHTTPFixture()
	for _, item := range []struct {
		path   string
		status int
	}{
		{"/api/version", 200}, {"/api/docs", 301}, {"/api/docs/index.html", 200}, {"/api/docs/doc.json", 200}, {"/does-not-exist", 404},
	} {
		response := perform(fixture.router, httptest.NewRequest("GET", item.path, nil))
		if response.Code != item.status {
			t.Fatalf("%s: %d %s", item.path, response.Code, response.Body.String())
		}
	}
	response := perform(fixture.router, httptest.NewRequest("POST", "/api/version", nil))
	if response.Code != 405 || !strings.Contains(response.Body.String(), "Method Not Allowed") {
		t.Fatal("missing JSON method response")
	}
	if fixture.client.calls != 0 {
		t.Fatal("public route authenticated client")
	}
}
func TestMalformedTransportDoesNotInvokeUsecase(t *testing.T) {
	for _, item := range []struct{ method, path, body string }{
		{"POST", "/api/v1/auth/sign-in", `null`}, {"POST", "/api/v1/auth/sign-in", `[]`},
		{"POST", "/api/v1/auth/sign-in", `{}`}, {"POST", "/api/v1/auth/sign-in", `{"username":"u","password":"p"} {}`},
		{"PATCH", "/api/v1/profile", `{"name":42}`}, {"GET", "/api/v1/management/spaces/not-int", ``},
		{"GET", "/api/v1/management/users?page=abc", ``}, {"GET", "/api/v1/management/spaces?is_active=nope", ``},
		{"GET", "/api/v1/management/spaces/7/role-permissions/by-pair?role_id=3", ``},
		{"GET", "/api/v1/management/classic-auths/by-email", ``},
	} {
		t.Run(item.path+item.body, func(t *testing.T) {
			fixture := newHTTPFixture()
			response := perform(fixture.router, businessRequest(item.method, item.path, item.body))
			if response.Code != 400 || len(fixture.recorder.calls) != 0 {
				t.Fatalf("bad transport accepted: %d %s", response.Code, response.Body.String())
			}
		})
	}
	fixture := newHTTPFixture()
	req := businessRequest("PATCH", "/api/v1/profile", `{}`)
	req.Header.Set("Content-Type", "text/plain")
	if response := perform(fixture.router, req); response.Code != 400 {
		t.Fatal("non-JSON body accepted")
	}
	response := perform(fixture.router, businessRequest("POST", "/api/v1/auth/sign-in", `{"username":"u","password":"  preserved  "}`))
	if response.Code != 200 {
		t.Fatal(response.Body.String())
	}
	request := fixture.recorder.calls[0].args[1].(domainusecasesauth.SignInRequest)
	if request.Password != "  preserved  " {
		t.Fatal("password was trimmed")
	}
}
func TestPaginationDefaultsAndLimit(t *testing.T) {
	fixture := newHTTPFixture()
	response := perform(fixture.router, businessRequest("GET", "/api/v1/management/users?page=-1&limit=200", ``))
	if response.Code != 200 {
		t.Fatal(response.Body.String())
	}
	request := fixture.recorder.calls[0].args[1].(domainusecasesmanagement.FilterUserRequest)
	if request.Page != 1 || request.Limit != 100 {
		t.Fatalf("pagination %+v", request)
	}
}
func TestMissingUsecaseResultIsSanitized(t *testing.T) {
	fixture := newHTTPFixture()
	fixture.recorder.nilResult = true
	response := perform(fixture.router, businessRequest("GET", "/api/v1/profile", ``))
	if response.Code != 500 || strings.Contains(response.Body.String(), "missing from usecase") {
		t.Fatal("nil result not sanitized")
	}
}
