package applicationseeder

import (
	"context"
	"errors"
	"strings"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
)

func (u *usecase) seedClients(ctx context.Context, spaceId int64, created *changes) error {
	const tag = tagPath + "/seedClients"

	for _, value := range u.data.Clients {
		client, err := u.client.ReadBySlug(ctx, value.Slug)
		var clientId int64
		if err == nil {
			if client == nil {
				err := domainmodelserror.NewError("seed client read returned no record", domainmodelserror.Failure, nil)
				u.logger.Error(ctx, tag, "seed client read returned no record", domainmodelslogger.Meta{"err": err})

				return err
			}

			if client.SpaceId != spaceId {
				err := domainmodelserror.NewError("seed client belongs to another space", domainmodelserror.BadState, nil)
				u.logger.Error(ctx, tag, "seed client belongs to another space", domainmodelslogger.Meta{"err": err})

				return err
			}

			clientId = client.Id
		} else if errors.Is(err, domainmodelserror.ClientNotFound) {
			// Compare redaction follows the existing client-key format; never retain a
			// generated alternative key or replace a configured secret on subsequent runs.
			redacted := value.SecretKey[:4] + strings.Repeat("*", len(value.SecretKey)-8) + value.SecretKey[len(value.SecretKey)-4:]
			clientId, err = u.client.Create(ctx, spaceId, domaincontractsrepository.CreateClientArgs{
				Slug: value.Slug, Name: value.Name, Description: &value.Description,
				SecretKeyHash: u.keygen.Hash(value.SecretKey), SecretKeyRedacted: redacted,
				IsActive: &value.IsActive, AuthRedirectUri: value.AuthRedirectUri,
			})
			if err != nil {
				if canRetrySeed(ctx, err) {
					u.logger.Warn(ctx, tag, "failed to create seed client", domainmodelslogger.Meta{"err": err})
				} else {
					u.logger.Error(ctx, tag, "failed to create seed client", domainmodelslogger.Meta{"err": err})
				}

				return err
			}

			created.clients = append(created.clients, record{id: clientId, identifier: value.Slug})
		} else {
			u.logger.Error(ctx, tag, "failed to read seed client", domainmodelslogger.Meta{"err": err})

			return err
		}

		for _, cidr := range value.IpAllowlist {
			_, err := u.clientIpAllowlist.ReadByClientIdAndCidr(ctx, spaceId, clientId, cidr)
			if err == nil {
				continue
			}

			if !errors.Is(err, domainmodelserror.ClientIpAllowlistNotFound) {
				u.logger.Error(ctx, tag, "failed to read seed client IP allowlist", domainmodelslogger.Meta{"err": err})

				return err
			}

			id, err := u.clientIpAllowlist.Create(ctx, spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{
				ClientId: clientId,
				Cidr:     cidr,
			})
			if err != nil {
				if canRetrySeed(ctx, err) {
					u.logger.Warn(ctx, tag, "failed to create seed client IP allowlist", domainmodelslogger.Meta{"err": err})
				} else {
					u.logger.Error(ctx, tag, "failed to create seed client IP allowlist", domainmodelslogger.Meta{"err": err})
				}

				return err
			}

			created.ipAllowlist = append(created.ipAllowlist, record{id: id, identifier: value.Slug})
		}
	}

	return nil
}
