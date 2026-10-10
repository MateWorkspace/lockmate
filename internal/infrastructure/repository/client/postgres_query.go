package infrastructurerepositoryclient

import (
	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	infrastructurerepositoryshared "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/shared"
)

var clientColumns = []string{
	"id",
	"space_id",
	"slug",
	"name",
	"description",
	"secret_key_hash",
	"secret_key_redacted",
	"is_active",
	"preferences",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

var summaryColumns = []string{
	"id",
	"space_id",
	"slug",
	"name",
	"description",
	"secret_key_redacted",
	"is_active",
	"preferences",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

func (p *postgresImpl) querySelect(spaceId int64) squirrel.SelectBuilder {
	q := p.SqrD.Select(infrastructurerepositoryshared.QualifiedColumns("e", clientColumns)...).From("clients e").Where("e.deleted_at IS NULL")
	q = q.Where("EXISTS (SELECT 1 FROM spaces s WHERE s.id=e.space_id AND s.deleted_at IS NULL)")
	if spaceId > 0 {
		q = q.Where(squirrel.Eq{"e.space_id": spaceId})
	}

	return q
}

func (p *postgresImpl) queryReadById(spaceId, id int64) (string, []any, error) {
	return p.querySelect(spaceId).Where(squirrel.Eq{"e.id": id}).ToSql()
}

func (p *postgresImpl) queryReadBySlug(spaceId int64, value string) (string, []any, error) {
	return p.querySelect(spaceId).Where(squirrel.Eq{"e.slug": value}).ToSql()
}

func (p *postgresImpl) queryCreate(spaceId int64, args domaincontractsrepository.CreateClientArgs) (string, []any, error) {
	columns := []string{"created_by"}
	values := []any{args.By}
	columns = append(columns, "space_id")
	values = append(values, spaceId)
	columns = append(columns, "slug")
	values = append(values, args.Slug)
	columns = append(columns, "name")
	values = append(values, args.Name)
	if args.Description != nil {
		columns = append(columns, "description")
		values = append(values, *args.Description)
	}
	columns = append(columns, "secret_key_hash")
	values = append(values, args.SecretKeyHash)
	columns = append(columns, "secret_key_redacted")
	values = append(values, args.SecretKeyRedacted)
	if args.IsActive != nil {
		columns = append(columns, "is_active")
		values = append(values, *args.IsActive)
	}

	if args.Preferences != nil {
		columns = append(columns, "preferences")
		values = append(values, *args.Preferences)
	}

	return p.SqrD.Insert("clients").Columns(columns...).Values(values...).Suffix("RETURNING id").ToSql()
}

func (p *postgresImpl) queryUpdateById(spaceId, id int64, args domaincontractsrepository.UpdateClientArgs) (string, []any, error) {
	q := p.SqrD.Update("clients e").Where(squirrel.Eq{"e.id": id}).Where("e.deleted_at IS NULL")
	q = q.Where("EXISTS (SELECT 1 FROM spaces s WHERE s.id=e.space_id AND s.deleted_at IS NULL)")
	q = q.Where(squirrel.Eq{"e.space_id": spaceId})
	if args.Name != nil {
		q = q.Set("name", *args.Name)
	}

	if args.Description != nil {
		q = q.Set("description", *args.Description)
	}

	if args.SecretKeyHash != nil {
		q = q.Set("secret_key_hash", *args.SecretKeyHash)
	}

	if args.SecretKeyRedacted != nil {
		q = q.Set("secret_key_redacted", *args.SecretKeyRedacted)
	}

	if args.IsActive != nil {
		q = q.Set("is_active", *args.IsActive)
	}

	if args.Preferences != nil {
		q = q.Set("preferences", *args.Preferences)
	}

	return q.Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).Set("updated_by", args.By).ToSql()
}

func (p *postgresImpl) queryDeleteById(spaceId, id int64, by *int64) (string, []any, error) {
	q := p.SqrD.Update("clients e").Where(squirrel.Eq{"e.id": id}).Where("e.deleted_at IS NULL")
	q = q.Where("EXISTS (SELECT 1 FROM spaces s WHERE s.id=e.space_id AND s.deleted_at IS NULL)")
	q = q.Where(squirrel.Eq{"e.space_id": spaceId})
	return q.Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).Set("deleted_by", by).ToSql()
}

func (p *postgresImpl) queryFilter(spaceId int64, filter domaincontractsrepository.ClientFilter) (string, []any, string, []any, error) {
	limit, offset, err := infrastructurerepositoryshared.Pagination(filter.Page, filter.Limit)
	if err != nil {
		return "", nil, "", nil, err
	}
	q := p.querySelect(spaceId)
	q = infrastructurerepositoryshared.ApplySearch(q, filter.Search, "e.slug", "e.name", "e.description")
	if filter.IsActive != nil {
		q = q.Where(squirrel.Eq{"e.is_active": *filter.IsActive})
	}

	count, countArgs, err := q.RemoveColumns().Column("COUNT(*)").ToSql()
	if err != nil {
		return "", nil, "", nil, err
	}
	q = q.RemoveColumns().Columns(infrastructurerepositoryshared.QualifiedColumns("e", summaryColumns)...)
	query, args, err := q.OrderBy("e.created_at DESC", "e.id ASC").Limit(limit).Offset(offset).ToSql()
	return count, countArgs, query, args, err
}
