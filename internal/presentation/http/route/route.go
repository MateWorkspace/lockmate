package presentationhttproute

import (
	"net/http"

	_ "github.com/MateWorkspace/lockmate/docs/swagger"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
	presentationhttphandlerauth "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/auth"
	presentationhttpmiddleware "github.com/MateWorkspace/lockmate/internal/presentation/http/middleware"
	presentationhttpresponse "github.com/MateWorkspace/lockmate/internal/presentation/http/response"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Args struct {
	Flow          AuthFlowHandler
	Auth          AuthHandler
	Profile       ProfileHandler
	Management    ManagementHandler
	Version       VersionHandler
	Client        domainusecasesauth.Client
	Session       domainusecasesauth.Session
	Tracer        domainusecasesmisc.Tracer
	AllowedOrigin domainusecasesmisc.AllowedOrigin
}

func Route(e *gin.Engine, args Args) {
	e.HandleMethodNotAllowed = true
	e.Use(presentationhttpmiddleware.Recovery())
	e.Use(presentationhttpmiddleware.Trace(args.Tracer))
	e.Use(presentationhttpmiddleware.CORS(args.AllowedOrigin))

	api := e.Group("/api")
	api.GET("/docs", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/api/docs/index.html") })
	api.GET("/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, ginSwagger.PersistAuthorization(false)))
	api.GET("/version", args.Version.VersionGet)

	publicV1 := api.Group("/v1")
	publicV1.Use(presentationhttpmiddleware.Client(args.Client, args.AllowedOrigin))
	routeAuth(publicV1, args.Auth)
	if args.Flow == nil {
		args.Flow = presentationhttphandlerauth.NewFlowHandler(nil, nil)
	}
	routeAuthFlow(publicV1, api.Group("/v1"), args.Flow)

	v1 := api.Group("/v1")
	v1.Use(presentationhttpmiddleware.Client(args.Client, args.AllowedOrigin))
	v1.Use(presentationhttpmiddleware.Auth(args.Session))
	routeProfile(v1, args.Profile)
	routeManagement(v1, args.Management)

	e.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, presentationhttpresponse.ErrorResponse{Error: "Not Found", Message: "The requested route does not exist."})
	})
	e.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, presentationhttpresponse.ErrorResponse{Error: "Method Not Allowed", Message: "The HTTP method is not supported for this route."})
	})
}
