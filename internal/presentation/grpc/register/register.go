// Package presentationgrpcregister wires gRPC service implementations onto a
// registrar with nil guards.
package presentationgrpcregister

import (
	lockmatev1 "github.com/MateWorkspace/lockmate-proto/gen/go/lockmate/v1"
	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
	domainusecasesprofile "github.com/MateWorkspace/lockmate/internal/domain/usecases/profile"
	presentationgrpcserverauthgoogle "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/auth/google"
	presentationgrpcserverauthhandoff "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/auth/handoff"
	presentationgrpcserverauthsession "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/auth/session"
	presentationgrpcservermanagementclassicauth "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/management/classic_auth"
	presentationgrpcservermanagementclient "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/management/client"
	presentationgrpcservermanagementgoogleauth "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/management/google_auth"
	presentationgrpcservermanagementpermission "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/management/permission"
	presentationgrpcservermanagementrole "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/management/role"
	presentationgrpcservermanagementrolepermission "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/management/role_permission"
	presentationgrpcservermanagementspace "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/management/space"
	presentationgrpcservermanagementspacemember "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/management/space_member"
	presentationgrpcservermanagementspacememberrole "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/management/space_member_role"
	presentationgrpcservermanagementuser "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/management/user"
	presentationgrpcservermiscjwks "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/misc/jwks"
	presentationgrpcservermiscversion "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/misc/version"
	presentationgrpcserverprofileaccount "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/profile/account"
	presentationgrpcserverprofilesecurity "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/profile/security"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type AuthDependencies struct {
	Session domainusecasesauth.Session
	Handoff domainusecasesauth.Handoff
	Google  domainusecasesauth.Google
}

func RegisterAuth(registrar grpc.ServiceRegistrar, dependencies AuthDependencies) {
	if dependencies.Session != nil {
		lockmatev1.RegisterSessionServiceServer(registrar, presentationgrpcserverauthsession.NewServer(dependencies.Session))
	}

	if dependencies.Handoff != nil {
		lockmatev1.RegisterHandoffServiceServer(registrar, presentationgrpcserverauthhandoff.NewServer(dependencies.Handoff))
	}

	if dependencies.Google != nil {
		lockmatev1.RegisterGoogleServiceServer(registrar, presentationgrpcserverauthgoogle.NewServer(dependencies.Google))
	}
}

type ProfileDependencies struct {
	Account  domainusecasesprofile.Account
	Security domainusecasesprofile.Security
}

func RegisterProfile(registrar grpc.ServiceRegistrar, dependencies ProfileDependencies) {
	if dependencies.Account != nil {
		lockmatev1.RegisterAccountServiceServer(registrar, presentationgrpcserverprofileaccount.NewServer(dependencies.Account))
	}

	if dependencies.Security != nil {
		lockmatev1.RegisterSecurityServiceServer(registrar, presentationgrpcserverprofilesecurity.NewServer(dependencies.Security))
	}
}

type ManagementDependencies struct {
	Space           domainusecasesmanagement.Space
	Permission      domainusecasesmanagement.Permission
	Role            domainusecasesmanagement.Role
	User            domainusecasesmanagement.User
	ClassicAuth     domainusecasesmanagement.ClassicAuth
	GoogleAuth      domainusecasesmanagement.GoogleAuth
	Client          domainusecasesmanagement.Client
	SpaceMember     domainusecasesmanagement.SpaceMember
	RolePermission  domainusecasesmanagement.RolePermission
	SpaceMemberRole domainusecasesmanagement.SpaceMemberRole
}

func RegisterManagement(registrar grpc.ServiceRegistrar, dependencies ManagementDependencies) {
	if dependencies.Space != nil {
		lockmatev1.RegisterSpaceServiceServer(registrar, presentationgrpcservermanagementspace.NewServer(dependencies.Space))
	}

	if dependencies.Permission != nil {
		lockmatev1.RegisterPermissionServiceServer(registrar, presentationgrpcservermanagementpermission.NewServer(dependencies.Permission))
	}

	if dependencies.Role != nil {
		lockmatev1.RegisterRoleServiceServer(registrar, presentationgrpcservermanagementrole.NewServer(dependencies.Role))
	}

	if dependencies.User != nil {
		lockmatev1.RegisterUserServiceServer(registrar, presentationgrpcservermanagementuser.NewServer(dependencies.User))
	}

	if dependencies.ClassicAuth != nil {
		lockmatev1.RegisterClassicAuthServiceServer(registrar, presentationgrpcservermanagementclassicauth.NewServer(dependencies.ClassicAuth))
	}

	if dependencies.GoogleAuth != nil {
		lockmatev1.RegisterGoogleAuthServiceServer(registrar, presentationgrpcservermanagementgoogleauth.NewServer(dependencies.GoogleAuth))
	}

	if dependencies.Client != nil {
		lockmatev1.RegisterClientServiceServer(registrar, presentationgrpcservermanagementclient.NewServer(dependencies.Client))
	}

	if dependencies.SpaceMember != nil {
		lockmatev1.RegisterSpaceMemberServiceServer(registrar, presentationgrpcservermanagementspacemember.NewServer(dependencies.SpaceMember))
	}

	if dependencies.RolePermission != nil {
		lockmatev1.RegisterRolePermissionServiceServer(registrar, presentationgrpcservermanagementrolepermission.NewServer(dependencies.RolePermission))
	}

	if dependencies.SpaceMemberRole != nil {
		lockmatev1.RegisterSpaceMemberRoleServiceServer(registrar, presentationgrpcservermanagementspacememberrole.NewServer(dependencies.SpaceMemberRole))
	}
}

type MiscDependencies struct {
	Jwks    domainusecasesmisc.Jwks
	Version domainmodelsapp.Info
}

// RegisterMisc registers the JWKS/version services and the unauthenticated
// standard health service.
func RegisterMisc(registrar grpc.ServiceRegistrar, dependencies MiscDependencies) {
	if dependencies.Jwks != nil {
		lockmatev1.RegisterJwksServiceServer(registrar, presentationgrpcservermiscjwks.NewServer(dependencies.Jwks))
	}

	lockmatev1.RegisterVersionServiceServer(registrar, presentationgrpcservermiscversion.NewServer(dependencies.Version))

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(registrar, healthServer)
}
