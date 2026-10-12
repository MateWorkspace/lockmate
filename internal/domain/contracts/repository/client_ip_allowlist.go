package domaincontractsrepository

import (
	"context"
	"net/netip"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

// ClientIpAllowlist persists canonical CIDR entries validated by
// Validator.ClientIpAllowlist. Space ownership is derived through the client.
// Ordinary reads include inactive records but exclude deleted entries and
// parents; allowlist checks require the ownership chain to be active and
// nondeleted.
type ClientIpAllowlist interface {
	Create(ctx context.Context, spaceId int64, args CreateClientIpAllowlistArgs) (id int64, err error)
	ReadById(ctx context.Context, spaceId, id int64) (*domainmodelsentity.ClientIpAllowlist, error)
	ReadByClientIdAndCidr(ctx context.Context, spaceId, clientId int64, cidr string) (*domainmodelsentity.ClientIpAllowlist, error)
	ReadByClientId(ctx context.Context, spaceId, clientId int64) ([]domainmodelsentity.ClientIpAllowlist, error)
	ReadByFilter(ctx context.Context, spaceId int64, filter ClientIpAllowlistFilter) ([]domainmodelsentity.ClientIpAllowlist, int64, error)
	UpdateById(ctx context.Context, spaceId, id int64, args UpdateClientIpAllowlistArgs) error
	DeleteById(ctx context.Context, spaceId, id int64, by *int64) error

	// IsAllowedForClient reports whether ip matches an active, nondeleted entry
	// owned by the live client and space. Missing or inactive matches return
	// false without a not-found error.
	IsAllowedForClient(ctx context.Context, spaceId, clientId int64, ip netip.Addr) (bool, error)
}

type CreateClientIpAllowlistArgs struct {
	ClientId int64
	Cidr     string
	IsActive *bool
	By       *int64
}

type UpdateClientIpAllowlistArgs struct {
	Cidr     *string
	IsActive *bool
	By       *int64
}

type ClientIpAllowlistFilter struct {
	Page     int64
	Limit    int64
	ClientId *int64
	Search   *string
	IsActive *bool
}
