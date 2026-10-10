package applicationsharedtest

import (
	"context"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	"github.com/google/uuid"
)

type SpaceMemberRole struct {
	CreateFn                                  func(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateSpaceMemberRoleArgs) (id int64, err error)
	ReadByIdFn                                func(ctx context.Context, spaceId, id int64) (*domaincontractsrepository.SpaceMemberRoleDetails, error)
	ReadBySpaceMemberIdAndRoleIdFn            func(ctx context.Context, spaceId, spaceMemberId, roleId int64) (*domaincontractsrepository.SpaceMemberRoleDetails, error)
	ReadBySpaceMemberIdFn                     func(ctx context.Context, spaceId, value int64) ([]domaincontractsrepository.SpaceMemberRoleWithRole, error)
	ReadByRoleIdFn                            func(ctx context.Context, spaceId, value int64) ([]domaincontractsrepository.SpaceMemberRoleWithSpaceMember, error)
	ReadEffectiveRolesBySpaceMemberIdFn       func(ctx context.Context, spaceId, spaceMemberId int64) ([]domainmodelsentity.Role, error)
	ReadEffectivePermissionsBySpaceMemberIdFn func(ctx context.Context, spaceId, spaceMemberId int64) ([]domainmodelsentity.Permission, error)
	DeleteByIdFn                              func(ctx context.Context, spaceId, id int64) error
	DeleteBySpaceMemberIdOrRoleIdFn           func(ctx context.Context, spaceId int64, args domaincontractsrepository.DeleteSpaceMemberRoleArgs) error
}

var _ domaincontractsrepository.SpaceMemberRole = (*SpaceMemberRole)(nil)

func (s *SpaceMemberRole) Create(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateSpaceMemberRoleArgs) (id int64, err error) {
	if s.CreateFn == nil {
		panic("unexpected SpaceMemberRole.Create call")
	}
	return s.CreateFn(ctx, spaceId, args)
}

func (s *SpaceMemberRole) ReadById(ctx context.Context, spaceId, id int64) (*domaincontractsrepository.SpaceMemberRoleDetails, error) {
	if s.ReadByIdFn == nil {
		panic("unexpected SpaceMemberRole.ReadById call")
	}
	return s.ReadByIdFn(ctx, spaceId, id)
}

func (s *SpaceMemberRole) ReadBySpaceMemberIdAndRoleId(ctx context.Context, spaceId, spaceMemberId, roleId int64) (*domaincontractsrepository.SpaceMemberRoleDetails, error) {
	if s.ReadBySpaceMemberIdAndRoleIdFn == nil {
		panic("unexpected SpaceMemberRole.ReadBySpaceMemberIdAndRoleId call")
	}
	return s.ReadBySpaceMemberIdAndRoleIdFn(ctx, spaceId, spaceMemberId, roleId)
}

func (s *SpaceMemberRole) ReadBySpaceMemberId(ctx context.Context, spaceId, value int64) ([]domaincontractsrepository.SpaceMemberRoleWithRole, error) {
	if s.ReadBySpaceMemberIdFn == nil {
		panic("unexpected SpaceMemberRole.ReadBySpaceMemberId call")
	}
	return s.ReadBySpaceMemberIdFn(ctx, spaceId, value)
}

func (s *SpaceMemberRole) ReadByRoleId(ctx context.Context, spaceId, value int64) ([]domaincontractsrepository.SpaceMemberRoleWithSpaceMember, error) {
	if s.ReadByRoleIdFn == nil {
		panic("unexpected SpaceMemberRole.ReadByRoleId call")
	}
	return s.ReadByRoleIdFn(ctx, spaceId, value)
}

func (s *SpaceMemberRole) ReadEffectiveRolesBySpaceMemberId(ctx context.Context, spaceId, spaceMemberId int64) ([]domainmodelsentity.Role, error) {
	if s.ReadEffectiveRolesBySpaceMemberIdFn == nil {
		panic("unexpected SpaceMemberRole.ReadEffectiveRolesBySpaceMemberId call")
	}
	return s.ReadEffectiveRolesBySpaceMemberIdFn(ctx, spaceId, spaceMemberId)
}

func (s *SpaceMemberRole) ReadEffectivePermissionsBySpaceMemberId(ctx context.Context, spaceId, spaceMemberId int64) ([]domainmodelsentity.Permission, error) {
	if s.ReadEffectivePermissionsBySpaceMemberIdFn == nil {
		panic("unexpected SpaceMemberRole.ReadEffectivePermissionsBySpaceMemberId call")
	}
	return s.ReadEffectivePermissionsBySpaceMemberIdFn(ctx, spaceId, spaceMemberId)
}

func (s *SpaceMemberRole) DeleteById(ctx context.Context, spaceId, id int64) error {
	if s.DeleteByIdFn == nil {
		panic("unexpected SpaceMemberRole.DeleteById call")
	}
	return s.DeleteByIdFn(ctx, spaceId, id)
}

func (s *SpaceMemberRole) DeleteBySpaceMemberIdOrRoleId(ctx context.Context, spaceId int64, args domaincontractsrepository.DeleteSpaceMemberRoleArgs) error {
	if s.DeleteBySpaceMemberIdOrRoleIdFn == nil {
		panic("unexpected SpaceMemberRole.DeleteBySpaceMemberIdOrRoleId call")
	}
	return s.DeleteBySpaceMemberIdOrRoleIdFn(ctx, spaceId, args)
}

