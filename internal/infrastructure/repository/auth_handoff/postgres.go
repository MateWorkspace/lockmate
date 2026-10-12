package infrastructurerepositoryauthhandoff

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

func NewPostgresImpl(dt pgxdt.Pgxdt, sqr *squirrel.StatementBuilderType) domaincontractsrepository.AuthHandoff {
	return &postgresImpl{
		dt:  dt,
		sqr: sqr,
	}
}

func (p *postgresImpl) Create(ctx context.Context, value domainmodelsentity.AuthHandoff) error {
	query, args, err := p.queryCreate(value)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build auth_handoff creation", err)
	}
	_, err = p.dt.Exec(ctx, query, args...)
	return infrastructurerepositoryshared.MapPgxError("failed to create auth_handoff", err)
}

func (p *postgresImpl) Consume(ctx context.Context, codeHash string, clientId int64) (*domainmodelsentity.AuthHandoff, error) {
	query, args, err := p.queryConsume(codeHash, clientId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build auth_handoff consumption", err)
	}

	var value domainmodelsentity.AuthHandoff
	if err := p.dt.QueryRow(ctx, query, args...).Scan(
		&value.CodeHash,
		&value.ClientId,
		&value.SpaceId,
		&value.UserId,
		&value.SpaceMemberId,
		&value.RedirectUri,
		&value.IsNewUser,
		&value.ExpiresAt,
	); err != nil {
		return nil, infrastructurerepositoryshared.ReadError("auth_handoff is invalid or already used", err, domainmodelserror.AuthHandoffInvalid)
	}

	return &value, nil
}

func (p *postgresImpl) PurgeExpired(ctx context.Context, limit int64) (int64, error) {
	if limit <= 0 || limit > 1000 {
		return 0, domainmodelserror.NewError("cleanup limit must be between 1 and 1000", domainmodelserror.BadArgs, nil)
	}

	query, args, err := p.queryPurge(limit)
	if err != nil {
		return 0, infrastructurerepositoryshared.QueryBuildError("failed to build auth_handoff cleanup", err)
	}

	result, err := p.dt.Exec(ctx, query, args...)
	if err != nil {
		return 0, infrastructurerepositoryshared.MapPgxError("failed to purge auth_handoff", err)
	}

	return result.RowsAffected(), nil
}
