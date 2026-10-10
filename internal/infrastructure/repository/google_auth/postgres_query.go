package infrastructurerepositorygoogleauth

import (
	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	infrastructurerepositoryshared "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/shared"
)

var googleAuthColumns = []string{
	"id",
	"user_id",
	"subject",
	"email",
	"is_email_verified",
	"preferences",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

func (p *postgresImpl) querySelect() squirrel.SelectBuilder {
	q := p.SqrD.Select(infrastructurerepositoryshared.QualifiedColumns("e", googleAuthColumns)...).From("google_auths e").Where("e.deleted_at IS NULL")
	q = q.Where("EXISTS (SELECT 1 FROM users u WHERE u.id=e.user_id AND u.deleted_at IS NULL)")
	return q
}

func (p *postgresImpl) queryReadById(id int64) (string, []any, error) {
	return p.querySelect().Where(squirrel.Eq{"e.id": id}).ToSql()
}

func (p *postgresImpl) queryReadByUserId(value int64) (string, []any, error) {
	return p.querySelect().Where(squirrel.Eq{"e.user_id": value}).ToSql()
}

func (p *postgresImpl) queryReadBySubject(value string) (string, []any, error) {
	return p.querySelect().Where(squirrel.Eq{"e.subject": value}).ToSql()
}

func (p *postgresImpl) queryCreate(args domaincontractsrepository.CreateGoogleAuthArgs) (string, []any, error) {
	columns := []string{"created_by"}
	values := []any{args.By}
	columns = append(columns, "user_id")
	values = append(values, args.UserId)
	columns = append(columns, "subject")
	values = append(values, args.Subject)
	columns = append(columns, "email")
	values = append(values, args.Email)
	if args.IsEmailVerified != nil {
		columns = append(columns, "is_email_verified")
		values = append(values, *args.IsEmailVerified)
	}

	if args.Preferences != nil {
		columns = append(columns, "preferences")
		values = append(values, *args.Preferences)
	}

	return p.SqrD.Insert("google_auths").Columns(columns...).Values(values...).Suffix("RETURNING id").ToSql()
}

func (p *postgresImpl) queryUpdateById(id int64, args domaincontractsrepository.UpdateGoogleAuthArgs) (string, []any, error) {
	q := p.SqrD.Update("google_auths e").Where(squirrel.Eq{"e.id": id}).Where("e.deleted_at IS NULL")
	q = q.Where("EXISTS (SELECT 1 FROM users u WHERE u.id=e.user_id AND u.deleted_at IS NULL)")
	if args.Email != nil {
		q = q.Set("email", *args.Email)
	}

	if args.IsEmailVerified != nil {
		q = q.Set("is_email_verified", *args.IsEmailVerified)
	}

	if args.Preferences != nil {
		q = q.Set("preferences", *args.Preferences)
	}

	if args.IsEmailVerified == nil && (args.Email != nil) {
		var value any
		if args.Email != nil {
			value = *args.Email
		}
		q = q.Set("is_email_verified", squirrel.Expr("CASE WHEN e.email IS DISTINCT FROM ? THEN FALSE ELSE e.is_email_verified END", value))
	}

	return q.Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).Set("updated_by", args.By).ToSql()
}

func (p *postgresImpl) queryDeleteById(id int64, by *int64) (string, []any, error) {
	q := p.SqrD.Update("google_auths e").Where(squirrel.Eq{"e.id": id}).Where("e.deleted_at IS NULL")
	q = q.Where("EXISTS (SELECT 1 FROM users u WHERE u.id=e.user_id AND u.deleted_at IS NULL)")
	return q.Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).Set("deleted_by", by).ToSql()
}

func (p *postgresImpl) queryFilter(filter domaincontractsrepository.GoogleAuthFilter) (string, []any, string, []any, error) {
	limit, offset, err := infrastructurerepositoryshared.Pagination(filter.Page, filter.Limit)
	if err != nil {
		return "", nil, "", nil, err
	}
	q := p.querySelect()
	q = infrastructurerepositoryshared.ApplySearch(q, filter.Search, "e.email")
	if filter.UserId != nil {
		q = q.Where(squirrel.Eq{"e.user_id": *filter.UserId})
	}

	if filter.IsEmailVerified != nil {
		q = q.Where(squirrel.Eq{"e.is_email_verified": *filter.IsEmailVerified})
	}

	count, countArgs, err := q.RemoveColumns().Column("COUNT(*)").ToSql()
	if err != nil {
		return "", nil, "", nil, err
	}

	query, args, err := q.OrderBy("e.created_at DESC", "e.id ASC").Limit(limit).Offset(offset).ToSql()
	return count, countArgs, query, args, err
}