type User struct {
	CreateFn       func(ctx context.Context, args domaincontractsrepository.CreateUserArgs) (id int64, err error)
	ReadByIdFn     func(ctx context.Context, id int64) (*domainmodelsentity.User, error)
	ReadByFilterFn func(ctx context.Context, filter domaincontractsrepository.UserFilter) (items []domainmodelsentity.User, total int64, err error)
	UpdateByIdFn   func(ctx context.Context, id int64, args domaincontractsrepository.UpdateUserArgs) error
	DeleteByIdFn   func(ctx context.Context, id int64, by *int64) error
}

var _ domaincontractsrepository.User = (*User)(nil)

func (s *User) Create(ctx context.Context, args domaincontractsrepository.CreateUserArgs) (id int64, err error) {
	if s.CreateFn == nil {
		panic("unexpected User.Create call")
	}
	return s.CreateFn(ctx, args)
}

func (s *User) ReadById(ctx context.Context, id int64) (*domainmodelsentity.User, error) {
	if s.ReadByIdFn == nil {
		panic("unexpected User.ReadById call")
	}
	return s.ReadByIdFn(ctx, id)
}

func (s *User) ReadByFilter(ctx context.Context, filter domaincontractsrepository.UserFilter) (items []domainmodelsentity.User, total int64, err error) {
	if s.ReadByFilterFn == nil {
		panic("unexpected User.ReadByFilter call")
	}
	return s.ReadByFilterFn(ctx, filter)
}

func (s *User) UpdateById(ctx context.Context, id int64, args domaincontractsrepository.UpdateUserArgs) error {
	if s.UpdateByIdFn == nil {
		panic("unexpected User.UpdateById call")
	}
	return s.UpdateByIdFn(ctx, id, args)
}

func (s *User) DeleteById(ctx context.Context, id int64, by *int64) error {
	if s.DeleteByIdFn == nil {
		panic("unexpected User.DeleteById call")
	}
	return s.DeleteByIdFn(ctx, id, by)
}

type RolePermission struct {
	CreateFn                       func(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateRolePermissionArgs) (id int64, err error)
	ReadByIdFn                     func(ctx context.Context, spaceId, id int64) (*domaincontractsrepository.RolePermissionDetails, error)
	ReadByRoleIdAndPermissionIdFn  func(ctx context.Context, spaceId, roleId, permissionId int64) (*domaincontractsrepository.RolePermissionDetails, error)
	ReadByRoleIdFn                 func(ctx context.Context, spaceId, value int64) ([]domaincontractsrepository.RolePermissionWithPermission, error)
	ReadByPermissionIdFn           func(ctx context.Context, spaceId, value int64) ([]domaincontractsrepository.RolePermissionWithRole, error)
	DeleteByIdFn                   func(ctx context.Context, spaceId, id int64) error
	DeleteByRoleIdOrPermissionIdFn func(ctx context.Context, spaceId int64, args domaincontractsrepository.DeleteRolePermissionArgs) error
}

var _ domaincontractsrepository.RolePermission = (*RolePermission)(nil)

func (s *RolePermission) Create(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateRolePermissionArgs) (id int64, err error) {
	if s.CreateFn == nil {
		panic("unexpected RolePermission.Create call")
	}
	return s.CreateFn(ctx, spaceId, args)
}

func (s *RolePermission) ReadById(ctx context.Context, spaceId, id int64) (*domaincontractsrepository.RolePermissionDetails, error) {
	if s.ReadByIdFn == nil {
		panic("unexpected RolePermission.ReadById call")
	}
	return s.ReadByIdFn(ctx, spaceId, id)
}

func (s *RolePermission) ReadByRoleIdAndPermissionId(ctx context.Context, spaceId, roleId, permissionId int64) (*domaincontractsrepository.RolePermissionDetails, error) {
	if s.ReadByRoleIdAndPermissionIdFn == nil {
		panic("unexpected RolePermission.ReadByRoleIdAndPermissionId call")
	}
	return s.ReadByRoleIdAndPermissionIdFn(ctx, spaceId, roleId, permissionId)
}

func (s *RolePermission) ReadByRoleId(ctx context.Context, spaceId, value int64) ([]domaincontractsrepository.RolePermissionWithPermission, error) {
	if s.ReadByRoleIdFn == nil {
		panic("unexpected RolePermission.ReadByRoleId call")
	}
	return s.ReadByRoleIdFn(ctx, spaceId, value)
}

func (s *RolePermission) ReadByPermissionId(ctx context.Context, spaceId, value int64) ([]domaincontractsrepository.RolePermissionWithRole, error) {
	if s.ReadByPermissionIdFn == nil {
		panic("unexpected RolePermission.ReadByPermissionId call")
	}
	return s.ReadByPermissionIdFn(ctx, spaceId, value)
}

func (s *RolePermission) DeleteById(ctx context.Context, spaceId, id int64) error {
	if s.DeleteByIdFn == nil {
		panic("unexpected RolePermission.DeleteById call")
	}
	return s.DeleteByIdFn(ctx, spaceId, id)
}

func (s *RolePermission) DeleteByRoleIdOrPermissionId(ctx context.Context, spaceId int64, args domaincontractsrepository.DeleteRolePermissionArgs) error {
	if s.DeleteByRoleIdOrPermissionIdFn == nil {
		panic("unexpected RolePermission.DeleteByRoleIdOrPermissionId call")
	}
	return s.DeleteByRoleIdOrPermissionIdFn(ctx, spaceId, args)
}

