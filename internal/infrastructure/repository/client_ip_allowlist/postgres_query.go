package infrastructurerepositoryclientipallowlist

import (
	"net/netip"

	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	infrastructurerepositoryshared "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/shared"
)

var clientIpAllowlistColumns = []string{
	"id",
	"client_id",
	"cidr",
	"is_active",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

const liveParents = "EXISTS (SELECT 1 FROM clients c JOIN spaces s ON s.id=c.space_id WHERE c.id=e.client_id AND c.space_id=? AND c.deleted_at IS NULL AND s.deleted_at IS NULL)"

func (p *postgresImpl) querySelect(spaceId int64) squirrel.SelectBuilder {
	return p.SqrD.Select(infrastructurerepositoryshared.QualifiedColumns("e", clientIpAllowlistColumns)...).
		From("client_ip_allowlist e").Where("e.deleted_at IS NULL").Where(liveParents, spaceId)
}

func (p *postgresImpl) queryReadById(spaceId, id int64) (string, []any, error) {
	return p.querySelect(spaceId).Where(squirrel.Eq{"e.id": id}).ToSql()
}

func (p *postgresImpl) queryReadByClientIdAndCidr(spaceId, clientId int64, cidr string) (string, []any, error) {
	return p.querySelect(spaceId).Where(squirrel.Eq{"e.client_id": clientId, "e.cidr": cidr}).ToSql()
}

func (p *postgresImpl) queryReadByClientId(spaceId, clientId int64) (string, []any, error) {
	return p.querySelect(spaceId).Where(squirrel.Eq{"e.client_id": clientId}).
		OrderBy("e.created_at DESC", "e.id ASC").ToSql()
}

func (p *postgresImpl) queryCreate(args domaincontractsrepository.CreateClientIpAllowlistArgs) (string, []any, error) {
	columns := []string{"client_id", "cidr", "created_by"}
	values := []any{args.ClientId, args.Cidr, args.By}
	if args.IsActive != nil {
		columns = append(columns, "is_active")
		values = append(values, *args.IsActive)
	}

	return p.SqrD.Insert("client_ip_allowlist").Columns(columns...).Values(values...).Suffix("RETURNING id").ToSql()
}

func (p *postgresImpl) queryUpdateById(spaceId, id int64, args domaincontractsrepository.UpdateClientIpAllowlistArgs) (string, []any, error) {
	q := p.SqrD.Update("client_ip_allowlist e").Where(squirrel.Eq{"e.id": id}).
		Where("e.deleted_at IS NULL").Where(liveParents, spaceId)
	if args.Cidr != nil {
		q = q.Set("cidr", *args.Cidr)
	}

	if args.IsActive != nil {
		q = q.Set("is_active", *args.IsActive)
	}

	return q.Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).Set("updated_by", args.By).ToSql()
}

func (p *postgresImpl) queryDeleteById(spaceId, id int64, by *int64) (string, []any, error) {
	return p.SqrD.Update("client_ip_allowlist e").Where(squirrel.Eq{"e.id": id}).
		Where("e.deleted_at IS NULL").Where(liveParents, spaceId).
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).Set("deleted_by", by).ToSql()
}

func (p *postgresImpl) queryFilter(spaceId int64, filter domaincontractsrepository.ClientIpAllowlistFilter) (string, []any, string, []any, error) {
	limit, offset, err := infrastructurerepositoryshared.Pagination(filter.Page, filter.Limit)
	if err != nil {
		return "", nil, "", nil, err
	}

	q := infrastructurerepositoryshared.ApplySearch(p.querySelect(spaceId), filter.Search, "e.cidr")
	if filter.ClientId != nil {
		q = q.Where(squirrel.Eq{"e.client_id": *filter.ClientId})
	}

	if filter.IsActive != nil {
		q = q.Where(squirrel.Eq{"e.is_active": *filter.IsActive})
	}

	count, countArgs, err := q.RemoveColumns().Column("COUNT(*)").ToSql()
	if err != nil {
		return "", nil, "", nil, err
	}

	query, args, err := q.OrderBy("e.created_at DESC", "e.id ASC").Limit(limit).Offset(offset).ToSql()
	return count, countArgs, query, args, err
}

func (p *postgresImpl) queryIsAllowedForClient(spaceId, clientId int64, ip netip.Addr) (string, []any, error) {
	return p.SqrD.Select("1").From("client_ip_allowlist e").
		Join("clients c ON c.id=e.client_id").Join("spaces s ON s.id=c.space_id").
		Where(squirrel.Eq{"e.client_id": clientId, "c.space_id": spaceId, "e.is_active": true, "c.is_active": true, "s.is_active": true}).
		Where("e.deleted_at IS NULL AND c.deleted_at IS NULL AND s.deleted_at IS NULL").
		Where(squirrel.Expr("(?::inet <<= e.cidr::cidr)", ip.String())).
		Suffix("LIMIT 1").
		ToSql()
}
