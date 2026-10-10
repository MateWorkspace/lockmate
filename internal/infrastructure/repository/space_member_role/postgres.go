package infrastructurerepositoryspacememberrole

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

func NewPostgresImpl(dt pgxdt.Pgxdt, sqrQuestion, sqrDollar *squirrel.StatementBuilderType) domaincontractsrepository.SpaceMemberRole {
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
			Constraint: "uq_space_member_role_space_member_role",
			Type:       domainmodelserror.SpaceMemberRoleConflict,
		},
	)
}

func (p *postgresImpl) Create(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateSpaceMemberRoleArgs) (id int64, err error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, args.SpaceMemberId, args.RoleId); err != nil {
		return 0, err
	}

	if err := infrastructurerepositoryshared.OptionalIds(args.By); err != nil {
		return 0, err
	}
	err = p.Dt.WithTx(ctx, func(txCtx context.Context) error {
		query, values, err := p.queryMemberUser(spaceId, args.SpaceMemberId)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build member lookup", err)
		}
		var userId int64
		if err := p.Dt.QueryRow(txCtx, query, values...).Scan(&userId); err != nil {
			return infrastructurerepositoryshared.ReadError("member not found", err, domainmodelserror.SpaceMemberNotFound)
		}

		if err := infrastructurerepositoryshared.LockRow(txCtx, p.Dt, "users", userId, nil, domainmodelserror.UserNotFound); err != nil {
			return err
		}

		if err := infrastructurerepositoryshared.LockRow(txCtx, p.Dt, "spaces", spaceId, nil, domainmodelserror.SpaceNotFound); err != nil {
			return err
		}

		if err := infrastructurerepositoryshared.LockRow(txCtx, p.Dt, "space_members", args.SpaceMemberId, &spaceId, domainmodelserror.SpaceMemberNotFound); err != nil {
			return err
		}

		if err := infrastructurerepositoryshared.LockRow(txCtx, p.Dt, "roles", args.RoleId, &spaceId, domainmodelserror.RoleNotFound); err != nil {
			return err
		}

		query, values, err = p.queryCreate(spaceId, args)
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

func (p *postgresImpl) ReadById(ctx context.Context, spaceId int64, id int64) (*domaincontractsrepository.SpaceMemberRoleDetails, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, id); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadById(spaceId, id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build assignment read", err)
	}

	item, err := infrastructurerepositoryshared.ScanSpaceMemberRoleDetails(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("assignment not found", err, domainmodelserror.SpaceMemberRoleNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadBySpaceMemberIdAndRoleId(ctx context.Context, spaceId int64, leftId, rightId int64) (*domaincontractsrepository.SpaceMemberRoleDetails, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, leftId, rightId); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadPair(spaceId, leftId, rightId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build assignment read", err)
	}

	item, err := infrastructurerepositoryshared.ScanSpaceMemberRoleDetails(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("assignment not found", err, domainmodelserror.SpaceMemberRoleNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadBySpaceMemberId(ctx context.Context, spaceId, value int64) ([]domaincontractsrepository.SpaceMemberRoleWithRole, error) {
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
	items, err := infrastructurerepositoryshared.ScanSpaceMemberRoleWithRoleList(rows)
	if err != nil {
		return nil, p.mapError(err)
	}

	return items, nil
}

func (p *postgresImpl) ReadByRoleId(ctx context.Context, spaceId, value int64) ([]domaincontractsrepository.SpaceMemberRoleWithSpaceMember, error) {
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
	items, err := infrastructurerepositoryshared.ScanSpaceMemberRoleWithSpaceMemberList(rows)
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
		return infrastructurerepositoryshared.NotFound("assignment not found", domainmodelserror.SpaceMemberRoleNotFound)
	}

	return nil
}

func (p *postgresImpl) DeleteBySpaceMemberIdOrRoleId(ctx context.Context, spaceId int64, args domaincontractsrepository.DeleteSpaceMemberRoleArgs) error {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId); err != nil {
		return err
	}

	if args.SpaceMemberId == nil && args.RoleId == nil {
		return domainmodelserror.NewError("at least one deletion selector is required", domainmodelserror.BadArgs, nil)
	}

	if err := infrastructurerepositoryshared.OptionalIds(args.SpaceMemberId, args.RoleId); err != nil {
		return err
	}

	query, values, err := p.queryDeleteByFilter(spaceId, args)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build bulk deletion", err)
	}
	_, err = p.Dt.Exec(ctx, query, values...)
	return p.mapError(err)
}

func (p *postgresImpl) eligibility(ctx context.Context, spaceId, id int64) error {
	query, values, err := p.queryEligibility(spaceId, id)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build membership eligibility", err)
	}
	var spaceActive, memberActive bool
	if err := p.Dt.QueryRow(ctx, query, values...).Scan(&spaceActive, &memberActive); err != nil {
		return infrastructurerepositoryshared.ReadError("member not found", err, domainmodelserror.SpaceMemberNotFound)
	}

	if !spaceActive || !memberActive {
		return domainmodelserror.NewError("membership access is disabled", domainmodelserror.Forbidden, nil)
	}

	return nil
}

func (p *postgresImpl) ReadEffectiveRolesBySpaceMemberId(ctx context.Context, spaceId, spaceMemberId int64) ([]domainmodelsentity.Role, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, spaceMemberId); err != nil {
		return nil, err
	}

	if err := p.eligibility(ctx, spaceId, spaceMemberId); err != nil {
		return nil, err
	}

	query, values, err := p.queryEffectiveRoles(spaceId, spaceMemberId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build effective grants", err)
	}

	rows, err := p.Dt.Query(ctx, query, values...)
	if err != nil {
		return nil, p.mapError(err)
	}
	defer rows.Close()
	items, err := infrastructurerepositoryshared.ScanRoleList(rows)
	if err != nil {
		return nil, p.mapError(err)
	}

	return items, nil
}

func (p *postgresImpl) ReadEffectivePermissionsBySpaceMemberId(ctx context.Context, spaceId, spaceMemberId int64) ([]domainmodelsentity.Permission, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, spaceMemberId); err != nil {
		return nil, err
	}

	if err := p.eligibility(ctx, spaceId, spaceMemberId); err != nil {
		return nil, err
	}

	query, values, err := p.queryEffectivePermissions(spaceId, spaceMemberId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build effective grants", err)
	}

	rows, err := p.Dt.Query(ctx, query, values...)
	if err != nil {
		return nil, p.mapError(err)
	}
	defer rows.Close()
	items, err := infrastructurerepositoryshared.ScanPermissionList(rows)
	if err != nil {
		return nil, p.mapError(err)
	}

	return items, nil
}
