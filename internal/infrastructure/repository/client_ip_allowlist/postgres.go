package infrastructurerepositoryclientipallowlist

import (
	"context"
	"errors"
	"net/netip"

	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructurerepositoryshared "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/shared"
	"github.com/MateWorkspace/mate-golang/pgxdt"
	"github.com/jackc/pgx/v5"
)

type postgresImpl struct {
	infrastructurerepositoryshared.BasePostgres
}

func NewPostgresImpl(dt pgxdt.Pgxdt, sqrQuestion, sqrDollar *squirrel.StatementBuilderType) domaincontractsrepository.ClientIpAllowlist {
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
		"client IP allowlist operation failed",
		err,
		infrastructurerepositoryshared.ConflictMatch{
			Constraint: "uq_client_ip_allowlist_client_id_cidr",
			Type:       domainmodelserror.ClientIpAllowlistConflict,
		},
	)
}

func (p *postgresImpl) lockParents(ctx context.Context, spaceId, clientId int64) error {
	if err := infrastructurerepositoryshared.LockRow(ctx, p.Dt, "spaces", spaceId, nil, domainmodelserror.SpaceNotFound); err != nil {
		return err
	}

	return infrastructurerepositoryshared.LockRow(ctx, p.Dt, "clients", clientId, &spaceId, domainmodelserror.ClientNotFound)
}

func (p *postgresImpl) Create(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateClientIpAllowlistArgs) (id int64, err error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, args.ClientId); err != nil {
		return 0, err
	}

	if err := infrastructurerepositoryshared.OptionalIds(args.By); err != nil {
		return 0, err
	}

	err = p.Dt.WithTx(ctx, func(txCtx context.Context) error {
		if err := p.lockParents(txCtx, spaceId, args.ClientId); err != nil {
			return err
		}

		query, values, err := p.queryCreate(args)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build create client IP allowlist query", err)
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

func (p *postgresImpl) ReadById(ctx context.Context, spaceId, id int64) (*domainmodelsentity.ClientIpAllowlist, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, id); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadById(spaceId, id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read client IP allowlist query", err)
	}

	item, err := infrastructurerepositoryshared.ScanClientIpAllowlist(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("client IP allowlist entry not found", err, domainmodelserror.ClientIpAllowlistNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByClientIdAndCidr(ctx context.Context, spaceId, clientId int64, cidr string) (*domainmodelsentity.ClientIpAllowlist, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, clientId); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadByClientIdAndCidr(spaceId, clientId, cidr)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read client IP allowlist query", err)
	}

	item, err := infrastructurerepositoryshared.ScanClientIpAllowlist(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("client IP allowlist entry not found", err, domainmodelserror.ClientIpAllowlistNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByClientId(ctx context.Context, spaceId, clientId int64) ([]domainmodelsentity.ClientIpAllowlist, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, clientId); err != nil {
		return nil, err
	}

	query, values, err := p.queryReadByClientId(spaceId, clientId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build list client IP allowlist query", err)
	}

	rows, err := p.Dt.Query(ctx, query, values...)
	if err != nil {
		return nil, p.mapError(err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanClientIpAllowlistList(rows)
	if err != nil {
		return nil, p.mapError(err)
	}

	return items, nil
}

func (p *postgresImpl) ReadByFilter(ctx context.Context, spaceId int64, filter domaincontractsrepository.ClientIpAllowlistFilter) ([]domainmodelsentity.ClientIpAllowlist, int64, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId); err != nil {
		return nil, 0, err
	}

	if err := infrastructurerepositoryshared.OptionalIds(filter.ClientId); err != nil {
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

	items, err := infrastructurerepositoryshared.ScanClientIpAllowlistList(rows)
	if err != nil {
		return nil, 0, p.mapError(err)
	}

	return items, total, nil
}

func (p *postgresImpl) UpdateById(ctx context.Context, spaceId, id int64, args domaincontractsrepository.UpdateClientIpAllowlistArgs) error {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, id); err != nil {
		return err
	}

	if err := infrastructurerepositoryshared.OptionalIds(args.By); err != nil {
		return err
	}

	err := p.Dt.WithTx(ctx, func(txCtx context.Context) error {
		item, err := p.ReadById(txCtx, spaceId, id)
		if err != nil {
			return err
		}

		if err := p.lockParents(txCtx, spaceId, item.ClientId); err != nil {
			return err
		}

		query, values, err := p.queryUpdateById(spaceId, id, args)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build update client IP allowlist query", err)
		}

		result, err := p.Dt.Exec(txCtx, query, values...)
		if err != nil {
			return p.mapError(err)
		}

		if result.RowsAffected() == 0 {
			return infrastructurerepositoryshared.NotFound("client IP allowlist entry not found", domainmodelserror.ClientIpAllowlistNotFound)
		}

		return nil
	})

	return p.mapError(err)
}

func (p *postgresImpl) DeleteById(ctx context.Context, spaceId, id int64, by *int64) error {
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

		if err := p.lockParents(txCtx, spaceId, item.ClientId); err != nil {
			return err
		}

		query, values, err := p.queryDeleteById(spaceId, id, by)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build delete client IP allowlist query", err)
		}

		result, err := p.Dt.Exec(txCtx, query, values...)
		if err != nil {
			return p.mapError(err)
		}

		if result.RowsAffected() == 0 {
			return infrastructurerepositoryshared.NotFound("client IP allowlist entry not found", domainmodelserror.ClientIpAllowlistNotFound)
		}

		return nil
	})

	return p.mapError(err)
}

func (p *postgresImpl) IsAllowedForClient(ctx context.Context, spaceId, clientId int64, ip netip.Addr) (bool, error) {
	if err := infrastructurerepositoryshared.PositiveIds(spaceId, clientId); err != nil {
		return false, err
	}

	if !ip.IsValid() {
		return false, domainmodelserror.NewError("client IP is invalid", domainmodelserror.BadArgs, nil)
	}

	query, values, err := p.queryIsAllowedForClient(spaceId, clientId, ip)
	if err != nil {
		return false, infrastructurerepositoryshared.QueryBuildError("failed to build client IP allowlist query", err)
	}

	var found int
	if err := p.Dt.QueryRow(ctx, query, values...).Scan(&found); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, p.mapError(err)
	}

	return true, nil
}
