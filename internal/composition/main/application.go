package compositionmain

import (
	applicationauthclient "github.com/MateWorkspace/lockmate/internal/application/auth/client"
	applicationauthgoogle "github.com/MateWorkspace/lockmate/internal/application/auth/google"
	applicationauthhandoff "github.com/MateWorkspace/lockmate/internal/application/auth/handoff"
	applicationauthsession "github.com/MateWorkspace/lockmate/internal/application/auth/session"
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
	applicationmiscallowedorigin "github.com/MateWorkspace/lockmate/internal/application/misc/allowed_origin"
	applicationmiscauthflow "github.com/MateWorkspace/lockmate/internal/application/misc/auth_flow"
	applicationmisctracer "github.com/MateWorkspace/lockmate/internal/application/misc/tracer"
	applicationprofileaccount "github.com/MateWorkspace/lockmate/internal/application/profile/account"
	applicationprofilesecurity "github.com/MateWorkspace/lockmate/internal/application/profile/security"
	"github.com/MateWorkspace/lockmate/internal/config"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
	domainusecasesprofile "github.com/MateWorkspace/lockmate/internal/domain/usecases/profile"
)

type application struct {
	authGoogle                domainusecasesauth.Google
	authHandoff               domainusecasesauth.Handoff
	miscAuthFlow              domainusecasesmisc.AuthFlow
	authClient                domainusecasesauth.Client
	authSession               domainusecasesauth.Session
	profileAccount            domainusecasesprofile.Account
	profileSecurity           domainusecasesprofile.Security
	managementSpace           domainusecasesmanagement.Space
	managementPermission      domainusecasesmanagement.Permission
	managementRole            domainusecasesmanagement.Role
	managementUser            domainusecasesmanagement.User
	managementClassicAuth     domainusecasesmanagement.ClassicAuth
	managementGoogleAuth      domainusecasesmanagement.GoogleAuth
	managementClient          domainusecasesmanagement.Client
	managementSpaceMember     domainusecasesmanagement.SpaceMember
	managementRolePermission  domainusecasesmanagement.RolePermission
	managementSpaceMemberRole domainusecasesmanagement.SpaceMemberRole
	miscTracer                domainusecasesmisc.Tracer
	miscAllowedOrigin         domainusecasesmisc.AllowedOrigin
}

