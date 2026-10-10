package infrastructurerepositoryuser

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

func NewPostgresImpl(dt pgxdt.Pgxdt, sqrQuestion, sqrDollar *squirrel.StatementBuilderType) domaincontractsrepository.User {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) mapError(err error) error {
	return infrastructurerepositoryshared.MapPgxError("user operation failed", err)
}

func (p *postgresImpl) Create(ctx context.Context, args domaincontractsrepository.CreateUserArgs) (id int64, err error) {
	if err := infrastructurerepositoryshared.OptionalIds(args.By); err != nil {
		return 0, err
	}

	if err := infrastructurerepositoryshared.ValidateJSON(args.Preferences); err != nil {
		return 0, err
	}
	err = p.Dt.WithTx(ctx, func(txCtx context.Context) error {
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

func (p *postgresImpl) ReadById(ctx context.Context, id int64) (*domainmodelsentity.User, error) {
	if err := infrastructurerepositoryshared.PositiveIds(id); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadById(id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read query", err)
	}

	item, err := infrastructurerepositoryshared.ScanUser(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("user not found", err, domainmodelserror.UserNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByFilter(ctx context.Context, filter domaincontractsrepository.UserFilter) ([]domainmodelsentity.User, int64, error) {
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
	items, err := infrastructurerepositoryshared.ScanUserList(rows)
	if err != nil {
		return nil, 0, p.mapError(err)
	}

	return items, total, nil
}

func (p *postgresImpl) UpdateById(ctx context.Context, id int64, args domaincontractsrepository.UpdateUserArgs) error {
	if err := infrastructurerepositoryshared.PositiveIds(id); err != nil {
		return err
	}

	if err := infrastructurerepositoryshared.OptionalIds(args.By); err != nil {
		return err
	}

	if err := infrastructurerepositoryshared.ValidateJSON(args.Preferences); err != nil {
		return err
	}

	if args.AvatarPath != nil && args.AvatarPathSetNull {
		return domainmodelserror.NewError("value and clear flag cannot both be supplied", domainmodelserror.BadArgs, nil)
	}
	err := p.Dt.WithTx(ctx, func(txCtx context.Context) error {
		query, values, err := p.queryUpdateById(id, args)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build mutation", err)
		}

		result, err := p.Dt.Exec(txCtx, query, values...)
		if err != nil {
			return p.mapError(err)
		}

		if result.RowsAffected() == 0 {
			return infrastructurerepositoryshared.NotFound("user not found", domainmodelserror.UserNotFound)
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
		query, values, err := p.queryLockUser(id)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build user lock", err)
		}
		var found int64
		if err := p.Dt.QueryRow(txCtx, query, values...).Scan(&found); err != nil {
			return infrastructurerepositoryshared.ReadError("user not found", err, domainmodelserror.UserNotFound)
		}
		for _, table := range []string{"classic_auths", "google_auths", "space_members"} {
			query, values, err := p.queryDeleteDependent(table, id, by)
			if err != nil {
				return infrastructurerepositoryshared.QueryBuildError("failed to build dependent deletion", err)
			}

			if _, err := p.Dt.Exec(txCtx, query, values...); err != nil {
				return p.mapError(err)
			}
		}

		query, values, err = p.queryDeleteAssignments(id)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build assignment deletion", err)
		}

		if _, err := p.Dt.Exec(txCtx, query, values...); err != nil {
			return p.mapError(err)
		}

		query, values, err = p.queryDeleteById(id, by)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build mutation", err)
		}

		result, err := p.Dt.Exec(txCtx, query, values...)
		if err != nil {
			return p.mapError(err)
		}

		if result.RowsAffected() == 0 {
			return infrastructurerepositoryshared.NotFound("user not found", domainmodelserror.UserNotFound)
		}

		return nil
	})
	return p.mapError(err)
}
