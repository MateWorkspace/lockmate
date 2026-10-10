package infrastructurerepositoryrolepermission

import (
	"context"

	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructurerepositoryshared "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/shared"
	"github.com/MateWorkspace/mate-golang/pgxdt"
)

type postgresImpl struct {
	infrastructurerepositoryshared.BasePostgres
}

func NewPostgresImpl(dt pgxdt.Pgxdt, sqrQuestion, sqrDollar *squirrel.StatementBuilderType) domaincontractsrepository.RolePermission {
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
		"assignment operation failed",
		err,
		infrastructurerepositoryshared.ConflictMatch{
			Constraint: "uq_role_permission_space_role_permission",
			Type:       domainmodelserror.RolePermissionConflict,
		},
	)
}

func (p *postgresImpl) Create(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateRolePermissionArgs) (id int64, err error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, args.RoleId, args.PermissionId); err != nil {
		return 0, err
	}

	if err := infrastructurerepositoryshared.OptionalIds(args.By); err != nil {
		return 0, err
	}
	err = p.Dt.WithTx(ctx, func(txCtx context.Context) error {
		if err := infrastructurerepositoryshared.LockRow(txCtx, p.Dt, "spaces", spaceId, nil, domainmodelserror.SpaceNotFound); err != nil {
			return err
		}

		if err := infrastructurerepositoryshared.LockRow(txCtx, p.Dt, "roles", args.RoleId, &spaceId, domainmodelserror.RoleNotFound); err != nil {
			return err
		}

		if err := infrastructurerepositoryshared.LockRow(txCtx, p.Dt, "permissions", args.PermissionId, &spaceId, domainmodelserror.PermissionNotFound); err != nil {
			return err
		}

		query, values, err := p.queryCreate(spaceId, args)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build assignment creation", err)
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

func (p *postgresImpl) ReadById(ctx context.Context, spaceId int64, id int64) (*domaincontractsrepository.RolePermissionDetails, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, id); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadById(spaceId, id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build assignment read", err)
	}

	item, err := infrastructurerepositoryshared.ScanRolePermissionDetails(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("assignment not found", err, domainmodelserror.RolePermissionNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByRoleIdAndPermissionId(ctx context.Context, spaceId int64, leftId, rightId int64) (*domaincontractsrepository.RolePermissionDetails, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, leftId, rightId); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadPair(spaceId, leftId, rightId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build assignment read", err)
	}

	item, err := infrastructurerepositoryshared.ScanRolePermissionDetails(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("assignment not found", err, domainmodelserror.RolePermissionNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByRoleId(ctx context.Context, spaceId, value int64) ([]domaincontractsrepository.RolePermissionWithPermission, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, value); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadLeft(spaceId, value)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build assignment list", err)
	}

	rows, err := p.Dt.Query(ctx, query, values...)
	if err != nil {
		return nil, p.mapError(err)
	}
	defer rows.Close()
	items, err := infrastructurerepositoryshared.ScanRolePermissionWithPermissionList(rows)
	if err != nil {
		return nil, p.mapError(err)
	}

	return items, nil
}

func (p *postgresImpl) ReadByPermissionId(ctx context.Context, spaceId, value int64) ([]domaincontractsrepository.RolePermissionWithRole, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, value); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadRight(spaceId, value)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build assignment list", err)
	}

	rows, err := p.Dt.Query(ctx, query, values...)
	if err != nil {
		return nil, p.mapError(err)
	}
	defer rows.Close()
	items, err := infrastructurerepositoryshared.ScanRolePermissionWithRoleList(rows)
	if err != nil {
		return nil, p.mapError(err)
	}

	return items, nil
}

func (p *postgresImpl) DeleteById(ctx context.Context, spaceId, id int64) error {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, id); err != nil {
		return err
	}

	query, values, err := p.queryDeleteById(spaceId, id)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build assignment deletion", err)
	}

	result, err := p.Dt.Exec(ctx, query, values...)
	if err != nil {
		return p.mapError(err)
	}

	if result.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("assignment not found", domainmodelserror.RolePermissionNotFound)
	}

	return nil
}

func (p *postgresImpl) DeleteByRoleIdOrPermissionId(ctx context.Context, spaceId int64, args domaincontractsrepository.DeleteRolePermissionArgs) error {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId); err != nil {
		return err
	}

	if args.RoleId == nil && args.PermissionId == nil {
		return domainmodelserror.NewError("at least one deletion selector is required", domainmodelserror.BadArgs, nil)
	}

	if err := infrastructurerepositoryshared.OptionalIds(args.RoleId, args.PermissionId); err != nil {
		return err
	}

	query, values, err := p.queryDeleteByFilter(spaceId, args)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build bulk deletion", err)
	}
	_, err = p.Dt.Exec(ctx, query, values...)
	return p.mapError(err)
}
