package infrastructurerepositorygoogleoauthlaunch

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

func NewPostgresImpl(dt pgxdt.Pgxdt, sqr *squirrel.StatementBuilderType) domaincontractsrepository.GoogleOAuthLaunch {
	return &postgresImpl{
		dt:  dt,
		sqr: sqr,
	}
}

func (p *postgresImpl) Create(ctx context.Context, value domainmodelsentity.GoogleOAuthLaunch) error {
	query, args, err := p.queryCreate(value)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build google_oauth_launch creation", err)
	}
	_, err = p.dt.Exec(ctx, query, args...)
	return infrastructurerepositoryshared.MapPgxError("failed to create google_oauth_launch", err)
}

func (p *postgresImpl) Consume(ctx context.Context, ticketHash string) (*domainmodelsentity.GoogleOAuthLaunch, error) {
	query, args, err := p.queryConsume(ticketHash)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build google_oauth_launch consumption", err)
	}

	var value domainmodelsentity.GoogleOAuthLaunch
	if err := p.dt.QueryRow(ctx, query, args...).Scan(
		&value.TicketHash,
		&value.ClientId,
		&value.SpaceId,
		&value.RedirectUri,
		&value.ClientState,
		&value.ExpiresAt,
	); err != nil {
		return nil, infrastructurerepositoryshared.ReadError("google_oauth_launch is invalid or already used", err, domainmodelserror.GoogleOAuthStateInvalid)
	}

	return &value, nil
}

func (p *postgresImpl) PurgeExpired(ctx context.Context, limit int64) (int64, error) {
	if limit <= 0 || limit > 1000 {
		return 0, domainmodelserror.NewError("cleanup limit must be between 1 and 1000", domainmodelserror.BadArgs, nil)
	}

	query, args, err := p.queryPurge(limit)
	if err != nil {
		return 0, infrastructurerepositoryshared.QueryBuildError("failed to build google_oauth_launch cleanup", err)
	}

	result, err := p.dt.Exec(ctx, query, args...)
	if err != nil {
		return 0, infrastructurerepositoryshared.MapPgxError("failed to purge google_oauth_launch", err)
	}

	return result.RowsAffected(), nil
}
