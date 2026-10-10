package applicationauthclient

import (
	"context"
	"errors"

	applicationshared "github.com/MateWorkspace/lockmate/internal/application/shared"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
)

const tagPath = "auth/client"

type usecase struct {
	client domaincontractsrepository.Client
	space  domaincontractsrepository.Space
	keygen domaincontractsutility.Keygen
	logger domaincontractsutility.Logger
}

func NewUsecaseImpl(
	client domaincontractsrepository.Client,
	space domaincontractsrepository.Space,
	keygen domaincontractsutility.Keygen,
	logger domaincontractsutility.Logger,
) domainusecasesauth.Client {
	return &usecase{client: client, space: space, keygen: keygen, logger: logger}
}

func (u *usecase) Authenticate(ctx context.Context, request domainusecasesauth.AuthenticateClientRequest) (result *domainmodelsentity.ClientIdentity, err error) {
	const tag = tagPath + "/Authenticate"

	if request.Slug == "" || request.Secret == "" {
		failure := domainmodelserror.NewError("client credentials are required", domainmodelserror.BadArgs, nil)
		u.logger.Error(ctx, tag, "client credentials are required", domainmodelslogger.Meta{
			"err": failure,
		})

		return nil, failure
	}

	client, err := u.client.ReadBySlug(ctx, request.Slug)
	if err != nil {
		failure := applicationshared.IdentityError(err)
		u.logger.Error(ctx, tag, "failed to read client by slug", domainmodelslogger.Meta{
			"err": failure,
		})

		return nil, failure
	}

	if client == nil || client.Slug != request.Slug || client.Id <= 0 || client.SpaceId <= 0 || !client.IsActive || client.AuditDelete.At != nil {
		failure := applicationshared.Unauthorized(nil)
		u.logger.Error(ctx, tag, "invalid client identity", domainmodelslogger.Meta{
			"err": failure,
		})

		return nil, failure
	}

	if err := u.keygen.Compare(client.SecretKeyHash, request.Secret); err != nil {
		if errors.Is(err, domainmodelserror.KeyInvalid) || errors.Is(err, domainmodelserror.KeyMismatch) {
			failure := applicationshared.Unauthorized(err)
			u.logger.Error(ctx, tag, "failed to compare client secret", domainmodelslogger.Meta{
				"err": failure,
			})

			return nil, failure
		}

		u.logger.Error(ctx, tag, "failed to compare client secret", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	result = &domainmodelsentity.ClientIdentity{Id: client.Id, SpaceId: client.SpaceId, Slug: client.Slug}
	if _, err := applicationshared.ReadClient(ctx, u.client, u.space, result); err != nil {
		u.logger.Error(ctx, tag, "failed to validate client identity", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	return result, nil
}
