package presentationhttpmiddleware

import (
	"strings"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-gonic/gin"
)

func Auth(session domainusecasesauth.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		values := c.Request.Header.Values("Authorization")
		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(values) != 1 || len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			presentationhttputils.Error(c, domainmodelserror.NewError("bearer authorization is required", domainmodelserror.Unauthorized, nil))
			return
		}
		client := presentationhttputils.ClientIdentity(c.Request.Context())
		if client == nil {
			presentationhttputils.Error(c, domainmodelserror.NewError("client authentication is required", domainmodelserror.Unauthorized, nil))
			return
		}
		claims, err := session.AuthenticateAccessToken(c.Request.Context(), client, parts[1])
		if err != nil {
			presentationhttputils.Error(c, err)
			return
		}
		if claims == nil || claims.UserId <= 0 || claims.SpaceId <= 0 || claims.SpaceMemberId <= 0 || claims.SpaceId != client.SpaceId {
			presentationhttputils.Error(c, presentationhttputils.MissingResponse("access claims"))
			return
		}
		c.Request = c.Request.WithContext(presentationhttputils.InjectAccessClaims(c.Request.Context(), claims))
		c.Next()
	}
}
