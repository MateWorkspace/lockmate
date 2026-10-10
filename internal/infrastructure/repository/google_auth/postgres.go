package infrastructurerepositorygoogleauth

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
	infrastructurerepositoryshared.BasePostgres
}

func NewPostgresImpl(dt pgxdt.Pgxdt, sqrQuestion, sqrDollar *squirrel.StatementBuilderType) domaincontractsrepository.GoogleAuth {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) mapError(err error) error {
	return infrastructurerepositoryshared.MapPgxError(
		"google auth operation failed",
		err,
		infrastructurerepositoryshared.ConflictMatch{
			Constraint: "uq_google_auths_user_id",
			Type:       domainmodelserror.GoogleAuthUserConflict,
		},
		infrastructurerepositoryshared.ConflictMatch{
			Constraint: "uq_google_auths_subject",
			Type:       domainmodelserror.GoogleAuthSubjectConflict,
		},
	)
}

func (p *postgresImpl) lockParents(ctx context.Context, userId int64) error {
	if err := infrastructurerepositoryshared.LockRow(ctx, p.Dt, "users", userId, nil, domainmodelserror.UserNotFound); err != nil {
		return err
	}

	return nil
}

func (p *postgresImpl) Create(ctx context.Context, args domaincontractsrepository.CreateGoogleAuthArgs) (id int64, err error) {
	if err := infrastructurerepositoryshared.PositiveIds(args.UserId); err != nil {
		return 0, err
	}

	if err := infrastructurerepositoryshared.OptionalIds(args.By); err != nil {
		return 0, err
	}

	if err := infrastructurerepositoryshared.ValidateJSON(args.Preferences); err != nil {
		return 0, err
	}
	err = p.Dt.WithTx(ctx, func(txCtx context.Context) error {
		if err := p.lockParents(txCtx, args.UserId); err != nil {
			return err
		}

		query, values, err := p.queryCreate(args)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build create query", err)
		}

		if err := p.Dt.QueryRow(txCtx, query, values...).Scan(&id); err != nil {
			return p.mapError(err)
		}

		return nil
	})
	if err != nil {
		return 0, p.mapError(err)
	}

	return id, nil
}

func (p *postgresImpl) ReadById(ctx context.Context, id int64) (*domainmodelsentity.GoogleAuth, error) {
	if err := infrastructurerepositoryshared.PositiveIds(id); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadById(id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read query", err)
	}

	item, err := infrastructurerepositoryshared.ScanGoogleAuth(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("google auth not found", err, domainmodelserror.GoogleAuthNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByUserId(ctx context.Context, value int64) (*domainmodelsentity.GoogleAuth, error) {
	if err := infrastructurerepositoryshared.PositiveIds(value); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadByUserId(value)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read query", err)
	}

	item, err := infrastructurerepositoryshared.ScanGoogleAuth(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("google auth not found", err, domainmodelserror.GoogleAuthNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadBySubject(ctx context.Context, value string) (*domainmodelsentity.GoogleAuth, error) {
	query, values, err := p.queryReadBySubject(value)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read query", err)
	}

	item, err := infrastructurerepositoryshared.ScanGoogleAuth(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("google auth not found", err, domainmodelserror.GoogleAuthNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByFilter(ctx context.Context, filter domaincontractsrepository.GoogleAuthFilter) ([]domainmodelsentity.GoogleAuth, int64, error) {
	if err := infrastructurerepositoryshared.OptionalIds(filter.UserId); err != nil {
		return nil, 0, err
	}

	count, countValues, query, values, err := p.queryFilter(filter)
	if err != nil {
		return nil, 0, p.mapError(err)
	}
	var total int64
	if err := p.Dt.QueryRow(ctx, count, countValues...).Scan(&total); err != nil {
		return nil, 0, p.mapError(err)
	}

	rows, err := p.Dt.Query(ctx, query, values...)
	if err != nil {
		return nil, 0, p.mapError(err)
	}
	defer rows.Close()
	items, err := infrastructurerepositoryshared.ScanGoogleAuthList(rows)
	if err != nil {
		return nil, 0, p.mapError(err)
	}

	return items, total, nil
}

func (p *postgresImpl) UpdateById(ctx context.Context, id int64, args domaincontractsrepository.UpdateGoogleAuthArgs) error {
	if err := infrastructurerepositoryshared.PositiveIds(id); err != nil {
		return err
	}

	if err := infrastructurerepositoryshared.OptionalIds(args.By); err != nil {
		return err
	}

	if err := infrastructurerepositoryshared.ValidateJSON(args.Preferences); err != nil {
		return err
	}
	err := p.Dt.WithTx(ctx, func(txCtx context.Context) error {
		item, err := p.ReadById(txCtx, id)
		if err != nil {
			return err
		}

		if err := p.lockParents(txCtx, item.UserId); err != nil {
			return err
		}

		query, values, err := p.queryUpdateById(id, args)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build mutation", err)
		}

		result, err := p.Dt.Exec(txCtx, query, values...)
		if err != nil {
			return p.mapError(err)
		}

		if result.RowsAffected() == 0 {
			return infrastructurerepositoryshared.NotFound("google auth not found", domainmodelserror.GoogleAuthNotFound)
		}

		return nil
	})
	return p.mapError(err)
}

func (p *postgresImpl) DeleteById(ctx context.Context, id int64, by *int64) error {
	if err := infrastructurerepositoryshared.PositiveIds(id); err != nil {
		return err
	}

	if err := infrastructurerepositoryshared.OptionalIds(by); err != nil {
		return err
	}
	err := p.Dt.WithTx(ctx, func(txCtx context.Context) error {
		item, err := p.ReadById(txCtx, id)
		if err != nil {
			return err
		}

		if err := p.lockParents(txCtx, item.UserId); err != nil {
			return err
		}

		query, values, err := p.queryDeleteById(id, by)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build mutation", err)
		}

		result, err := p.Dt.Exec(txCtx, query, values...)
		if err != nil {
			return p.mapError(err)
		}

		if result.RowsAffected() == 0 {
			return infrastructurerepositoryshared.NotFound("google auth not found", domainmodelserror.GoogleAuthNotFound)
		}

		return nil
	})
	return p.mapError(err)
}
