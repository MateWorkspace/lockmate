package applicationauthhandoff

import (
	"context"
	"time"

	applicationshared "github.com/MateWorkspace/lockmate/internal/application/shared"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
)

const tagPath = "auth/handoff"

type usecase struct {
	handoffs   domaincontractsrepository.AuthHandoff
	client     domaincontractsrepository.Client
	space      domaincontractsrepository.Space
	grants     domaincontractsrepository.SpaceMemberRole
	flow       domaincontractsutility.AuthFlow
	token      domaincontractsutility.Token
	transactor domaincontractsutility.Transactor
	logger     domaincontractsutility.Logger
	identity   *applicationshared.IdentityReader
}

func NewUsecaseImpl(
	handoffs domaincontractsrepository.AuthHandoff,
	client domaincontractsrepository.Client,
	space domaincontractsrepository.Space,
	user domaincontractsrepository.User,
	members domaincontractsrepository.SpaceMember,
	grants domaincontractsrepository.SpaceMemberRole,
	flow domaincontractsutility.AuthFlow,
	token domaincontractsutility.Token,
	transactor domaincontractsutility.Transactor,
	logger domaincontractsutility.Logger,
) domainusecasesauth.Handoff {
	return &usecase{
		handoffs:   handoffs,
		client:     client,
		space:      space,
		grants:     grants,
		flow:       flow,
		token:      token,
		transactor: transactor,
		logger:     logger,
		identity:   applicationshared.NewIdentityReader(user, space, members, ""),
	}
}

func (u *usecase) Exchange(ctx context.Context, client *domainmodelsentity.ClientIdentity, request domainusecasesauth.ExchangeAuthHandoffRequest) (*domainusecasesauth.SessionResult, error) {
	const tag = tagPath + "/Exchange"

	if len(request.Code) != 43 {
		err := domainmodelserror.NewError("invalid authentication handoff", domainmodelserror.AuthHandoffInvalid, nil)
		u.logger.Error(ctx, tag, "invalid authentication handoff code", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	var result *domainusecasesauth.SessionResult
	err := u.transactor.WithTx(ctx, func(txCtx context.Context) error {
		live, err := applicationshared.ReadClient(txCtx, u.client, u.space, client)
		if err != nil {
			u.logger.Error(txCtx, tag, "failed to validate handoff exchange client", domainmodelslogger.Meta{"err": err})

			return err
		}

		handoff, err := u.handoffs.Consume(txCtx, u.flow.Hash(request.Code), live.Id)
		if err != nil {
			u.logger.Error(txCtx, tag, "failed to consume authentication handoff", domainmodelslogger.Meta{"err": err})

			return err
		}

		if handoff == nil {
			err := domainmodelserror.NewError("authentication handoff record is missing", domainmodelserror.Failure, nil)
			u.logger.Error(txCtx, tag, "authentication handoff record is missing", domainmodelslogger.Meta{"err": err})

			return err
		}

		if !time.Now().Before(handoff.ExpiresAt) {
			err := domainmodelserror.NewError("authentication handoff has expired", domainmodelserror.AuthHandoffExpired, nil)
			u.logger.Error(txCtx, tag, "authentication handoff has expired", domainmodelslogger.Meta{"err": err})

			return err
		}

		if handoff.SpaceId != live.SpaceId || live.AuthRedirectUri == nil || *live.AuthRedirectUri != handoff.RedirectUri {
			err := domainmodelserror.NewError("authentication redirect or scope changed", domainmodelserror.BadState, nil)
			u.logger.Error(txCtx, tag, "authentication handoff no longer matches client configuration", domainmodelslogger.Meta{"err": err})

			return err
		}

		result, err = applicationshared.ResolveSession(txCtx, u.identity, u.grants, handoff.UserId, handoff.SpaceId, handoff.SpaceMemberId)
		if err != nil {
			u.logger.Error(txCtx, tag, "failed to resolve live handoff identity and grants", domainmodelslogger.Meta{"err": err})

			return err
		}

		result.IsNewUser = handoff.IsNewUser
		access, err := u.token.GenerateAccess(applicationshared.SessionClaims(result))
		if err != nil {
			u.logger.Error(txCtx, tag, "failed to generate handoff access token", domainmodelslogger.Meta{"err": err})

			return err
		}

		refresh, err := u.token.GenerateRefresh(&domainmodelsentity.RefreshClaims{
			UserId:        result.User.Id,
			SpaceId:       result.Space.Id,
			SpaceMemberId: result.SpaceMember.Id,
			Name:          result.User.Name,
		})
		if err != nil {
			u.logger.Error(txCtx, tag, "failed to generate handoff refresh token", domainmodelslogger.Meta{"err": err})

			return err
		}

		result.AccessToken, result.RefreshToken = access, refresh
		return nil
	})
	if err != nil {
		u.logger.Error(ctx, tag, "failed to commit authentication handoff exchange", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	return result, nil
}
