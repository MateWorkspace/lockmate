package infrastructurerepositoryuser

import (
	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	infrastructurerepositoryshared "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/shared"
)

var userColumns = []string{
	"id",
	"name",
	"bio",
	"avatar_path",
	"preferences",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

func (p *postgresImpl) querySelect() squirrel.SelectBuilder {
	q := p.SqrD.Select(infrastructurerepositoryshared.QualifiedColumns("e", userColumns)...).From("users e").Where("e.deleted_at IS NULL")
	return q
}

func (p *postgresImpl) queryReadById(id int64) (string, []any, error) {
	return p.querySelect().Where(squirrel.Eq{"e.id": id}).ToSql()
}

func (p *postgresImpl) queryCreate(args domaincontractsrepository.CreateUserArgs) (string, []any, error) {
	columns := []string{"created_by"}
	values := []any{args.By}
	columns = append(columns, "name")
	values = append(values, args.Name)
	if args.Bio != nil {
		columns = append(columns, "bio")
		values = append(values, *args.Bio)
	}

	if args.AvatarPath != nil {
		columns = append(columns, "avatar_path")
		values = append(values, *args.AvatarPath)
	}

	if args.Preferences != nil {
		columns = append(columns, "preferences")
		values = append(values, *args.Preferences)
	}

	return p.SqrD.Insert("users").Columns(columns...).Values(values...).Suffix("RETURNING id").ToSql()
}

func (p *postgresImpl) queryUpdateById(id int64, args domaincontractsrepository.UpdateUserArgs) (string, []any, error) {
	q := p.SqrD.Update("users e").Where(squirrel.Eq{"e.id": id}).Where("e.deleted_at IS NULL")
	if args.Name != nil {
		q = q.Set("name", *args.Name)
	}

	if args.Bio != nil {
		q = q.Set("bio", *args.Bio)
	}

	if args.AvatarPathSetNull {
		q = q.Set("avatar_path", nil)
	} else if args.AvatarPath != nil {
		q = q.Set("avatar_path", *args.AvatarPath)
	}

	if args.Preferences != nil {
		q = q.Set("preferences", *args.Preferences)
	}

	return q.Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).Set("updated_by", args.By).ToSql()
}

func (p *postgresImpl) queryDeleteById(id int64, by *int64) (string, []any, error) {
	q := p.SqrD.Update("users e").Where(squirrel.Eq{"e.id": id}).Where("e.deleted_at IS NULL")
	return q.Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).Set("deleted_by", by).ToSql()
}

func (p *postgresImpl) queryFilter(filter domaincontractsrepository.UserFilter) (string, []any, string, []any, error) {
	limit, offset, err := infrastructurerepositoryshared.Pagination(filter.Page, filter.Limit)
	if err != nil {
		return "", nil, "", nil, err
	}
	q := p.querySelect()
	q = infrastructurerepositoryshared.ApplySearch(q, filter.Search, "e.name", "e.bio")
	count, countArgs, err := q.RemoveColumns().Column("COUNT(*)").ToSql()
	if err != nil {
		return "", nil, "", nil, err
	}

	query, args, err := q.OrderBy("e.created_at DESC", "e.id ASC").Limit(limit).Offset(offset).ToSql()
	return count, countArgs, query, args, err
}

func (p *postgresImpl) queryLockUser(id int64) (string, []any, error) {
	return p.querySelect().RemoveColumns().Column("e.id").Where(squirrel.Eq{"e.id": id}).Suffix("FOR UPDATE").ToSql()
}

func (p *postgresImpl) queryDeleteDependent(table string, userId int64, by *int64) (string, []any, error) {
	return p.SqrD.Update(table).
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", by).
		Where(squirrel.Eq{"user_id": userId}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryDeleteAssignments(userId int64) (string, []any, error) {
	return p.SqrD.Delete("space_member_role").Where(squirrel.Expr("space_member_id IN (SELECT id FROM space_members WHERE user_id=?)", userId)).ToSql()
}
