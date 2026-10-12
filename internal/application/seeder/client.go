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

		for _, origin := range value.Origins {
			_, err := u.clientOrigin.ReadByClientIdAndOrigin(ctx, spaceId, clientId, origin)
			if err == nil {
				continue
			}

			if !errors.Is(err, domainmodelserror.ClientOriginNotFound) {
				u.logger.Error(ctx, tag, "failed to read seed client origin", domainmodelslogger.Meta{"err": err})

				return err
			}

			id, err := u.clientOrigin.Create(ctx, spaceId, domaincontractsrepository.CreateClientOriginArgs{
				ClientId: clientId,
				Origin:   origin,
			})
			if err != nil {
				if canRetrySeed(ctx, err) {
					u.logger.Warn(ctx, tag, "failed to create seed client origin", domainmodelslogger.Meta{"err": err})
				} else {
					u.logger.Error(ctx, tag, "failed to create seed client origin", domainmodelslogger.Meta{"err": err})
				}

				return err
			}

			created.origins = append(created.origins, record{id: id, identifier: value.Slug})
		}
	}

	return nil
}
