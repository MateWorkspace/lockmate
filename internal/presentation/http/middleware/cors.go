package presentationhttpmiddleware

import (
	"net/http"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS(origins domainusecasesmisc.AllowedOrigin) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Add("Vary", "Origin")
		values := c.Request.Header.Values("Origin")
		if len(values) == 0 {
			c.Next()
			return
		}
		if len(values) != 1 || values[0] == "" {
			presentationhttputils.Error(c, domainmodelserror.NewError("origin is not allowed", domainmodelserror.Forbidden, nil))
			return
		}
		origin := values[0]
		allowed, err := origins.IsAllowedOrigin(c.Request.Context(), origin)
		if err != nil {
			presentationhttputils.Error(c, err)
			return
		}
		if !allowed {
			presentationhttputils.Error(c, domainmodelserror.NewError("origin is not allowed", domainmodelserror.Forbidden, nil))
			return
		}
		// Lookup and error handling run outside the library's boolean callback.
		// Rebuild the exact-origin configuration per request; no allowlist cache.
		apply := cors.New(cors.Config{
			AllowOrigins:              []string{origin},
			AllowMethods:              []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete, http.MethodOptions},
			AllowHeaders:              []string{"Accept", "Content-Type", "Authorization", ClientSlugHeader, ClientSecretHeader, ClientExecIdHeader},
			AllowCredentials:          false,
			MaxAge:                    0,
			OptionsResponseStatusCode: http.StatusNoContent,
		})
		if c.Request.Method == http.MethodOptions {
			// The library omits max-age at zero; send zero explicitly because
			// browsers otherwise apply their default preflight cache lifetime.
			c.Header("Access-Control-Max-Age", "0")
		}
		apply(c)
		if c.IsAborted() {
			return
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
