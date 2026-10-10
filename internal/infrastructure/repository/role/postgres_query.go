package infrastructurerepositoryrole

import (
	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	infrastructurerepositoryshared "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/shared"
)

var roleColumns = []string{
	"id",
	"space_id",
	"slug",
	"name",
	"description",
	"is_default",
	"preferences",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

func (p *postgresImpl) querySelect(spaceId int64) squirrel.SelectBuilder {
	q := p.SqrD.Select(infrastructurerepositoryshared.QualifiedColumns("e", roleColumns)...).From("roles e").Where("e.deleted_at IS NULL")
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

func (p *postgresImpl) queryCreate(spaceId int64, args domaincontractsrepository.CreateRoleArgs) (string, []any, error) {
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

	if args.IsDefault != nil {
		columns = append(columns, "is_default")
		values = append(values, *args.IsDefault)
	}

	if args.Preferences != nil {
		columns = append(columns, "preferences")
		values = append(values, *args.Preferences)
	}

	return p.SqrD.Insert("roles").Columns(columns...).Values(values...).Suffix("RETURNING id").ToSql()
}

func (p *postgresImpl) queryUpdateById(spaceId, id int64, args domaincontractsrepository.UpdateRoleArgs) (string, []any, error) {
	q := p.SqrD.Update("roles e").Where(squirrel.Eq{"e.id": id}).Where("e.deleted_at IS NULL")
	q = q.Where("EXISTS (SELECT 1 FROM spaces s WHERE s.id=e.space_id AND s.deleted_at IS NULL)")
	q = q.Where(squirrel.Eq{"e.space_id": spaceId})
	if args.Name != nil {
		q = q.Set("name", *args.Name)
	}

	if args.Description != nil {
		q = q.Set("description", *args.Description)
	}

	if args.IsDefault != nil {
		q = q.Set("is_default", *args.IsDefault)
	}

	if args.Preferences != nil {
		q = q.Set("preferences", *args.Preferences)
	}

	return q.Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).Set("updated_by", args.By).ToSql()
}

func (p *postgresImpl) queryDeleteById(spaceId, id int64, by *int64) (string, []any, error) {
	q := p.SqrD.Update("roles e").Where(squirrel.Eq{"e.id": id}).Where("e.deleted_at IS NULL")
	q = q.Where("EXISTS (SELECT 1 FROM spaces s WHERE s.id=e.space_id AND s.deleted_at IS NULL)")
	q = q.Where(squirrel.Eq{"e.space_id": spaceId})
	q = q.Set("is_default", false)
	return q.Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).Set("deleted_by", by).ToSql()
}

func (p *postgresImpl) queryFilter(spaceId int64, filter domaincontractsrepository.RoleFilter) (string, []any, string, []any, error) {
	limit, offset, err := infrastructurerepositoryshared.Pagination(filter.Page, filter.Limit)
	if err != nil {
		return "", nil, "", nil, err
	}
	q := p.querySelect(spaceId)
	q = infrastructurerepositoryshared.ApplySearch(q, filter.Search, "e.slug", "e.name", "e.description")
	if filter.IsDefault != nil {
		q = q.Where(squirrel.Eq{"e.is_default": *filter.IsDefault})
	}

	count, countArgs, err := q.RemoveColumns().Column("COUNT(*)").ToSql()
	if err != nil {
		return "", nil, "", nil, err
	}

	query, args, err := q.OrderBy("e.created_at DESC", "e.id ASC").Limit(limit).Offset(offset).ToSql()
	return count, countArgs, query, args, err
}

func (p *postgresImpl) queryReadDefault(spaceId int64) (string, []any, error) {
	return p.querySelect(spaceId).Where(squirrel.Eq{"e.is_default": true}).ToSql()
}

func (p *postgresImpl) queryUnsetDefault(spaceId, id int64, by *int64) (string, []any, error) {
	return p.SqrD.Update("roles").
		Set("is_default", false).
		Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("updated_by", by).
		Where(squirrel.Eq{"space_id": spaceId, "is_default": true}).
		Where(squirrel.NotEq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
}
