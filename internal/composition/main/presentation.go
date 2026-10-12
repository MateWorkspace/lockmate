package compositionmain

import (
	"context"

	"github.com/MateWorkspace/lockmate/internal/config"
	presentationcronjob "github.com/MateWorkspace/lockmate/internal/presentation/cron/job"
	presentationcronrunner "github.com/MateWorkspace/lockmate/internal/presentation/cron/runner"
	presentationgrpcinterceptor "github.com/MateWorkspace/lockmate/internal/presentation/grpc/interceptor"
	presentationgrpcregister "github.com/MateWorkspace/lockmate/internal/presentation/grpc/register"
	presentationhttphandlerauth "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/auth"
	presentationhttphandlermanagement "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/management"
	presentationhttphandlerprofile "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/profile"
	presentationhttphandlerversion "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/version"
	presentationhttproute "github.com/MateWorkspace/lockmate/internal/presentation/http/route"
	"google.golang.org/grpc"
)

// presentation owns the HTTP routes and the cron scheduler lifecycle.
type presentation struct {
	cronStop func()
}

func (p *presentation) stop() {
	if p.cronStop != nil {
		p.cronStop()
	}
}

func (l *launcher) newPresentation(ctx context.Context) error {
	auth := presentationhttphandlerauth.NewHandler(l.app.authSession)
	jwks := presentationhttphandlerauth.NewJwksHandler(l.app.miscJwks)
	profile := presentationhttphandlerprofile.NewHandler(l.app.profileAccount, l.app.profileSecurity)
	management := presentationhttphandlermanagement.NewHandler(
		l.app.managementSpace, l.app.managementPermission, l.app.managementRole,
		l.app.managementUser, l.app.managementClassicAuth, l.app.managementGoogleAuth,
		l.app.managementClient, l.app.managementSpaceMember, l.app.managementRolePermission, l.app.managementSpaceMemberRole,
	)
	version := presentationhttphandlerversion.NewHandler(config.App)

	presentationhttproute.Route(l.drv.gin, presentationhttproute.Args{
		Flow: presentationhttphandlerauth.NewFlowHandler(l.app.authGoogle, l.app.authHandoff),
		Auth: auth, Jwks: jwks, Profile: profile, Management: management, Version: version,
		Client: l.app.authClient, Session: l.app.authSession,
		Tracer: l.app.miscTracer, IpAllowlist: l.app.miscIpAllowlist,
		TrustedProxies: config.Env.TrustedProxies, ClientIpHeader: config.Env.ClientIpHeader,
	})

	cronStop, err := presentationcronrunner.Run(ctx, presentationcronrunner.Args{
		Timeout: config.Env.PostgresConnectTimeout,
		Jobs: []presentationcronrunner.Schedule{
			{Name: "signing_key", Interval: config.Env.SigningKeyReconcileInterval, Run: presentationcronjob.SigningKey(l.app.miscSigningKey)},
			{Name: "auth_flow", Interval: config.Env.AuthFlowCleanupInterval, Optional: true, Run: presentationcronjob.AuthFlow(l.app.miscAuthFlow)},
		},
	})
	if err != nil {
		return err
	}

	l.newGrpcServer()

	l.pres = &presentation{cronStop: cronStop}

	return nil
}

func (l *launcher) newGrpcServer() {
	server := grpc.NewServer(presentationgrpcinterceptor.Unary(presentationgrpcinterceptor.Args{
		Client:         l.app.authClient,
		Session:        l.app.authSession,
		Tracer:         l.app.miscTracer,
		IpAllowlist:    l.app.miscIpAllowlist,
		TrustedProxies: config.Env.TrustedProxies,
		ClientIpHeader: config.Env.ClientIpHeader,
	}))

	presentationgrpcregister.RegisterAuth(server, presentationgrpcregister.AuthDependencies{
		Session: l.app.authSession,
		Handoff: l.app.authHandoff,
		Google:  l.app.authGoogle,
	})
	presentationgrpcregister.RegisterProfile(server, presentationgrpcregister.ProfileDependencies{
		Account:  l.app.profileAccount,
		Security: l.app.profileSecurity,
	})
	presentationgrpcregister.RegisterManagement(server, presentationgrpcregister.ManagementDependencies{
		Space:           l.app.managementSpace,
		Permission:      l.app.managementPermission,
		Role:            l.app.managementRole,
		User:            l.app.managementUser,
		ClassicAuth:     l.app.managementClassicAuth,
		GoogleAuth:      l.app.managementGoogleAuth,
		Client:          l.app.managementClient,
		SpaceMember:     l.app.managementSpaceMember,
		RolePermission:  l.app.managementRolePermission,
		SpaceMemberRole: l.app.managementSpaceMemberRole,
	})
	presentationgrpcregister.RegisterMisc(server, presentationgrpcregister.MiscDependencies{
		Jwks:    l.app.miscJwks,
		Version: config.App,
	})

	l.drv.grpcServer = server
}
