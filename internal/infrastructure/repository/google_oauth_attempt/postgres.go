package infrastructurerepositorygoogleoauthattempt

import (
	"context"

	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructurerepositoryshared "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/shared"
	"github.com/MateWorkspace/mate-golang/pgxdt"
)

type postgresImpl struct {
	dt  pgxdt.Pgxdt
	sqr *squirrel.StatementBuilderType
}

func NewPostgresImpl(dt pgxdt.Pgxdt, sqr *squirrel.StatementBuilderType) domaincontractsrepository.GoogleOAuthAttempt {
	return &postgresImpl{
		dt:  dt,
		sqr: sqr,
	}
}

func (p *postgresImpl) Create(ctx context.Context, value domainmodelsentity.GoogleOAuthAttempt) error {
	query, args, err := p.queryCreate(value)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build google_oauth_attempt creation", err)
	}
	_, err = p.dt.Exec(ctx, query, args...)
	return infrastructurerepositoryshared.MapPgxError("failed to create google_oauth_attempt", err)
}

func (p *postgresImpl) Consume(ctx context.Context, stateHash string, bindingHash string) (*domainmodelsentity.GoogleOAuthAttempt, error) {
	query, args, err := p.queryConsume(stateHash, bindingHash)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build google_oauth_attempt consumption", err)
	}

	var value domainmodelsentity.GoogleOAuthAttempt
	if err := p.dt.QueryRow(ctx, query, args...).Scan(
		&value.StateHash,
		&value.BindingHash,
		&value.NonceHash,
		&value.Verifier,
		&value.ClientId,
		&value.SpaceId,
		&value.RedirectUri,
		&value.ClientState,
		&value.ExpiresAt,
	); err != nil {
		return nil, infrastructurerepositoryshared.ReadError("google_oauth_attempt is invalid or already used", err, domainmodelserror.GoogleOAuthStateInvalid)
	}

	return &value, nil
}

func (p *postgresImpl) PurgeExpired(ctx context.Context, limit int64) (int64, error) {
	if limit <= 0 || limit > 1000 {
		return 0, domainmodelserror.NewError("cleanup limit must be between 1 and 1000", domainmodelserror.BadArgs, nil)
	}

	query, args, err := p.queryPurge(limit)
	if err != nil {
		return 0, infrastructurerepositoryshared.QueryBuildError("failed to build google_oauth_attempt cleanup", err)
	}

	result, err := p.dt.Exec(ctx, query, args...)
	if err != nil {
		return 0, infrastructurerepositoryshared.MapPgxError("failed to purge google_oauth_attempt", err)
	}

	return result.RowsAffected(), nil
}
