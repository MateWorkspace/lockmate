package compositionmain

import (
	"context"

	"github.com/MateWorkspace/lockmate/internal/config"
	presentationcronjob "github.com/MateWorkspace/lockmate/internal/presentation/cron/job"
	presentationcronrunner "github.com/MateWorkspace/lockmate/internal/presentation/cron/runner"
	presentationhttphandlerauth "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/auth"
	presentationhttphandlermanagement "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/management"
	presentationhttphandlerprofile "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/profile"
	presentationhttphandlerversion "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/version"
	presentationhttproute "github.com/MateWorkspace/lockmate/internal/presentation/http/route"
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
		Tracer: l.app.miscTracer, AllowedOrigin: l.app.miscAllowedOrigin,
	})

	cronStop, err := presentationcronrunner.Run(ctx, presentationcronrunner.Args{
		Timeout: config.Env.PostgresConnectTimeout,
		Jobs: []presentationcronrunner.Schedule{
			{Name: "signing_key", Interval: config.Env.SigningKeyRotationInterval, Run: presentationcronjob.SigningKey(l.app.miscSigningKey)},
			{Name: "auth_flow", Interval: config.Env.AuthFlowCleanupInterval, Optional: true, Run: presentationcronjob.AuthFlow(l.app.miscAuthFlow)},
		},
	})
	if err != nil {
		return err
	}

	l.pres = &presentation{cronStop: cronStop}

	return nil
}