type SpaceMember struct {
	CreateFn       func(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateSpaceMemberArgs) (id int64, err error)
	ReadByIdFn     func(ctx context.Context, spaceId int64, id int64) (*domainmodelsentity.SpaceMember, error)
	ReadByUserIdFn func(ctx context.Context, spaceId int64, value int64) (*domainmodelsentity.SpaceMember, error)
	ReadByFilterFn func(ctx context.Context, spaceId int64, filter domaincontractsrepository.SpaceMemberFilter) (items []domaincontractsrepository.SpaceMemberWithUser, total int64, err error)
	UpdateByIdFn   func(ctx context.Context, spaceId int64, id int64, args domaincontractsrepository.UpdateSpaceMemberArgs) error
	DeleteByIdFn   func(ctx context.Context, spaceId int64, id int64, by *int64) error
}

var _ domaincontractsrepository.SpaceMember = (*SpaceMember)(nil)

func (s *SpaceMember) Create(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateSpaceMemberArgs) (id int64, err error) {
	if s.CreateFn == nil {
		panic("unexpected SpaceMember.Create call")
	}
	return s.CreateFn(ctx, spaceId, args)
}

func (s *SpaceMember) ReadById(ctx context.Context, spaceId int64, id int64) (*domainmodelsentity.SpaceMember, error) {
	if s.ReadByIdFn == nil {
		panic("unexpected SpaceMember.ReadById call")
	}
	return s.ReadByIdFn(ctx, spaceId, id)
}

func (s *SpaceMember) ReadByUserId(ctx context.Context, spaceId int64, value int64) (*domainmodelsentity.SpaceMember, error) {
	if s.ReadByUserIdFn == nil {
		panic("unexpected SpaceMember.ReadByUserId call")
	}
	return s.ReadByUserIdFn(ctx, spaceId, value)
}

func (s *SpaceMember) ReadByFilter(ctx context.Context, spaceId int64, filter domaincontractsrepository.SpaceMemberFilter) (items []domaincontractsrepository.SpaceMemberWithUser, total int64, err error) {
	if s.ReadByFilterFn == nil {
		panic("unexpected SpaceMember.ReadByFilter call")
	}
	return s.ReadByFilterFn(ctx, spaceId, filter)
}

func (s *SpaceMember) UpdateById(ctx context.Context, spaceId int64, id int64, args domaincontractsrepository.UpdateSpaceMemberArgs) error {
	if s.UpdateByIdFn == nil {
		panic("unexpected SpaceMember.UpdateById call")
	}
	return s.UpdateByIdFn(ctx, spaceId, id, args)
}

func (s *SpaceMember) DeleteById(ctx context.Context, spaceId int64, id int64, by *int64) error {
	if s.DeleteByIdFn == nil {
		panic("unexpected SpaceMember.DeleteById call")
	}
	return s.DeleteByIdFn(ctx, spaceId, id, by)
}

type Role struct {
	CreateFn       func(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateRoleArgs) (id int64, err error)
	ReadByIdFn     func(ctx context.Context, spaceId int64, id int64) (*domainmodelsentity.Role, error)
	ReadBySlugFn   func(ctx context.Context, spaceId int64, value string) (*domainmodelsentity.Role, error)
	ReadByFilterFn func(ctx context.Context, spaceId int64, filter domaincontractsrepository.RoleFilter) (items []domainmodelsentity.Role, total int64, err error)
	ReadDefaultFn  func(ctx context.Context, spaceId int64) (*domainmodelsentity.Role, error)
	UpdateByIdFn   func(ctx context.Context, spaceId int64, id int64, args domaincontractsrepository.UpdateRoleArgs) error
	DeleteByIdFn   func(ctx context.Context, spaceId int64, id int64, by *int64) error
}

var _ domaincontractsrepository.Role = (*Role)(nil)

func (s *Role) Create(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateRoleArgs) (id int64, err error) {
	if s.CreateFn == nil {
		panic("unexpected Role.Create call")
	}
	return s.CreateFn(ctx, spaceId, args)
}

func (s *Role) ReadById(ctx context.Context, spaceId int64, id int64) (*domainmodelsentity.Role, error) {
	if s.ReadByIdFn == nil {
		panic("unexpected Role.ReadById call")
	}
	return s.ReadByIdFn(ctx, spaceId, id)
}

func (s *Role) ReadBySlug(ctx context.Context, spaceId int64, value string) (*domainmodelsentity.Role, error) {
	if s.ReadBySlugFn == nil {
		panic("unexpected Role.ReadBySlug call")
	}
	return s.ReadBySlugFn(ctx, spaceId, value)
}

func (s *Role) ReadByFilter(ctx context.Context, spaceId int64, filter domaincontractsrepository.RoleFilter) (items []domainmodelsentity.Role, total int64, err error) {
	if s.ReadByFilterFn == nil {
		panic("unexpected Role.ReadByFilter call")
	}
	return s.ReadByFilterFn(ctx, spaceId, filter)
}

func (s *Role) ReadDefault(ctx context.Context, spaceId int64) (*domainmodelsentity.Role, error) {
	if s.ReadDefaultFn == nil {
		panic("unexpected Role.ReadDefault call")
	}
	return s.ReadDefaultFn(ctx, spaceId)
}

func (s *Role) UpdateById(ctx context.Context, spaceId int64, id int64, args domaincontractsrepository.UpdateRoleArgs) error {
	if s.UpdateByIdFn == nil {
		panic("unexpected Role.UpdateById call")
	}
	return s.UpdateByIdFn(ctx, spaceId, id, args)
}

