package domainusecasesmisc

import "context"

// IpAllowlist checks whether a caller IP is permitted for a client.
// Invalid IPs are ordinary nonmatches; operational failures remain errors.
// An allowance never authenticates or identifies a client.
type IpAllowlist interface {
	IsAllowedIpForClient(ctx context.Context, spaceId, clientId int64, ip string) (bool, error)
}
