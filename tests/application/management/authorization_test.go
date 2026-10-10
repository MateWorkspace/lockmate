package management_test

import (
	"context"
	"errors"
	applicationmanagementclassicauth "github.com/MateWorkspace/lockmate/internal/application/management/classic_auth"
	applicationmanagementclient "github.com/MateWorkspace/lockmate/internal/application/management/client"
	applicationmanagementgoogleauth "github.com/MateWorkspace/lockmate/internal/application/management/google_auth"
	applicationmanagementpermission "github.com/MateWorkspace/lockmate/internal/application/management/permission"
	applicationmanagementrole "github.com/MateWorkspace/lockmate/internal/application/management/role"
	applicationmanagementrolepermission "github.com/MateWorkspace/lockmate/internal/application/management/role_permission"
	applicationmanagementspace "github.com/MateWorkspace/lockmate/internal/application/management/space"
	applicationmanagementspacemember "github.com/MateWorkspace/lockmate/internal/application/management/space_member"
	applicationmanagementspacememberrole "github.com/MateWorkspace/lockmate/internal/application/management/space_member_role"
	applicationmanagementuser "github.com/MateWorkspace/lockmate/internal/application/management/user"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
	"strings"
	"testing"
)

func TestEveryProtectedOperation(t *testing.T) {
	cases := []struct {
		name, permission string
		scoped           bool
		call             func(*testshared.Fixture) error
	}{
		{name: "space_member_role/Create", permission: "space_member_role:add", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementspacememberrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Logger, f.Transactor, "").Create(context.Background(), f.Caller, 10, domainusecasesmanagement.CreateSpaceMemberRoleRequest{})
			return err
		}},
		{name: "space_member_role/ReadById", permission: "space_member_role:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementspacememberrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Logger, f.Transactor, "").ReadById(context.Background(), f.Caller, 10, 40)
			return err
		}},
		{name: "space_member_role/ReadBySpaceMemberIdAndRoleId", permission: "space_member_role:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementspacememberrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Logger, f.Transactor, "").ReadBySpaceMemberIdAndRoleId(context.Background(), f.Caller, 10, 1, 1)
			return err
		}},
		{name: "space_member_role/ReadBySpaceMemberId", permission: "space_member_role:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementspacememberrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Logger, f.Transactor, "").ReadBySpaceMemberId(context.Background(), f.Caller, 10, 1)
			return err
		}},
		{name: "space_member_role/ReadByRoleId", permission: "space_member_role:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementspacememberrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Logger, f.Transactor, "").ReadByRoleId(context.Background(), f.Caller, 10, 1)
			return err
		}},
		{name: "space_member_role/DeleteById", permission: "space_member_role:remove", scoped: true, call: func(f *testshared.Fixture) error {
			err := applicationmanagementspacememberrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Logger, f.Transactor, "").DeleteById(context.Background(), f.Caller, 10, 40)
			return err
		}},
		{name: "space_member_role/DeleteBySpaceMemberIdOrRoleId", permission: "space_member_role:remove", scoped: true, call: func(f *testshared.Fixture) error {
			err := applicationmanagementspacememberrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Logger, f.Transactor, "").DeleteBySpaceMemberIdOrRoleId(context.Background(), f.Caller, 10, domainusecasesmanagement.DeleteSpaceMemberRoleRequest{})
			return err
		}},
		{name: "user/Create", permission: "user:add", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementuser.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.Guard, f.Validator, f.Logger, f.Transactor, f.Password, "").Create(context.Background(), f.Caller, domainusecasesmanagement.CreateUserRequest{})
			return err
		}},
		{name: "user/ReadById", permission: "user:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementuser.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.Guard, f.Validator, f.Logger, f.Transactor, f.Password, "").ReadById(context.Background(), f.Caller, 40)
			return err
		}},
		{name: "user/ReadByFilter", permission: "user:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementuser.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.Guard, f.Validator, f.Logger, f.Transactor, f.Password, "").ReadByFilter(context.Background(), f.Caller, domainusecasesmanagement.FilterUserRequest{})
			return err
		}},
		{name: "user/UpdateById", permission: "user:set", scoped: false, call: func(f *testshared.Fixture) error {
			err := applicationmanagementuser.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.Guard, f.Validator, f.Logger, f.Transactor, f.Password, "").UpdateById(context.Background(), f.Caller, 40, domainusecasesmanagement.UpdateUserRequest{})
			return err
		}},
		{name: "user/DeleteById", permission: "user:remove", scoped: false, call: func(f *testshared.Fixture) error {
			err := applicationmanagementuser.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.Guard, f.Validator, f.Logger, f.Transactor, f.Password, "").DeleteById(context.Background(), f.Caller, 40)
			return err
		}},
		{name: "role_permission/Create", permission: "role_permission:add", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementrolepermission.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.RolePermission, f.Guard, f.Logger, f.Transactor, "").Create(context.Background(), f.Caller, 10, domainusecasesmanagement.CreateRolePermissionRequest{})
			return err
		}},
		{name: "role_permission/ReadById", permission: "role_permission:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementrolepermission.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.RolePermission, f.Guard, f.Logger, f.Transactor, "").ReadById(context.Background(), f.Caller, 10, 40)
			return err
		}},
		{name: "role_permission/ReadByRoleIdAndPermissionId", permission: "role_permission:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementrolepermission.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.RolePermission, f.Guard, f.Logger, f.Transactor, "").ReadByRoleIdAndPermissionId(context.Background(), f.Caller, 10, 1, 1)
			return err
		}},
		{name: "role_permission/ReadByRoleId", permission: "role_permission:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementrolepermission.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.RolePermission, f.Guard, f.Logger, f.Transactor, "").ReadByRoleId(context.Background(), f.Caller, 10, 1)
			return err
		}},
		{name: "role_permission/ReadByPermissionId", permission: "role_permission:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementrolepermission.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.RolePermission, f.Guard, f.Logger, f.Transactor, "").ReadByPermissionId(context.Background(), f.Caller, 10, 1)
			return err
		}},
		{name: "role_permission/DeleteById", permission: "role_permission:remove", scoped: true, call: func(f *testshared.Fixture) error {
			err := applicationmanagementrolepermission.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.RolePermission, f.Guard, f.Logger, f.Transactor, "").DeleteById(context.Background(), f.Caller, 10, 40)
			return err
		}},
		{name: "role_permission/DeleteByRoleIdOrPermissionId", permission: "role_permission:remove", scoped: true, call: func(f *testshared.Fixture) error {
			err := applicationmanagementrolepermission.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.RolePermission, f.Guard, f.Logger, f.Transactor, "").DeleteByRoleIdOrPermissionId(context.Background(), f.Caller, 10, domainusecasesmanagement.DeleteRolePermissionRequest{})
			return err
		}},
		{name: "space_member/Create", permission: "space_member:add", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementspacemember.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Logger, f.Transactor, "").Create(context.Background(), f.Caller, 10, domainusecasesmanagement.CreateSpaceMemberRequest{})
			return err
		}},
		{name: "space_member/ReadById", permission: "space_member:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementspacemember.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Logger, f.Transactor, "").ReadById(context.Background(), f.Caller, 10, 40)
			return err
		}},
		{name: "space_member/ReadByUserId", permission: "space_member:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementspacemember.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Logger, f.Transactor, "").ReadByUserId(context.Background(), f.Caller, 10, 1)
			return err
		}},
		{name: "space_member/ReadByFilter", permission: "space_member:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementspacemember.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Logger, f.Transactor, "").ReadByFilter(context.Background(), f.Caller, 10, domainusecasesmanagement.FilterSpaceMemberRequest{})
			return err
		}},
		{name: "space_member/UpdateById", permission: "space_member:set", scoped: true, call: func(f *testshared.Fixture) error {
			err := applicationmanagementspacemember.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Logger, f.Transactor, "").UpdateById(context.Background(), f.Caller, 10, 40, domainusecasesmanagement.UpdateSpaceMemberRequest{})
			return err
		}},
		{name: "space_member/ReadEffectiveRolesBySpaceMemberId", permission: "user_permission:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementspacemember.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Logger, f.Transactor, "").ReadEffectiveRolesBySpaceMemberId(context.Background(), f.Caller, 10, 1)
			return err
		}},
		{name: "space_member/ReadEffectivePermissionsBySpaceMemberId", permission: "user_permission:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementspacemember.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Logger, f.Transactor, "").ReadEffectivePermissionsBySpaceMemberId(context.Background(), f.Caller, 10, 1)
			return err
		}},
		{name: "space_member/DeleteById", permission: "space_member:remove", scoped: true, call: func(f *testshared.Fixture) error {
			err := applicationmanagementspacemember.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Logger, f.Transactor, "").DeleteById(context.Background(), f.Caller, 10, 40)
			return err
		}},
		{name: "role/Create", permission: "role:add", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Role, f.Guard, f.Validator, f.Logger, f.Transactor, "").Create(context.Background(), f.Caller, 10, domainusecasesmanagement.CreateRoleRequest{})
			return err
		}},
		{name: "role/ReadById", permission: "role:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Role, f.Guard, f.Validator, f.Logger, f.Transactor, "").ReadById(context.Background(), f.Caller, 10, 40)
			return err
		}},
		{name: "role/ReadBySlug", permission: "role:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Role, f.Guard, f.Validator, f.Logger, f.Transactor, "").ReadBySlug(context.Background(), f.Caller, 10, "user")
			return err
		}},
		{name: "role/ReadByFilter", permission: "role:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Role, f.Guard, f.Validator, f.Logger, f.Transactor, "").ReadByFilter(context.Background(), f.Caller, 10, domainusecasesmanagement.FilterRoleRequest{})
			return err
		}},
		{name: "role/ReadDefault", permission: "role:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Role, f.Guard, f.Validator, f.Logger, f.Transactor, "").ReadDefault(context.Background(), f.Caller, 10)
			return err
		}},
		{name: "role/UpdateById", permission: "role:set", scoped: true, call: func(f *testshared.Fixture) error {
			err := applicationmanagementrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Role, f.Guard, f.Validator, f.Logger, f.Transactor, "").UpdateById(context.Background(), f.Caller, 10, 40, domainusecasesmanagement.UpdateRoleRequest{})
			return err
		}},
		{name: "role/SetDefault", permission: "role:set", scoped: true, call: func(f *testshared.Fixture) error {
			err := applicationmanagementrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Role, f.Guard, f.Validator, f.Logger, f.Transactor, "").SetDefault(context.Background(), f.Caller, 10, 40)
			return err
		}},
		{name: "role/DeleteById", permission: "role:remove", scoped: true, call: func(f *testshared.Fixture) error {
			err := applicationmanagementrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Role, f.Guard, f.Validator, f.Logger, f.Transactor, "").DeleteById(context.Background(), f.Caller, 10, 40)
			return err
		}},
		{name: "google_auth/ReadById", permission: "google_auth:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementgoogleauth.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.GoogleAuth, f.Guard, f.Logger, "").ReadById(context.Background(), f.Caller, 40)
			return err
		}},
		{name: "google_auth/ReadByUserId", permission: "google_auth:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementgoogleauth.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.GoogleAuth, f.Guard, f.Logger, "").ReadByUserId(context.Background(), f.Caller, 1)
			return err
		}},
		{name: "google_auth/ReadBySubject", permission: "google_auth:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementgoogleauth.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.GoogleAuth, f.Guard, f.Logger, "").ReadBySubject(context.Background(), f.Caller, "user")
			return err
		}},
		{name: "google_auth/ReadByFilter", permission: "google_auth:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementgoogleauth.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.GoogleAuth, f.Guard, f.Logger, "").ReadByFilter(context.Background(), f.Caller, domainusecasesmanagement.FilterGoogleAuthRequest{})
			return err
		}},
		{name: "classic_auth/Create", permission: "classic_auth:add", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementclassicauth.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.GoogleAuth, f.Guard, f.Validator, f.Logger, f.Transactor, f.Password, "").Create(context.Background(), f.Caller, domainusecasesmanagement.CreateClassicAuthRequest{})
			return err
		}},
		{name: "classic_auth/ReadById", permission: "classic_auth:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementclassicauth.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.GoogleAuth, f.Guard, f.Validator, f.Logger, f.Transactor, f.Password, "").ReadById(context.Background(), f.Caller, 40)
			return err
		}},
		{name: "classic_auth/ReadByUserId", permission: "classic_auth:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementclassicauth.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.GoogleAuth, f.Guard, f.Validator, f.Logger, f.Transactor, f.Password, "").ReadByUserId(context.Background(), f.Caller, 1)
			return err
		}},
		{name: "classic_auth/ReadByUsername", permission: "classic_auth:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementclassicauth.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.GoogleAuth, f.Guard, f.Validator, f.Logger, f.Transactor, f.Password, "").ReadByUsername(context.Background(), f.Caller, "user")
			return err
		}},
		{name: "classic_auth/ReadByEmail", permission: "classic_auth:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementclassicauth.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.GoogleAuth, f.Guard, f.Validator, f.Logger, f.Transactor, f.Password, "").ReadByEmail(context.Background(), f.Caller, "user")
			return err
		}},
		{name: "classic_auth/ReadByPhone", permission: "classic_auth:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementclassicauth.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.GoogleAuth, f.Guard, f.Validator, f.Logger, f.Transactor, f.Password, "").ReadByPhone(context.Background(), f.Caller, "user")
			return err
		}},
		{name: "classic_auth/ReadByFilter", permission: "classic_auth:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementclassicauth.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.GoogleAuth, f.Guard, f.Validator, f.Logger, f.Transactor, f.Password, "").ReadByFilter(context.Background(), f.Caller, domainusecasesmanagement.FilterClassicAuthRequest{})
			return err
		}},
		{name: "classic_auth/UpdateById", permission: "classic_auth:set", scoped: false, call: func(f *testshared.Fixture) error {
			err := applicationmanagementclassicauth.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.GoogleAuth, f.Guard, f.Validator, f.Logger, f.Transactor, f.Password, "").UpdateById(context.Background(), f.Caller, 40, domainusecasesmanagement.UpdateClassicAuthRequest{})
			return err
		}},
		{name: "classic_auth/ResetPassword", permission: "user_password:set", scoped: false, call: func(f *testshared.Fixture) error {
			err := applicationmanagementclassicauth.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.GoogleAuth, f.Guard, f.Validator, f.Logger, f.Transactor, f.Password, "").ResetPassword(context.Background(), f.Caller, 1, domainusecasesmanagement.ResetUserPasswordRequest{})
			return err
		}},
		{name: "classic_auth/DeleteById", permission: "classic_auth:remove", scoped: false, call: func(f *testshared.Fixture) error {
			err := applicationmanagementclassicauth.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.ClassicAuth, f.GoogleAuth, f.Guard, f.Validator, f.Logger, f.Transactor, f.Password, "").DeleteById(context.Background(), f.Caller, 40)
			return err
		}},
		{name: "client/Create", permission: "client:add", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementclient.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Client, f.ClientOrigin, f.Guard, f.Validator, f.Logger, f.Transactor, f.Keygen, "").Create(context.Background(), f.Caller, 10, domainusecasesmanagement.CreateClientRequest{})
			return err
		}},
		{name: "client/ReadById", permission: "client:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementclient.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Client, f.ClientOrigin, f.Guard, f.Validator, f.Logger, f.Transactor, f.Keygen, "").ReadById(context.Background(), f.Caller, 10, 40)
			return err
		}},
		{name: "client/ReadBySlug", permission: "client:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementclient.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Client, f.ClientOrigin, f.Guard, f.Validator, f.Logger, f.Transactor, f.Keygen, "").ReadBySlug(context.Background(), f.Caller, 10, "user")
			return err
		}},
		{name: "client/ReadByFilter", permission: "client:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementclient.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Client, f.ClientOrigin, f.Guard, f.Validator, f.Logger, f.Transactor, f.Keygen, "").ReadByFilter(context.Background(), f.Caller, 10, domainusecasesmanagement.FilterClientRequest{})
			return err
		}},
		{name: "client/UpdateById", permission: "client:set", scoped: true, call: func(f *testshared.Fixture) error {
			err := applicationmanagementclient.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Client, f.ClientOrigin, f.Guard, f.Validator, f.Logger, f.Transactor, f.Keygen, "").UpdateById(context.Background(), f.Caller, 10, 40, domainusecasesmanagement.UpdateClientRequest{})
			return err
		}},
		{name: "client/RotateSecret", permission: "client_secret:set", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementclient.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Client, f.ClientOrigin, f.Guard, f.Validator, f.Logger, f.Transactor, f.Keygen, "").RotateSecret(context.Background(), f.Caller, 10, 40)
			return err
		}},
		{name: "client/DeleteById", permission: "client:remove", scoped: true, call: func(f *testshared.Fixture) error {
			err := applicationmanagementclient.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Client, f.ClientOrigin, f.Guard, f.Validator, f.Logger, f.Transactor, f.Keygen, "").DeleteById(context.Background(), f.Caller, 10, 40)
			return err
		}},
		{name: "space/Create", permission: "space:add", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementspace.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Guard, f.Validator, f.Logger, f.Transactor, "").Create(context.Background(), f.Caller, domainusecasesmanagement.CreateSpaceRequest{})
			return err
		}},
		{name: "space/ReadById", permission: "space:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementspace.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Guard, f.Validator, f.Logger, f.Transactor, "").ReadById(context.Background(), f.Caller, 40)
			return err
		}},
		{name: "space/ReadBySlug", permission: "space:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementspace.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Guard, f.Validator, f.Logger, f.Transactor, "").ReadBySlug(context.Background(), f.Caller, "user")
			return err
		}},
		{name: "space/ReadByFilter", permission: "space:get", scoped: false, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementspace.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Guard, f.Validator, f.Logger, f.Transactor, "").ReadByFilter(context.Background(), f.Caller, domainusecasesmanagement.FilterSpaceRequest{})
			return err
		}},
		{name: "space/UpdateById", permission: "space:set", scoped: false, call: func(f *testshared.Fixture) error {
			err := applicationmanagementspace.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Guard, f.Validator, f.Logger, f.Transactor, "").UpdateById(context.Background(), f.Caller, 40, domainusecasesmanagement.UpdateSpaceRequest{})
			return err
		}},
		{name: "space/DeleteById", permission: "space:remove", scoped: false, call: func(f *testshared.Fixture) error {
			err := applicationmanagementspace.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Guard, f.Validator, f.Logger, f.Transactor, "").DeleteById(context.Background(), f.Caller, 40)
			return err
		}},
		{name: "permission/Create", permission: "permission:add", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementpermission.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Permission, f.Guard, f.Validator, f.Logger, f.Transactor, "").Create(context.Background(), f.Caller, 10, domainusecasesmanagement.CreatePermissionRequest{})
			return err
		}},
		{name: "permission/ReadById", permission: "permission:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementpermission.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Permission, f.Guard, f.Validator, f.Logger, f.Transactor, "").ReadById(context.Background(), f.Caller, 10, 40)
			return err
		}},
		{name: "permission/ReadBySlug", permission: "permission:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementpermission.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Permission, f.Guard, f.Validator, f.Logger, f.Transactor, "").ReadBySlug(context.Background(), f.Caller, 10, "user")
			return err
		}},
		{name: "permission/ReadByFilter", permission: "permission:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := applicationmanagementpermission.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Permission, f.Guard, f.Validator, f.Logger, f.Transactor, "").ReadByFilter(context.Background(), f.Caller, 10, domainusecasesmanagement.FilterPermissionRequest{})
			return err
		}},
		{name: "permission/UpdateById", permission: "permission:set", scoped: true, call: func(f *testshared.Fixture) error {
			err := applicationmanagementpermission.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Permission, f.Guard, f.Validator, f.Logger, f.Transactor, "").UpdateById(context.Background(), f.Caller, 10, 40, domainusecasesmanagement.UpdatePermissionRequest{})
			return err
		}},
		{name: "permission/DeleteById", permission: "permission:remove", scoped: true, call: func(f *testshared.Fixture) error {
			err := applicationmanagementpermission.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Permission, f.Guard, f.Validator, f.Logger, f.Transactor, "").DeleteById(context.Background(), f.Caller, 10, 40)
			return err
		}},

		{name: "client/CreateOrigin", permission: "client_origin:add", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := client(f).CreateOrigin(context.Background(), f.Caller, 10, 30, domainusecasesmanagement.CreateClientOriginRequest{})
			return err
		}},
		{name: "client/ReadOriginById", permission: "client_origin:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := client(f).ReadOriginById(context.Background(), f.Caller, 10, 30, 50)
			return err
		}},
		{name: "client/ReadOriginByValue", permission: "client_origin:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := client(f).ReadOriginByValue(context.Background(), f.Caller, 10, 30, "https://app.example.com")
			return err
		}},
		{name: "client/ReadOriginsByClientId", permission: "client_origin:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := client(f).ReadOriginsByClientId(context.Background(), f.Caller, 10, 30)
			return err
		}},
		{name: "client/ReadOriginsByFilter", permission: "client_origin:get", scoped: true, call: func(f *testshared.Fixture) error {
			_, err := client(f).ReadOriginsByFilter(context.Background(), f.Caller, 10, 30, domainusecasesmanagement.FilterClientOriginRequest{})
			return err
		}},
		{name: "client/UpdateOriginById", permission: "client_origin:set", scoped: true, call: func(f *testshared.Fixture) error {
			err := client(f).UpdateOriginById(context.Background(), f.Caller, 10, 30, 50, domainusecasesmanagement.UpdateClientOriginRequest{})
			return err
		}},
		{name: "client/DeleteOriginById", permission: "client_origin:remove", scoped: true, call: func(f *testshared.Fixture) error {
			err := client(f).DeleteOriginById(context.Background(), f.Caller, 10, 30, 50)
			return err
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Run("permission denied", func(t *testing.T) {
				f := testshared.NewFixture()
				f.User.ReadByIdFn = nil
				f.Space.ReadByIdFn = nil
				f.SpaceMember.ReadByIdFn = nil
				f.Transactor.WithTxFn = nil
				denied := domainmodelserror.NewError("do not log secret-content", domainmodelserror.Forbidden, errors.New("private-secret"))
				called := false
				f.Guard.CheckFn = func(caller *domainmodelsentity.AccessClaims, permissions []string) error {
					called = true
					if caller != f.Caller || len(permissions) != 1 || permissions[0] != test.permission {
						t.Fatalf("wrong guard arguments: %v", permissions)
					}
					return denied
				}
				err := test.call(f)
				if err != denied || !called {
					t.Fatalf("permission failure changed or guard skipped: %v", err)
				}
				if len(f.Logger.Entries) != 1 || f.Logger.Entries[0].Level != "error" {
					t.Fatalf("expected one error log: %v", f.Logger.Entries)
				}
				if f.Logger.Entries[0].Meta["err"] != denied || !strings.HasPrefix(f.Logger.Entries[0].Message, "permission denied to ") {
					t.Fatal("permission failure lacks specific message or original error")
				}
				if f.Transactions != 0 {
					t.Fatal("permission denial started a transaction")
				}
			})
			t.Run("nil caller", func(t *testing.T) {
				f := testshared.NewFixture()
				f.Caller = nil
				err := test.call(f)
				testshared.RequireError(t, err, domainmodelserror.BadArgs)
				if len(f.Logger.Entries) != 1 || f.Logger.Entries[0].Level != "error" {
					t.Fatal("missing single error log")
				}
			})
			t.Run("wrong space", func(t *testing.T) {
				f := testshared.NewFixture()
				f.Caller.SpaceId = 11
				f.Members[20].SpaceId = 11
				checked := false
				f.Guard.CheckFn = func(*domainmodelsentity.AccessClaims, []string) error {
					checked = true
					return nil
				}
				testshared.RequireError(t, test.call(f), domainmodelserror.Forbidden)
				if !checked {
					t.Fatal("permission check did not precede space authorization")
				}
			})

		})
	}
}
