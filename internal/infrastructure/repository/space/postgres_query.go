package infrastructurerepositoryspace

import (
	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	infrastructurerepositoryshared "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/shared"
)

var spaceColumns = []string{
	"id",
	"slug",
	"name",
	"description",
	"is_active",
	"preferences",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

func (p *postgresImpl) querySelect() squirrel.SelectBuilder {
	q := p.SqrD.Select(infrastructurerepositoryshared.QualifiedColumns("e", spaceColumns)...).From("spaces e").Where("e.deleted_at IS NULL")
	return q
}

func (p *postgresImpl) queryReadById(id int64) (string, []any, error) {
	return p.querySelect().Where(squirrel.Eq{"e.id": id}).ToSql()
}

func (p *postgresImpl) queryReadBySlug(value string) (string, []any, error) {
	return p.querySelect().Where(squirrel.Eq{"e.slug": value}).ToSql()
}

func (p *postgresImpl) queryCreate(args domaincontractsrepository.CreateSpaceArgs) (string, []any, error) {
	columns := []string{"created_by"}
	values := []any{args.By}
	columns = append(columns, "slug")
	values = append(values, args.Slug)
	columns = append(columns, "name")
	values = append(values, args.Name)
	if args.Description != nil {
		columns = append(columns, "description")
		values = append(values, *args.Description)
	}

	if args.IsActive != nil {
		columns = append(columns, "is_active")
		values = append(values, *args.IsActive)
	}

	if args.Preferences != nil {
		columns = append(columns, "preferences")
		values = append(values, *args.Preferences)
	}

	return p.SqrD.Insert("spaces").Columns(columns...).Values(values...).Suffix("RETURNING id").ToSql()
}

func (p *postgresImpl) queryUpdateById(id int64, args domaincontractsrepository.UpdateSpaceArgs) (string, []any, error) {
	q := p.SqrD.Update("spaces e").Where(squirrel.Eq{"e.id": id}).Where("e.deleted_at IS NULL")
	if args.Name != nil {
		q = q.Set("name", *args.Name)
	}

	if args.Description != nil {
		q = q.Set("description", *args.Description)
	}

	if args.IsActive != nil {
		q = q.Set("is_active", *args.IsActive)
	}

	if args.Preferences != nil {
		q = q.Set("preferences", *args.Preferences)
	}

	return q.Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).Set("updated_by", args.By).ToSql()
}

func (p *postgresImpl) queryDeleteById(id int64, by *int64) (string, []any, error) {
	q := p.SqrD.Update("spaces e").Where(squirrel.Eq{"e.id": id}).Where("e.deleted_at IS NULL")
	return q.Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).Set("deleted_by", by).ToSql()
}

func (p *postgresImpl) queryFilter(filter domaincontractsrepository.SpaceFilter) (string, []any, string, []any, error) {
	limit, offset, err := infrastructurerepositoryshared.Pagination(filter.Page, filter.Limit)
	if err != nil {
		return "", nil, "", nil, err
	}
	q := p.querySelect()
	q = infrastructurerepositoryshared.ApplySearch(q, filter.Search, "e.slug", "e.name", "e.description")
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
