package infrastructurerepositoryclassicauth

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

func NewPostgresImpl(dt pgxdt.Pgxdt, sqrQuestion, sqrDollar *squirrel.StatementBuilderType) domaincontractsrepository.ClassicAuth {
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
		"classic auth operation failed",
		err,
		infrastructurerepositoryshared.ConflictMatch{
			Constraint: "uq_classic_auths_user_id",
			Type:       domainmodelserror.ClassicAuthUserConflict,
		},
		infrastructurerepositoryshared.ConflictMatch{
			Constraint: "uq_classic_auths_username",
			Type:       domainmodelserror.ClassicAuthUsernameConflict,
		},
		infrastructurerepositoryshared.ConflictMatch{
			Constraint: "uq_classic_auths_email",
			Type:       domainmodelserror.ClassicAuthEmailConflict,
		},
		infrastructurerepositoryshared.ConflictMatch{
			Constraint: "uq_classic_auths_phone",
			Type:       domainmodelserror.ClassicAuthPhoneConflict,
		},
	)
}

func (p *postgresImpl) lockParents(ctx context.Context, userId int64) error {
	if err := infrastructurerepositoryshared.LockRow(ctx, p.Dt, "users", userId, nil, domainmodelserror.UserNotFound); err != nil {
		return err
	}

	return nil
}

func (p *postgresImpl) Create(ctx context.Context, args domaincontractsrepository.CreateClassicAuthArgs) (id int64, err error) {
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

func (p *postgresImpl) ReadById(ctx context.Context, id int64) (*domainmodelsentity.ClassicAuth, error) {
	if err := infrastructurerepositoryshared.PositiveIds(id); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadById(id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read query", err)
	}

	item, err := infrastructurerepositoryshared.ScanClassicAuth(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("classic auth not found", err, domainmodelserror.ClassicAuthNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByUserId(ctx context.Context, value int64) (*domainmodelsentity.ClassicAuth, error) {
	if err := infrastructurerepositoryshared.PositiveIds(value); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadByUserId(value)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read query", err)
	}

	item, err := infrastructurerepositoryshared.ScanClassicAuth(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("classic auth not found", err, domainmodelserror.ClassicAuthNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByUsername(ctx context.Context, value string) (*domainmodelsentity.ClassicAuth, error) {
	query, values, err := p.queryReadByUsername(value)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read query", err)
	}

	item, err := infrastructurerepositoryshared.ScanClassicAuth(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("classic auth not found", err, domainmodelserror.ClassicAuthNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByEmail(ctx context.Context, value string) (*domainmodelsentity.ClassicAuth, error) {
	query, values, err := p.queryReadByEmail(value)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read query", err)
	}

	item, err := infrastructurerepositoryshared.ScanClassicAuth(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("classic auth not found", err, domainmodelserror.ClassicAuthNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByPhone(ctx context.Context, value string) (*domainmodelsentity.ClassicAuth, error) {
	query, values, err := p.queryReadByPhone(value)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read query", err)
	}

	item, err := infrastructurerepositoryshared.ScanClassicAuth(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("classic auth not found", err, domainmodelserror.ClassicAuthNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByFilter(ctx context.Context, filter domaincontractsrepository.ClassicAuthFilter) ([]domaincontractsrepository.ClassicAuthSummary, int64, error) {
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
	items, err := infrastructurerepositoryshared.ScanClassicAuthSummaryList(rows)
	if err != nil {
		return nil, 0, p.mapError(err)
	}

	return items, total, nil
}

func (p *postgresImpl) UpdateById(ctx context.Context, id int64, args domaincontractsrepository.UpdateClassicAuthArgs) error {
	if err := infrastructurerepositoryshared.PositiveIds(id); err != nil {
		return err
	}

	if err := infrastructurerepositoryshared.OptionalIds(args.By); err != nil {
		return err
	}

	if err := infrastructurerepositoryshared.ValidateJSON(args.Preferences); err != nil {
		return err
	}

	if args.Email != nil && args.EmailSetNull {
		return domainmodelserror.NewError("value and clear flag cannot both be supplied", domainmodelserror.BadArgs, nil)
	}

	if args.Phone != nil && args.PhoneSetNull {
		return domainmodelserror.NewError("value and clear flag cannot both be supplied", domainmodelserror.BadArgs, nil)
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
			return infrastructurerepositoryshared.NotFound("classic auth not found", domainmodelserror.ClassicAuthNotFound)
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
			return infrastructurerepositoryshared.NotFound("classic auth not found", domainmodelserror.ClassicAuthNotFound)
		}

		return nil
	})
	return p.mapError(err)
}
