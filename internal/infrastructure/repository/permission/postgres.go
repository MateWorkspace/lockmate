package infrastructurerepositorypermission

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

func NewPostgresImpl(dt pgxdt.Pgxdt, sqrQuestion, sqrDollar *squirrel.StatementBuilderType) domaincontractsrepository.Permission {
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
		"permission operation failed",
		err,
		infrastructurerepositoryshared.ConflictMatch{
			Constraint: "uq_permissions_space_id_slug",
			Type:       domainmodelserror.PermissionSlugConflict,
		},
	)
}

func (p *postgresImpl) lockParents(ctx context.Context, spaceId int64) error {
	if err := infrastructurerepositoryshared.LockRow(ctx, p.Dt, "spaces", spaceId, nil, domainmodelserror.SpaceNotFound); err != nil {
		return err
	}

	return nil
}

func (p *postgresImpl) Create(ctx context.Context, spaceId int64, args domaincontractsrepository.CreatePermissionArgs) (id int64, err error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId); err != nil {
		return 0, err
	}

	if err := infrastructurerepositoryshared.OptionalIds(args.By); err != nil {
		return 0, err
	}

	if err := infrastructurerepositoryshared.ValidateJSON(args.Preferences); err != nil {
		return 0, err
	}
	err = p.Dt.WithTx(ctx, func(txCtx context.Context) error {
		if err := p.lockParents(txCtx, spaceId); err != nil {
			return err
		}

		query, values, err := p.queryCreate(spaceId, args)
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

func (p *postgresImpl) ReadById(ctx context.Context, spaceId int64, id int64) (*domainmodelsentity.Permission, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, id); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadById(spaceId, id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPermission(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("permission not found", err, domainmodelserror.PermissionNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadBySlug(ctx context.Context, spaceId int64, value string) (*domainmodelsentity.Permission, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadBySlug(spaceId, value)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPermission(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("permission not found", err, domainmodelserror.PermissionNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByFilter(ctx context.Context, spaceId int64, filter domaincontractsrepository.PermissionFilter) ([]domainmodelsentity.Permission, int64, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId); err != nil {
		return nil, 0, err
	}

	count, countValues, query, values, err := p.queryFilter(spaceId, filter)
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
	items, err := infrastructurerepositoryshared.ScanPermissionList(rows)
	if err != nil {
		return nil, 0, p.mapError(err)
	}

	return items, total, nil
}

func (p *postgresImpl) UpdateById(ctx context.Context, spaceId int64, id int64, args domaincontractsrepository.UpdatePermissionArgs) error {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, id); err != nil {
		return err
	}

	if err := infrastructurerepositoryshared.OptionalIds(args.By); err != nil {
		return err
	}

	if err := infrastructurerepositoryshared.ValidateJSON(args.Preferences); err != nil {
		return err
	}
	err := p.Dt.WithTx(ctx, func(txCtx context.Context) error {
		if err := p.lockParents(txCtx, spaceId); err != nil {
			return err
		}

		query, values, err := p.queryUpdateById(spaceId, id, args)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build mutation", err)
		}

		result, err := p.Dt.Exec(txCtx, query, values...)
		if err != nil {
			return p.mapError(err)
		}

		if result.RowsAffected() == 0 {
			return infrastructurerepositoryshared.NotFound("permission not found", domainmodelserror.PermissionNotFound)
		}

		return nil
	})
	return p.mapError(err)
}

func (p *postgresImpl) DeleteById(ctx context.Context, spaceId int64, id int64, by *int64) error {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, id); err != nil {
		return err
	}

	if err := infrastructurerepositoryshared.OptionalIds(by); err != nil {
		return err
	}
	err := p.Dt.WithTx(ctx, func(txCtx context.Context) error {
		if err := p.lockParents(txCtx, spaceId); err != nil {
			return err
		}

		query, values, err := p.queryDeleteById(spaceId, id, by)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build mutation", err)
		}

		result, err := p.Dt.Exec(txCtx, query, values...)
		if err != nil {
			return p.mapError(err)
		}

		if result.RowsAffected() == 0 {
			return infrastructurerepositoryshared.NotFound("permission not found", domainmodelserror.PermissionNotFound)
		}

		return nil
	})
	return p.mapError(err)
}
