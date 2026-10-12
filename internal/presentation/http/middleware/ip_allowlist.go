package presentationhttpmiddleware

import (
	"net/netip"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	presentationshared "github.com/MateWorkspace/lockmate/internal/presentation/shared"
	"github.com/gin-gonic/gin"
)

// IPAllowlist rejects requests whose resolved caller IP is not permitted for the
// authenticated client. It runs after Client so the trusted identity is present.
func IPAllowlist(allowlist domainusecasesmisc.IpAllowlist, trusted []netip.Prefix, header string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client := presentationhttputils.ClientIdentity(c.Request.Context())
		if client == nil {
			presentationhttputils.Error(c, domainmodelserror.NewError("client authentication is required", domainmodelserror.Unauthorized, nil))
			return
		}

		address := presentationshared.Resolve(presentationshared.FromRemoteAddr(c.Request.RemoteAddr), c.Request.Header.Values(header), trusted)
		if !address.IsValid() {
			presentationhttputils.Error(c, domainmodelserror.NewError("caller IP address is unavailable", domainmodelserror.Forbidden, nil))
			return
		}

		allowed, err := allowlist.IsAllowedIpForClient(c.Request.Context(), client.SpaceId, client.Id, address.String())
		if err != nil {
			presentationhttputils.Error(c, err)
			return
		}
		if !allowed {
			presentationhttputils.Error(c, domainmodelserror.NewError("caller IP address is not allowed for this client", domainmodelserror.Forbidden, nil))
			return
		}

		c.Next()
	}
}
