package presentationhttpmiddleware

import (
	"strings"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-gonic/gin"
)

const (
	ClientSlugHeader   = "X-Client-Slug"
	ClientSecretHeader = "X-Client-Secret"
	ClientExecIdHeader = "X-Client-Exec-Id"
)

func Client(auth domainusecasesauth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, header := range []string{ClientSlugHeader, ClientSecretHeader, ClientExecIdHeader} {
			if len(c.Request.Header.Values(header)) > 1 {
				presentationhttputils.Error(c, domainmodelserror.NewError(header+" must occur once", domainmodelserror.BadArgs, nil))
				return
			}
		}
		slug, secret := c.GetHeader(ClientSlugHeader), c.GetHeader(ClientSecretHeader)
		if strings.TrimSpace(slug) == "" || strings.TrimSpace(secret) == "" {
			presentationhttputils.Error(c, domainmodelserror.NewError("client credentials are required", domainmodelserror.Unauthorized, nil))
			return
		}
		if _, err := presentationhttputils.RequiredUUID(c.GetHeader(ClientExecIdHeader), ClientExecIdHeader); err != nil {
			presentationhttputils.Error(c, err)
			return
		}
		client, err := auth.Authenticate(c.Request.Context(), domainusecasesauth.AuthenticateClientRequest{Slug: slug, Secret: secret})
		if err != nil {
			presentationhttputils.Error(c, err)
			return
		}
		if client == nil || client.Id <= 0 || client.SpaceId <= 0 {
			presentationhttputils.Error(c, presentationhttputils.MissingResponse("client identity"))
			return
		}
		ctx := presentationhttputils.InjectClientIdentity(c.Request.Context(), client)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