func (s *Role) DeleteById(ctx context.Context, spaceId int64, id int64, by *int64) error {
	if s.DeleteByIdFn == nil {
		panic("unexpected Role.DeleteById call")
	}
	return s.DeleteByIdFn(ctx, spaceId, id, by)
}

type GoogleAuth struct {
	CreateFn        func(ctx context.Context, args domaincontractsrepository.CreateGoogleAuthArgs) (id int64, err error)
	ReadByIdFn      func(ctx context.Context, id int64) (*domainmodelsentity.GoogleAuth, error)
	ReadByUserIdFn  func(ctx context.Context, value int64) (*domainmodelsentity.GoogleAuth, error)
	ReadBySubjectFn func(ctx context.Context, value string) (*domainmodelsentity.GoogleAuth, error)
	ReadByFilterFn  func(ctx context.Context, filter domaincontractsrepository.GoogleAuthFilter) (items []domainmodelsentity.GoogleAuth, total int64, err error)
	UpdateByIdFn    func(ctx context.Context, id int64, args domaincontractsrepository.UpdateGoogleAuthArgs) error
	DeleteByIdFn    func(ctx context.Context, id int64, by *int64) error
}

var _ domaincontractsrepository.GoogleAuth = (*GoogleAuth)(nil)

func (s *GoogleAuth) Create(ctx context.Context, args domaincontractsrepository.CreateGoogleAuthArgs) (id int64, err error) {
	if s.CreateFn == nil {
		panic("unexpected GoogleAuth.Create call")
	}
	return s.CreateFn(ctx, args)
}

func (s *GoogleAuth) ReadById(ctx context.Context, id int64) (*domainmodelsentity.GoogleAuth, error) {
	if s.ReadByIdFn == nil {
		panic("unexpected GoogleAuth.ReadById call")
	}
	return s.ReadByIdFn(ctx, id)
}

func (s *GoogleAuth) ReadByUserId(ctx context.Context, value int64) (*domainmodelsentity.GoogleAuth, error) {
	if s.ReadByUserIdFn == nil {
		panic("unexpected GoogleAuth.ReadByUserId call")
	}
	return s.ReadByUserIdFn(ctx, value)
}

func (s *GoogleAuth) ReadBySubject(ctx context.Context, value string) (*domainmodelsentity.GoogleAuth, error) {
	if s.ReadBySubjectFn == nil {
		panic("unexpected GoogleAuth.ReadBySubject call")
	}
	return s.ReadBySubjectFn(ctx, value)
}

func (s *GoogleAuth) ReadByFilter(ctx context.Context, filter domaincontractsrepository.GoogleAuthFilter) (items []domainmodelsentity.GoogleAuth, total int64, err error) {
	if s.ReadByFilterFn == nil {
		panic("unexpected GoogleAuth.ReadByFilter call")
	}
	return s.ReadByFilterFn(ctx, filter)
}

func (s *GoogleAuth) UpdateById(ctx context.Context, id int64, args domaincontractsrepository.UpdateGoogleAuthArgs) error {
	if s.UpdateByIdFn == nil {
		panic("unexpected GoogleAuth.UpdateById call")
	}
	return s.UpdateByIdFn(ctx, id, args)
}

func (s *GoogleAuth) DeleteById(ctx context.Context, id int64, by *int64) error {
	if s.DeleteByIdFn == nil {
		panic("unexpected GoogleAuth.DeleteById call")
	}
	return s.DeleteByIdFn(ctx, id, by)
}

type ClassicAuth struct {
	CreateFn         func(ctx context.Context, args domaincontractsrepository.CreateClassicAuthArgs) (id int64, err error)
	ReadByIdFn       func(ctx context.Context, id int64) (*domainmodelsentity.ClassicAuth, error)
	ReadByUserIdFn   func(ctx context.Context, value int64) (*domainmodelsentity.ClassicAuth, error)
	ReadByUsernameFn func(ctx context.Context, value string) (*domainmodelsentity.ClassicAuth, error)
	ReadByEmailFn    func(ctx context.Context, value string) (*domainmodelsentity.ClassicAuth, error)
	ReadByPhoneFn    func(ctx context.Context, value string) (*domainmodelsentity.ClassicAuth, error)
	ReadByFilterFn   func(ctx context.Context, filter domaincontractsrepository.ClassicAuthFilter) (items []domaincontractsrepository.ClassicAuthSummary, total int64, err error)
	UpdateByIdFn     func(ctx context.Context, id int64, args domaincontractsrepository.UpdateClassicAuthArgs) error
	DeleteByIdFn     func(ctx context.Context, id int64, by *int64) error
}

var _ domaincontractsrepository.ClassicAuth = (*ClassicAuth)(nil)

func (s *ClassicAuth) Create(ctx context.Context, args domaincontractsrepository.CreateClassicAuthArgs) (id int64, err error) {
	if s.CreateFn == nil {
		panic("unexpected ClassicAuth.Create call")
	}
	return s.CreateFn(ctx, args)
}

func (s *ClassicAuth) ReadById(ctx context.Context, id int64) (*domainmodelsentity.ClassicAuth, error) {
	if s.ReadByIdFn == nil {
		panic("unexpected ClassicAuth.ReadById call")
	}
	return s.ReadByIdFn(ctx, id)
}

