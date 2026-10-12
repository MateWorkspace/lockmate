package presentationhttpmiddleware

import (
	"net/http"
	"net/netip"
	"strings"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
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

		address := clientIP(c.Request.RemoteAddr, c.Request.Header, trusted, header)
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

// clientIP resolves the caller IP. Forwarded headers are honored only when the
// direct peer is a configured trusted proxy; otherwise the socket peer is used.
// The configured header is typically CF-Connecting-IP behind cloudflared/Cloudflare,
// or X-Forwarded-For behind nginx/Caddy.
func clientIP(remoteAddr string, headers http.Header, trusted []netip.Prefix, header string) netip.Addr {
	peer := peerAddr(remoteAddr)
	if !peer.IsValid() {
		return netip.Addr{}
	}

	if !inAnyPrefix(peer, trusted) {
		return peer
	}

	for _, value := range headers.Values(header) {
		parts := strings.Split(value, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			address, err := netip.ParseAddr(strings.TrimSpace(parts[i]))
			if err != nil {
				continue
			}

			address = address.Unmap()
			if !inAnyPrefix(address, trusted) {
				return address
			}
		}
	}

	return peer
}

func peerAddr(remoteAddr string) netip.Addr {
	if addressPort, err := netip.ParseAddrPort(remoteAddr); err == nil {
		return addressPort.Addr().Unmap()
	}

	if address, err := netip.ParseAddr(remoteAddr); err == nil {
		return address.Unmap()
	}

	return netip.Addr{}
}

func inAnyPrefix(address netip.Addr, prefixes []netip.Prefix) bool {
	for _, prefix := range prefixes {
		if prefix.Contains(address) {
			return true
		}
	}

	return false
}
