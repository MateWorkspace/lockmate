package infrastructurerepositoryspacemember

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

func NewPostgresImpl(dt pgxdt.Pgxdt, sqrQuestion, sqrDollar *squirrel.StatementBuilderType) domaincontractsrepository.SpaceMember {
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
		"space member operation failed",
		err,
		infrastructurerepositoryshared.ConflictMatch{
			Constraint: "uq_space_members_space_id_user_id",
			Type:       domainmodelserror.SpaceMemberConflict,
		},
	)
}

func (p *postgresImpl) lockParents(ctx context.Context, spaceId, userId int64) error {
	if err := infrastructurerepositoryshared.LockRow(ctx, p.Dt, "users", userId, nil, domainmodelserror.UserNotFound); err != nil {
		return err
	}

	if err := infrastructurerepositoryshared.LockRow(ctx, p.Dt, "spaces", spaceId, nil, domainmodelserror.SpaceNotFound); err != nil {
		return err
	}

	if err := infrastructurerepositoryshared.LockDefault(ctx, p.Dt, spaceId); err != nil {
		return err
	}

	return nil
}

func (p *postgresImpl) Create(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateSpaceMemberArgs) (id int64, err error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, args.UserId); err != nil {
		return 0, err
	}

	if err := infrastructurerepositoryshared.OptionalIds(args.By); err != nil {
		return 0, err
	}

	if err := infrastructurerepositoryshared.ValidateJSON(args.Preferences); err != nil {
		return 0, err
	}
	err = p.Dt.WithTx(ctx, func(txCtx context.Context) error {
		if err := p.lockParents(txCtx, spaceId, args.UserId); err != nil {
			return err
		}
		var defaultId int64
		query, values, err := p.queryDefaultForShare(spaceId)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build default role lookup", err)
		}

		if err := p.Dt.QueryRow(txCtx, query, values...).Scan(&defaultId); err != nil {
			return infrastructurerepositoryshared.ReadError("default role not found", err, domainmodelserror.DefaultRoleNotFound)
		}

		query, values, err = p.queryCreate(spaceId, args)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build create query", err)
		}

		if err := p.Dt.QueryRow(txCtx, query, values...).Scan(&id); err != nil {
			return p.mapError(err)
		}

		query, values, err = p.queryAssignDefault(spaceId, id, defaultId, args.By)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build default assignment", err)
		}

		if _, err := p.Dt.Exec(txCtx, query, values...); err != nil {
			return p.mapError(err)
		}

		return nil
	})
	if err != nil {
		return 0, p.mapError(err)
	}

	return id, nil
}

func (p *postgresImpl) ReadById(ctx context.Context, spaceId int64, id int64) (*domainmodelsentity.SpaceMember, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, id); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadById(spaceId, id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read query", err)
	}

	item, err := infrastructurerepositoryshared.ScanSpaceMember(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("space member not found", err, domainmodelserror.SpaceMemberNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByUserId(ctx context.Context, spaceId int64, value int64) (*domainmodelsentity.SpaceMember, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, value); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadByUserId(spaceId, value)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read query", err)
	}

	item, err := infrastructurerepositoryshared.ScanSpaceMember(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("space member not found", err, domainmodelserror.SpaceMemberNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByFilter(ctx context.Context, spaceId int64, filter domaincontractsrepository.SpaceMemberFilter) ([]domaincontractsrepository.SpaceMemberWithUser, int64, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId); err != nil {
		return nil, 0, err
	}

	if err := infrastructurerepositoryshared.OptionalIds(filter.UserId, filter.RoleId); err != nil {
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
	items, err := infrastructurerepositoryshared.ScanSpaceMemberWithUserList(rows)
	if err != nil {
		return nil, 0, p.mapError(err)
	}

	return items, total, nil
}

func (p *postgresImpl) UpdateById(ctx context.Context, spaceId int64, id int64, args domaincontractsrepository.UpdateSpaceMemberArgs) error {
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
		item, err := p.ReadById(txCtx, spaceId, id)
		if err != nil {
			return err
		}

		if err := p.lockParents(txCtx, spaceId, item.UserId); err != nil {
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
			return infrastructurerepositoryshared.NotFound("space member not found", domainmodelserror.SpaceMemberNotFound)
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
		item, err := p.ReadById(txCtx, spaceId, id)
		if err != nil {
			return err
		}

		if err := p.lockParents(txCtx, spaceId, item.UserId); err != nil {
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
			return infrastructurerepositoryshared.NotFound("space member not found", domainmodelserror.SpaceMemberNotFound)
		}

		return nil
	})
	return p.mapError(err)
}