func (s *ClassicAuth) ReadByUserId(ctx context.Context, value int64) (*domainmodelsentity.ClassicAuth, error) {
	if s.ReadByUserIdFn == nil {
		panic("unexpected ClassicAuth.ReadByUserId call")
	}
	return s.ReadByUserIdFn(ctx, value)
}

func (s *ClassicAuth) ReadByUsername(ctx context.Context, value string) (*domainmodelsentity.ClassicAuth, error) {
	if s.ReadByUsernameFn == nil {
		panic("unexpected ClassicAuth.ReadByUsername call")
	}
	return s.ReadByUsernameFn(ctx, value)
}

func (s *ClassicAuth) ReadByEmail(ctx context.Context, value string) (*domainmodelsentity.ClassicAuth, error) {
	if s.ReadByEmailFn == nil {
		panic("unexpected ClassicAuth.ReadByEmail call")
	}
	return s.ReadByEmailFn(ctx, value)
}

func (s *ClassicAuth) ReadByPhone(ctx context.Context, value string) (*domainmodelsentity.ClassicAuth, error) {
	if s.ReadByPhoneFn == nil {
		panic("unexpected ClassicAuth.ReadByPhone call")
	}
	return s.ReadByPhoneFn(ctx, value)
}

func (s *ClassicAuth) ReadByFilter(ctx context.Context, filter domaincontractsrepository.ClassicAuthFilter) (items []domaincontractsrepository.ClassicAuthSummary, total int64, err error) {
	if s.ReadByFilterFn == nil {
		panic("unexpected ClassicAuth.ReadByFilter call")
	}
	return s.ReadByFilterFn(ctx, filter)
}

func (s *ClassicAuth) UpdateById(ctx context.Context, id int64, args domaincontractsrepository.UpdateClassicAuthArgs) error {
	if s.UpdateByIdFn == nil {
		panic("unexpected ClassicAuth.UpdateById call")
	}
	return s.UpdateByIdFn(ctx, id, args)
}

func (s *ClassicAuth) DeleteById(ctx context.Context, id int64, by *int64) error {
	if s.DeleteByIdFn == nil {
		panic("unexpected ClassicAuth.DeleteById call")
	}
	return s.DeleteByIdFn(ctx, id, by)
}

type Client struct {
	CreateFn       func(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateClientArgs) (id int64, err error)
	ReadByIdFn     func(ctx context.Context, spaceId int64, id int64) (*domainmodelsentity.Client, error)
	ReadBySlugFn   func(ctx context.Context, value string) (*domainmodelsentity.Client, error)
	ReadByFilterFn func(ctx context.Context, spaceId int64, filter domaincontractsrepository.ClientFilter) (items []domaincontractsrepository.ClientSummary, total int64, err error)
	UpdateByIdFn   func(ctx context.Context, spaceId int64, id int64, args domaincontractsrepository.UpdateClientArgs) error
	DeleteByIdFn   func(ctx context.Context, spaceId int64, id int64, by *int64) error
}

var _ domaincontractsrepository.Client = (*Client)(nil)

func (s *Client) Create(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateClientArgs) (id int64, err error) {
	if s.CreateFn == nil {
		panic("unexpected Client.Create call")
	}
	return s.CreateFn(ctx, spaceId, args)
}

func (s *Client) ReadById(ctx context.Context, spaceId int64, id int64) (*domainmodelsentity.Client, error) {
	if s.ReadByIdFn == nil {
		panic("unexpected Client.ReadById call")
	}
	return s.ReadByIdFn(ctx, spaceId, id)
}

func (s *Client) ReadBySlug(ctx context.Context, value string) (*domainmodelsentity.Client, error) {
	if s.ReadBySlugFn == nil {
		panic("unexpected Client.ReadBySlug call")
	}
	return s.ReadBySlugFn(ctx, value)
}

func (s *Client) ReadByFilter(ctx context.Context, spaceId int64, filter domaincontractsrepository.ClientFilter) (items []domaincontractsrepository.ClientSummary, total int64, err error) {
	if s.ReadByFilterFn == nil {
		panic("unexpected Client.ReadByFilter call")
	}
	return s.ReadByFilterFn(ctx, spaceId, filter)
}

func (s *Client) UpdateById(ctx context.Context, spaceId int64, id int64, args domaincontractsrepository.UpdateClientArgs) error {
	if s.UpdateByIdFn == nil {
		panic("unexpected Client.UpdateById call")
	}
	return s.UpdateByIdFn(ctx, spaceId, id, args)
}

func (s *Client) DeleteById(ctx context.Context, spaceId int64, id int64, by *int64) error {
	if s.DeleteByIdFn == nil {
		panic("unexpected Client.DeleteById call")
	}
	return s.DeleteByIdFn(ctx, spaceId, id, by)
}

type Space struct {
	CreateFn       func(ctx context.Context, args domaincontractsrepository.CreateSpaceArgs) (id int64, err error)
	ReadByIdFn     func(ctx context.Context, id int64) (*domainmodelsentity.Space, error)
	ReadBySlugFn   func(ctx context.Context, value string) (*domainmodelsentity.Space, error)
	ReadByFilterFn func(ctx context.Context, filter domaincontractsrepository.SpaceFilter) (items []domainmodelsentity.Space, total int64, err error)
	UpdateByIdFn   func(ctx context.Context, id int64, args domaincontractsrepository.UpdateSpaceArgs) error
	DeleteByIdFn   func(ctx context.Context, id int64, by *int64) error
}

