package presentationshared

import (
	"net"
	"net/netip"
	"strings"
)

// Resolve returns the caller IP. Forwarded header values are honored only when
// the direct peer is a configured trusted proxy; otherwise the socket peer is
// used. The configured header is typically CF-Connecting-IP behind
// cloudflared/Cloudflare, or X-Forwarded-For behind nginx/Caddy.
func Resolve(peer netip.Addr, headerValues []string, trusted []netip.Prefix) netip.Addr {
	if !peer.IsValid() {
		return netip.Addr{}
	}

	if !InAny(peer, trusted) {
		return peer
	}

	for _, value := range headerValues {
		parts := strings.Split(value, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			address, err := netip.ParseAddr(strings.TrimSpace(parts[i]))
			if err != nil {
				continue
			}

			address = address.Unmap()
			if !InAny(address, trusted) {
				return address
			}
		}
	}

	return peer
}

// FromRemoteAddr parses a host:port or bare address like http.Request.RemoteAddr.
func FromRemoteAddr(remoteAddr string) netip.Addr {
	if addressPort, err := netip.ParseAddrPort(remoteAddr); err == nil {
		return addressPort.Addr().Unmap()
	}

	if address, err := netip.ParseAddr(remoteAddr); err == nil {
		return address.Unmap()
	}

	return netip.Addr{}
}

// FromAddr parses a net.Addr such as a gRPC peer address.
func FromAddr(addr net.Addr) netip.Addr {
	if addr == nil {
		return netip.Addr{}
	}

	if tcp, ok := addr.(*net.TCPAddr); ok {
		if address, ok := netip.AddrFromSlice(tcp.IP); ok {
			return address.Unmap()
		}
	}

	switch value := addr.(type) {
	case *net.IPAddr:
		if address, ok := netip.AddrFromSlice(value.IP); ok {
			return address.Unmap()
		}
		return netip.Addr{}
	default:
		return FromRemoteAddr(addr.String())
	}
}

// InAny reports whether the address falls within any prefix.
func InAny(address netip.Addr, prefixes []netip.Prefix) bool {
	for _, prefix := range prefixes {
		if prefix.Contains(address) {
			return true
		}
	}

	return false
}