func (l *launcher) newApplication() {
	l.app = &application{
		authGoogle: applicationauthgoogle.NewUsecaseImpl(
			l.infra.client,
			l.infra.space,
			l.infra.user,
			l.infra.spaceMember,
			l.infra.spaceMemberRole,
			l.infra.validator,
			l.infra.transactor,
			l.infra.logger,
			l.infra.googleAuth,
			l.infra.googleOAuth,
			l.infra.googleLaunch,
			l.infra.authHandoff,
			l.infra.authFlow,
			applicationauthgoogle.Config{
				Enabled:         config.Env.GoogleOAuthEnabled,
				PublicUrl:       config.Env.HttpPublicUrl,
				LaunchDuration:  config.Env.GoogleOAuthLaunchDuration,
				AttemptDuration: config.Env.GoogleOAuthAttemptDuration,
				HandoffDuration: config.Env.AuthHandoffDuration,
			},
		),
		authHandoff: applicationauthhandoff.NewUsecaseImpl(
			l.infra.authHandoff,
			l.infra.client,
			l.infra.space,
			l.infra.user,
			l.infra.spaceMember,
			l.infra.spaceMemberRole,
			l.infra.authFlow,
			l.infra.token,
			l.infra.transactor,
			l.infra.logger,
		),
		miscAuthFlow: applicationmiscauthflow.NewUsecaseImpl(
			l.infra.googleLaunch,
			l.infra.googleAttempt,
			l.infra.authHandoff,
			l.infra.logger,
		),
		authClient: applicationauthclient.NewUsecaseImpl(
			l.infra.client,
			l.infra.space,
			l.infra.keygen,
			l.infra.logger,
		),
		authSession: applicationauthsession.NewUsecaseImpl(
			l.infra.client,
			l.infra.space,
			l.infra.user,
			l.infra.classicAuth,
			l.infra.spaceMember,
			l.infra.spaceMemberRole,
			l.infra.password,
			l.infra.token,
			l.infra.validator,
			l.infra.transactor,
			l.infra.logger,
		),
		profileAccount: applicationprofileaccount.NewUsecaseImpl(
			l.infra.user,
			l.infra.space,
			l.infra.spaceMember,
			l.infra.spaceMemberRole,
			l.infra.permissionGuard,
			l.infra.validator,
			l.infra.transactor,
			l.infra.logger,
		),
		profileSecurity: applicationprofilesecurity.NewUsecaseImpl(
			l.infra.user,
			l.infra.space,
			l.infra.spaceMember,
			l.infra.classicAuth,
			l.infra.googleAuth,
			l.infra.password,
			l.infra.permissionGuard,
			l.infra.validator,
			l.infra.transactor,
			l.infra.logger,
		),
		managementSpace: applicationmanagementspace.NewUsecaseImpl(
			l.infra.user,
			l.infra.space,
			l.infra.spaceMember,
			l.infra.permissionGuard,
			l.infra.validator,
			l.infra.logger,
			l.infra.transactor,
			config.Env.ControlSpaceSlug,
		),
		managementPermission: applicationmanagementpermission.NewUsecaseImpl(
			l.infra.user,
			l.infra.space,
			l.infra.spaceMember,
			l.infra.permission,
			l.infra.permissionGuard,
			l.infra.validator,
			l.infra.logger,
			l.infra.transactor,
			config.Env.ControlSpaceSlug,
		),
		managementRole: applicationmanagementrole.NewUsecaseImpl(
			l.infra.user,
			l.infra.space,
			l.infra.spaceMember,
			l.infra.role,
			l.infra.permissionGuard,
			l.infra.validator,
			l.infra.logger,
			l.infra.transactor,
			config.Env.ControlSpaceSlug,
		),
		managementUser: applicationmanagementuser.NewUsecaseImpl(
			l.infra.user,
			l.infra.space,
			l.infra.spaceMember,
			l.infra.classicAuth,
			l.infra.permissionGuard,
			l.infra.validator,
			l.infra.logger,
			l.infra.transactor,
			l.infra.password,
			config.Env.ControlSpaceSlug,
		),
		managementClassicAuth: applicationmanagementclassicauth.NewUsecaseImpl(
			l.infra.user,
			l.infra.space,
			l.infra.spaceMember,
			l.infra.classicAuth,
			l.infra.googleAuth,
			l.infra.permissionGuard,
			l.infra.validator,
			l.infra.logger,
			l.infra.transactor,
			l.infra.password,
			config.Env.ControlSpaceSlug,
		),
		managementGoogleAuth: applicationmanagementgoogleauth.NewUsecaseImpl(
			l.infra.user,
			l.infra.space,
			l.infra.spaceMember,
			l.infra.googleAuth,
			l.infra.permissionGuard,
			l.infra.logger,
			config.Env.ControlSpaceSlug,
		),
		managementClient: applicationmanagementclient.NewUsecaseImpl(
			l.infra.user,
			l.infra.space,
			l.infra.spaceMember,
			l.infra.client,
			l.infra.clientOrigin,
			l.infra.permissionGuard,
			l.infra.validator,
			l.infra.logger,
			l.infra.transactor,
			l.infra.keygen,
			config.Env.ControlSpaceSlug,
		),
		managementSpaceMember: applicationmanagementspacemember.NewUsecaseImpl(
			l.infra.user,
			l.infra.space,
			l.infra.spaceMember,
			l.infra.spaceMemberRole,
			l.infra.permissionGuard,
			l.infra.logger,
			l.infra.transactor,
			config.Env.ControlSpaceSlug,
		),
		managementRolePermission: applicationmanagementrolepermission.NewUsecaseImpl(
			l.infra.user,
			l.infra.space,
			l.infra.spaceMember,
			l.infra.rolePermission,
			l.infra.permissionGuard,
			l.infra.logger,
			l.infra.transactor,
			config.Env.ControlSpaceSlug,
		),
		managementSpaceMemberRole: applicationmanagementspacememberrole.NewUsecaseImpl(
			l.infra.user,
			l.infra.space,
			l.infra.spaceMember,
			l.infra.spaceMemberRole,
			l.infra.permissionGuard,
			l.infra.logger,
			l.infra.transactor,
			config.Env.ControlSpaceSlug,
		),
		miscTracer: applicationmisctracer.NewUsecaseImpl(
			l.infra.tracer,
			l.infra.logger,
		),
		miscAllowedOrigin: applicationmiscallowedorigin.NewUsecaseImpl(
			l.infra.clientOrigin,
			l.infra.validator,
			l.infra.logger,
		),
	}
}