var _ domaincontractsrepository.Space = (*Space)(nil)

func (s *Space) Create(ctx context.Context, args domaincontractsrepository.CreateSpaceArgs) (id int64, err error) {
	if s.CreateFn == nil {
		panic("unexpected Space.Create call")
	}
	return s.CreateFn(ctx, args)
}

func (s *Space) ReadById(ctx context.Context, id int64) (*domainmodelsentity.Space, error) {
	if s.ReadByIdFn == nil {
		panic("unexpected Space.ReadById call")
	}
	return s.ReadByIdFn(ctx, id)
}

func (s *Space) ReadBySlug(ctx context.Context, value string) (*domainmodelsentity.Space, error) {
	if s.ReadBySlugFn == nil {
		panic("unexpected Space.ReadBySlug call")
	}
	return s.ReadBySlugFn(ctx, value)
}

func (s *Space) ReadByFilter(ctx context.Context, filter domaincontractsrepository.SpaceFilter) (items []domainmodelsentity.Space, total int64, err error) {
	if s.ReadByFilterFn == nil {
		panic("unexpected Space.ReadByFilter call")
	}
	return s.ReadByFilterFn(ctx, filter)
}

func (s *Space) UpdateById(ctx context.Context, id int64, args domaincontractsrepository.UpdateSpaceArgs) error {
	if s.UpdateByIdFn == nil {
		panic("unexpected Space.UpdateById call")
	}
	return s.UpdateByIdFn(ctx, id, args)
}

func (s *Space) DeleteById(ctx context.Context, id int64, by *int64) error {
	if s.DeleteByIdFn == nil {
		panic("unexpected Space.DeleteById call")
	}
	return s.DeleteByIdFn(ctx, id, by)
}

type Permission struct {
	CreateFn       func(ctx context.Context, spaceId int64, args domaincontractsrepository.CreatePermissionArgs) (id int64, err error)
	ReadByIdFn     func(ctx context.Context, spaceId int64, id int64) (*domainmodelsentity.Permission, error)
	ReadBySlugFn   func(ctx context.Context, spaceId int64, value string) (*domainmodelsentity.Permission, error)
	ReadByFilterFn func(ctx context.Context, spaceId int64, filter domaincontractsrepository.PermissionFilter) (items []domainmodelsentity.Permission, total int64, err error)
	UpdateByIdFn   func(ctx context.Context, spaceId int64, id int64, args domaincontractsrepository.UpdatePermissionArgs) error
	DeleteByIdFn   func(ctx context.Context, spaceId int64, id int64, by *int64) error
}

var _ domaincontractsrepository.Permission = (*Permission)(nil)

func (s *Permission) Create(ctx context.Context, spaceId int64, args domaincontractsrepository.CreatePermissionArgs) (id int64, err error) {
	if s.CreateFn == nil {
		panic("unexpected Permission.Create call")
	}
	return s.CreateFn(ctx, spaceId, args)
}

func (s *Permission) ReadById(ctx context.Context, spaceId int64, id int64) (*domainmodelsentity.Permission, error) {
	if s.ReadByIdFn == nil {
		panic("unexpected Permission.ReadById call")
	}
	return s.ReadByIdFn(ctx, spaceId, id)
}

func (s *Permission) ReadBySlug(ctx context.Context, spaceId int64, value string) (*domainmodelsentity.Permission, error) {
	if s.ReadBySlugFn == nil {
		panic("unexpected Permission.ReadBySlug call")
	}
	return s.ReadBySlugFn(ctx, spaceId, value)
}

func (s *Permission) ReadByFilter(ctx context.Context, spaceId int64, filter domaincontractsrepository.PermissionFilter) (items []domainmodelsentity.Permission, total int64, err error) {
	if s.ReadByFilterFn == nil {
		panic("unexpected Permission.ReadByFilter call")
	}
	return s.ReadByFilterFn(ctx, spaceId, filter)
}

func (s *Permission) UpdateById(ctx context.Context, spaceId int64, id int64, args domaincontractsrepository.UpdatePermissionArgs) error {
	if s.UpdateByIdFn == nil {
		panic("unexpected Permission.UpdateById call")
	}
	return s.UpdateByIdFn(ctx, spaceId, id, args)
}

func (s *Permission) DeleteById(ctx context.Context, spaceId int64, id int64, by *int64) error {
	if s.DeleteByIdFn == nil {
		panic("unexpected Permission.DeleteById call")
	}
	return s.DeleteByIdFn(ctx, spaceId, id, by)
}

type Tracer struct {
	GenerateEventIdFn func() (uuid.UUID, error)
	InjectFn          func(ctx context.Context, trace domainmodelsapp.Trace) (newCtx context.Context)
	ExtractFn         func(ctx context.Context) (trace *domainmodelsapp.Trace, err error)
}

var _ domaincontractsutility.Tracer = (*Tracer)(nil)

func (s *Tracer) GenerateEventId() (uuid.UUID, error) {
	if s.GenerateEventIdFn == nil {
		panic("unexpected Tracer.GenerateEventId call")
	}
	return s.GenerateEventIdFn()
}

func (s *Tracer) Inject(ctx context.Context, trace domainmodelsapp.Trace) (newCtx context.Context) {
	if s.InjectFn == nil {
		panic("unexpected Tracer.Inject call")
	}
	return s.InjectFn(ctx, trace)
}

func (s *Tracer) Extract(ctx context.Context) (trace *domainmodelsapp.Trace, err error) {
	if s.ExtractFn == nil {
		panic("unexpected Tracer.Extract call")
	}
	return s.ExtractFn(ctx)
}

