package compositionmain

import (
	"github.com/MateWorkspace/lockmate/internal/config"
	presentationhttphandlerauth "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/auth"
	presentationhttphandlermanagement "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/management"
	presentationhttphandlerprofile "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/profile"
	presentationhttphandlerversion "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/version"
	presentationhttproute "github.com/MateWorkspace/lockmate/internal/presentation/http/route"
)

type presentation struct{}

func (l *launcher) newPresentation() {
	auth := presentationhttphandlerauth.NewHandler(l.app.authSession)
	profile := presentationhttphandlerprofile.NewHandler(l.app.profileAccount, l.app.profileSecurity)
	management := presentationhttphandlermanagement.NewHandler(
		l.app.managementSpace, l.app.managementPermission, l.app.managementRole,
		l.app.managementUser, l.app.managementClassicAuth, l.app.managementGoogleAuth,
		l.app.managementClient, l.app.managementSpaceMember, l.app.managementRolePermission, l.app.managementSpaceMemberRole,
	)
	version := presentationhttphandlerversion.NewHandler(config.App)

	presentationhttproute.Route(l.drv.gin, presentationhttproute.Args{
		Flow: presentationhttphandlerauth.NewFlowHandler(l.app.authGoogle, l.app.authHandoff),
		Auth: auth, Profile: profile, Management: management, Version: version,
		Client: l.app.authClient, Session: l.app.authSession,
		Tracer: l.app.miscTracer, AllowedOrigin: l.app.miscAllowedOrigin,
	})
	l.pres = &presentation{}
}
