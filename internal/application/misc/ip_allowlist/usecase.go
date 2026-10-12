package applicationmiscipallowlist

import (
	"context"
	"net/netip"

	applicationshared "github.com/MateWorkspace/lockmate/internal/application/shared"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
)

const tagPath = "misc/ip_allowlist"

type usecase struct {
	clientIpAllowlist domaincontractsrepository.ClientIpAllowlist
	logger            domaincontractsutility.Logger
}

func NewUsecaseImpl(
	clientIpAllowlist domaincontractsrepository.ClientIpAllowlist,
	logger domaincontractsutility.Logger,
) domainusecasesmisc.IpAllowlist {
	return &usecase{clientIpAllowlist: clientIpAllowlist, logger: logger}
}

func (u *usecase) IsAllowedIpForClient(ctx context.Context, spaceId, clientId int64, ip string) (bool, error) {
	const tag = tagPath + "/IsAllowedIpForClient"

	if err := applicationshared.ValidateIds(spaceId, clientId); err != nil {
		u.logger.Error(ctx, tag, "invalid identifiers for client IP allowlist check", domainmodelslogger.Meta{
			"space_id":  spaceId,
			"client_id": clientId,
			"err":       err,
		})

		return false, err
	}

	address, err := netip.ParseAddr(ip)
	if err != nil {
		// An unparseable caller address is an ordinary nonmatch.
		return false, nil
	}

	address = address.Unmap()
	if address.IsLoopback() {
		return true, nil
	}

	allowed, err := u.clientIpAllowlist.IsAllowedForClient(ctx, spaceId, clientId, address)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to check client IP allowlist", domainmodelslogger.Meta{
			"space_id":  spaceId,
			"client_id": clientId,
			"err":       err,
		})

		return false, err
	}

	return allowed, nil
}