type Validator struct {
	SpaceSlugFn      func(val string) error
	SpaceNameFn      func(val string) error
	SpaceDescFn      func(val string) error
	PermissionSlugFn func(val string) error
	PermissionNameFn func(val string) error
	PermissionDescFn func(val string) error
	RoleSlugFn       func(val string) error
	RoleNameFn       func(val string) error
	RoleDescFn       func(val string) error
	UserNameFn       func(val string) error
	UserBioFn        func(val string) error
	UserUsernameFn   func(val string) error
	UserEmailFn      func(val string) error
	UserPhoneFn      func(val string) error
	UserPasswordFn   func(val string) error
	ClientSlugFn     func(val string) error
	ClientNameFn     func(val string) error
	ClientDescFn     func(val string) error
	ClientOriginFn   func(val string) error
}

var _ domaincontractsutility.Validator = (*Validator)(nil)

func (s *Validator) SpaceSlug(val string) error {
	if s.SpaceSlugFn == nil {
		panic("unexpected Validator.SpaceSlug call")
	}
	return s.SpaceSlugFn(val)
}

func (s *Validator) SpaceName(val string) error {
	if s.SpaceNameFn == nil {
		panic("unexpected Validator.SpaceName call")
	}
	return s.SpaceNameFn(val)
}

func (s *Validator) SpaceDesc(val string) error {
	if s.SpaceDescFn == nil {
		panic("unexpected Validator.SpaceDesc call")
	}
	return s.SpaceDescFn(val)
}

func (s *Validator) PermissionSlug(val string) error {
	if s.PermissionSlugFn == nil {
		panic("unexpected Validator.PermissionSlug call")
	}
	return s.PermissionSlugFn(val)
}

func (s *Validator) PermissionName(val string) error {
	if s.PermissionNameFn == nil {
		panic("unexpected Validator.PermissionName call")
	}
	return s.PermissionNameFn(val)
}

func (s *Validator) PermissionDesc(val string) error {
	if s.PermissionDescFn == nil {
		panic("unexpected Validator.PermissionDesc call")
	}
	return s.PermissionDescFn(val)
}

func (s *Validator) RoleSlug(val string) error {
	if s.RoleSlugFn == nil {
		panic("unexpected Validator.RoleSlug call")
	}
	return s.RoleSlugFn(val)
}

func (s *Validator) RoleName(val string) error {
	if s.RoleNameFn == nil {
		panic("unexpected Validator.RoleName call")
	}
	return s.RoleNameFn(val)
}

func (s *Validator) RoleDesc(val string) error {
	if s.RoleDescFn == nil {
		panic("unexpected Validator.RoleDesc call")
	}
	return s.RoleDescFn(val)
}

func (s *Validator) UserName(val string) error {
	if s.UserNameFn == nil {
		panic("unexpected Validator.UserName call")
	}
	return s.UserNameFn(val)
}

func (s *Validator) UserBio(val string) error {
	if s.UserBioFn == nil {
		panic("unexpected Validator.UserBio call")
	}
	return s.UserBioFn(val)
}

func (s *Validator) UserUsername(val string) error {
	if s.UserUsernameFn == nil {
		panic("unexpected Validator.UserUsername call")
	}
	return s.UserUsernameFn(val)
}

func (s *Validator) UserEmail(val string) error {
	if s.UserEmailFn == nil {
		panic("unexpected Validator.UserEmail call")
	}
	return s.UserEmailFn(val)
}

func (s *Validator) UserPhone(val string) error {
	if s.UserPhoneFn == nil {
		panic("unexpected Validator.UserPhone call")
	}
	return s.UserPhoneFn(val)
}

func (s *Validator) UserPassword(val string) error {
	if s.UserPasswordFn == nil {
		panic("unexpected Validator.UserPassword call")
	}
	return s.UserPasswordFn(val)
}

func (s *Validator) ClientSlug(val string) error {
	if s.ClientSlugFn == nil {
		panic("unexpected Validator.ClientSlug call")
	}
	return s.ClientSlugFn(val)
}

func (s *Validator) ClientName(val string) error {
	if s.ClientNameFn == nil {
		panic("unexpected Validator.ClientName call")
	}
	return s.ClientNameFn(val)
}

func (s *Validator) ClientDesc(val string) error {
	if s.ClientDescFn == nil {
		panic("unexpected Validator.ClientDesc call")
	}
	return s.ClientDescFn(val)
}

