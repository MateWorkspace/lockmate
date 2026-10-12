package compositionmain

import (
	seederdata "github.com/MateWorkspace/lockmate/database/seeder"
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
	applicationmiscjwks "github.com/MateWorkspace/lockmate/internal/application/misc/jwks"
	applicationmiscsigningkey "github.com/MateWorkspace/lockmate/internal/application/misc/signing_key"
	applicationmisctracer "github.com/MateWorkspace/lockmate/internal/application/misc/tracer"
	applicationprofileaccount "github.com/MateWorkspace/lockmate/internal/application/profile/account"
	applicationprofilesecurity "github.com/MateWorkspace/lockmate/internal/application/profile/security"
	applicationseeder "github.com/MateWorkspace/lockmate/internal/application/seeder"
	"github.com/MateWorkspace/lockmate/internal/config"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
	domainusecasesprofile "github.com/MateWorkspace/lockmate/internal/domain/usecases/profile"
	domainusecasesseeder "github.com/MateWorkspace/lockmate/internal/domain/usecases/seeder"
)

type application struct {
	seeder                    domainusecasesseeder.Seeder
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
	miscSigningKey            domainusecasesmisc.SigningKey
	miscJwks                  domainusecasesmisc.Jwks
}

func (l *launcher) newApplication() {
	seedData := newSeedData()

	l.app = &application{
		seeder: applicationseeder.NewUsecaseImpl(
			l.infra.space,
			l.infra.permission,
			l.infra.role,
			l.infra.rolePermission,
			l.infra.user,
			l.infra.classicAuth,
			l.infra.spaceMember,
			l.infra.spaceMemberRole,
			l.infra.client,
			l.infra.clientOrigin,
			l.infra.password,
			l.infra.keygen,
			l.infra.validator,
			l.infra.transactor,
			l.infra.logger,
			&seedData,
		),
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
			l.infra.signingKey,
			l.infra.authFlow,
			l.infra.token,
			l.infra.secret,
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
			l.infra.signingKey,
			l.infra.password,
			l.infra.token,
			l.infra.secret,
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
			seederdata.SeedData.Space.Slug,
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
			seederdata.SeedData.Space.Slug,
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
			seederdata.SeedData.Space.Slug,
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
			seederdata.SeedData.Space.Slug,
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
			seederdata.SeedData.Space.Slug,
		),
		managementGoogleAuth: applicationmanagementgoogleauth.NewUsecaseImpl(
			l.infra.user,
			l.infra.space,
			l.infra.spaceMember,
			l.infra.googleAuth,
			l.infra.permissionGuard,
			l.infra.logger,
			seederdata.SeedData.Space.Slug,
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
			seederdata.SeedData.Space.Slug,
		),
		managementSpaceMember: applicationmanagementspacemember.NewUsecaseImpl(
			l.infra.user,
			l.infra.space,
			l.infra.spaceMember,
			l.infra.spaceMemberRole,
			l.infra.permissionGuard,
			l.infra.logger,
			l.infra.transactor,
			seederdata.SeedData.Space.Slug,
		),
		managementRolePermission: applicationmanagementrolepermission.NewUsecaseImpl(
			l.infra.user,
			l.infra.space,
			l.infra.spaceMember,
			l.infra.rolePermission,
			l.infra.permissionGuard,
			l.infra.logger,
			l.infra.transactor,
			seederdata.SeedData.Space.Slug,
		),
		managementSpaceMemberRole: applicationmanagementspacememberrole.NewUsecaseImpl(
			l.infra.user,
			l.infra.space,
			l.infra.spaceMember,
			l.infra.spaceMemberRole,
			l.infra.permissionGuard,
			l.infra.logger,
			l.infra.transactor,
			seederdata.SeedData.Space.Slug,
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
		miscSigningKey: applicationmiscsigningkey.NewUsecaseImpl(
			l.infra.signingKey,
			l.infra.secret,
			l.infra.transactor,
			l.infra.logger,
			applicationmiscsigningkey.Config{
				RotationInterval:  config.Env.SigningKeyRotationInterval,
				RetentionInterval: config.Env.SigningKeyRetentionInterval,
			},
		),
		miscJwks: applicationmiscjwks.NewUsecaseImpl(l.infra.signingKey),
	}
}