type Transactor struct {
	WithTxFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

var _ domaincontractsutility.Transactor = (*Transactor)(nil)

func (s *Transactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if s.WithTxFn == nil {
		panic("unexpected Transactor.WithTx call")
	}
	return s.WithTxFn(ctx, fn)
}

type Keygen struct {
	GenerateFn func() (raw string, hash string, redacted string, err error)
	HashFn     func(raw string) string
	ValidateFn func(raw string) error
	CompareFn  func(storedHash string, raw string) error
}

var _ domaincontractsutility.Keygen = (*Keygen)(nil)

func (s *Keygen) Generate() (raw string, hash string, redacted string, err error) {
	if s.GenerateFn == nil {
		panic("unexpected Keygen.Generate call")
	}
	return s.GenerateFn()
}

func (s *Keygen) Hash(raw string) string {
	if s.HashFn == nil {
		panic("unexpected Keygen.Hash call")
	}
	return s.HashFn(raw)
}

func (s *Keygen) Validate(raw string) error {
	if s.ValidateFn == nil {
		panic("unexpected Keygen.Validate call")
	}
	return s.ValidateFn(raw)
}

func (s *Keygen) Compare(storedHash string, raw string) error {
	if s.CompareFn == nil {
		panic("unexpected Keygen.Compare call")
	}
	return s.CompareFn(storedHash, raw)
}

type PermissionGuard struct {
	CheckFn func(claims *domainmodelsentity.AccessClaims, permissions []string) error
}

var _ domaincontractsutility.PermissionGuard = (*PermissionGuard)(nil)

func (s *PermissionGuard) Check(claims *domainmodelsentity.AccessClaims, permissions []string) error {
	if s.CheckFn == nil {
		panic("unexpected PermissionGuard.Check call")
	}
	return s.CheckFn(claims, permissions)
}

type Token struct {
	GenerateAccessFn  func(claims *domainmodelsentity.AccessClaims) (string, error)
	ValidateAccessFn  func(spaceId int64, token string) (*domainmodelsentity.AccessClaims, error)
	GenerateRefreshFn func(claims *domainmodelsentity.RefreshClaims) (string, error)
	ValidateRefreshFn func(spaceId int64, token string) (*domainmodelsentity.RefreshClaims, error)
}

var _ domaincontractsutility.Token = (*Token)(nil)

func (s *Token) GenerateAccess(claims *domainmodelsentity.AccessClaims) (string, error) {
	if s.GenerateAccessFn == nil {
		panic("unexpected Token.GenerateAccess call")
	}
	return s.GenerateAccessFn(claims)
}

func (s *Token) ValidateAccess(spaceId int64, token string) (*domainmodelsentity.AccessClaims, error) {
	if s.ValidateAccessFn == nil {
		panic("unexpected Token.ValidateAccess call")
	}
	return s.ValidateAccessFn(spaceId, token)
}

func (s *Token) GenerateRefresh(claims *domainmodelsentity.RefreshClaims) (string, error) {
	if s.GenerateRefreshFn == nil {
		panic("unexpected Token.GenerateRefresh call")
	}
	return s.GenerateRefreshFn(claims)
}

func (s *Token) ValidateRefresh(spaceId int64, token string) (*domainmodelsentity.RefreshClaims, error) {
	if s.ValidateRefreshFn == nil {
		panic("unexpected Token.ValidateRefresh call")
	}
	return s.ValidateRefreshFn(spaceId, token)
}

type Password struct {
	HashFn    func(password string) (hashed string, err error)
	CompareFn func(storedHash string, password string) (err error)
}

var _ domaincontractsutility.Password = (*Password)(nil)

func (s *Password) Hash(password string) (hashed string, err error) {
	if s.HashFn == nil {
		panic("unexpected Password.Hash call")
	}
	return s.HashFn(password)
}

func (s *Password) Compare(storedHash string, password string) (err error) {
	if s.CompareFn == nil {
		panic("unexpected Password.Compare call")
	}
	return s.CompareFn(storedHash, password)
}

type GoogleOAuth struct {
	BeginFn    func(ctx context.Context, args domaincontractsutility.BeginGoogleOAuthArgs) (authorizationUrl string, err error)
	CompleteFn func(ctx context.Context, args domaincontractsutility.CompleteGoogleOAuthArgs) (*domaincontractsutility.CompleteGoogleOAuthResult, error)
}

var _ domaincontractsutility.GoogleOAuth = (*GoogleOAuth)(nil)

func (s *GoogleOAuth) Begin(ctx context.Context, args domaincontractsutility.BeginGoogleOAuthArgs) (authorizationUrl string, err error) {
	if s.BeginFn == nil {
		panic("unexpected GoogleOAuth.Begin call")
	}
	return s.BeginFn(ctx, args)
}

func (s *GoogleOAuth) Complete(ctx context.Context, args domaincontractsutility.CompleteGoogleOAuthArgs) (*domaincontractsutility.CompleteGoogleOAuthResult, error) {
	if s.CompleteFn == nil {
		panic("unexpected GoogleOAuth.Complete call")
	}
	return s.CompleteFn(ctx, args)
}

type LogEntry struct {
	Level, Tag, Message string
	Meta                domainmodelslogger.Meta
}
type Logger struct{ Entries []LogEntry }

func (s *Logger) record(level, tag, message string, meta domainmodelslogger.Meta) {
	copied := domainmodelslogger.Meta{}
	for k, v := range meta {
		copied[k] = v
	}
	s.Entries = append(s.Entries, LogEntry{Level: level, Tag: tag, Message: message, Meta: copied})
}
func (s *Logger) Error(ctx context.Context, tag, message string, meta domainmodelslogger.Meta) {
	s.record("error", tag, message, meta)
}
func (s *Logger) Warn(ctx context.Context, tag, message string, meta domainmodelslogger.Meta) {
	s.record("warn", tag, message, meta)
}
func (s *Logger) Info(ctx context.Context, tag, message string, meta domainmodelslogger.Meta) {
	s.record("info", tag, message, meta)
}
func (s *Logger) Debug(ctx context.Context, tag, message string, meta domainmodelslogger.Meta) {
	s.record("debug", tag, message, meta)
}

func (s *Validator) ClientOrigin(val string) error {
	if s.ClientOriginFn == nil {
		panic("unexpected Validator.ClientOrigin call")
	}

	return s.ClientOriginFn(val)
}
