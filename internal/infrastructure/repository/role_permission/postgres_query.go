package infrastructurerepositoryrolepermission

import (
	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	infrastructurerepositoryshared "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/shared"
)

var assignmentColumns = []string{
	"id",
	"space_id",
	"role_id",
	"permission_id",
	"created_at",
	"created_by",
}

var leftColumns = []string{
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

var rightColumns = []string{
	"id",
	"space_id",
	"slug",
	"name",
	"description",
	"preferences",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

func (p *postgresImpl) querySelect(spaceId int64, projection string) squirrel.SelectBuilder {
	columns := infrastructurerepositoryshared.QualifiedColumns("e", assignmentColumns)
	if projection == "details" || projection == "left" {
		columns = append(columns, infrastructurerepositoryshared.QualifiedColumns("l", leftColumns)...)
	}

	if projection == "details" || projection == "right" {
		columns = append(columns, infrastructurerepositoryshared.QualifiedColumns("r", rightColumns)...)
	}
	q := p.SqrD.Select(columns...).
		From("role_permission e").
		Join("roles l ON l.id=e.role_id AND l.space_id=e.space_id AND l.deleted_at IS NULL").
		Join("permissions r ON r.id=e.permission_id AND r.space_id=e.space_id AND r.deleted_at IS NULL").
		Join("spaces s ON s.id=e.space_id AND s.deleted_at IS NULL").
		Where(squirrel.Eq{"e.space_id": spaceId})
	return q
}

func (p *postgresImpl) queryCreate(spaceId int64, args domaincontractsrepository.CreateRolePermissionArgs) (string, []any, error) {
	return p.SqrD.Insert("role_permission").
		Columns("space_id", "role_id", "permission_id", "created_by").
		Values(spaceId, args.RoleId, args.PermissionId, args.By).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryReadById(spaceId, id int64) (string, []any, error) {
	return p.querySelect(spaceId, "details").Where(squirrel.Eq{"e.id": id}).ToSql()
}

func (p *postgresImpl) queryReadPair(spaceId, leftId, rightId int64) (string, []any, error) {
	return p.querySelect(spaceId, "details").Where(squirrel.Eq{"e.role_id": leftId, "e.permission_id": rightId}).ToSql()
}

func (p *postgresImpl) queryReadLeft(spaceId, id int64) (string, []any, error) {
	return p.querySelect(spaceId, "right").Where(squirrel.Eq{"e.role_id": id}).OrderBy("r.created_at DESC", "r.id ASC", "e.id ASC").ToSql()
}

func (p *postgresImpl) queryReadRight(spaceId, id int64) (string, []any, error) {
	return p.querySelect(spaceId, "left").Where(squirrel.Eq{"e.permission_id": id}).OrderBy("l.created_at DESC", "l.id ASC", "e.id ASC").ToSql()
}

func (p *postgresImpl) queryDeleteById(spaceId, id int64) (string, []any, error) {
	return p.SqrD.Delete("role_permission").Where(squirrel.Eq{"space_id": spaceId, "id": id}).ToSql()
}

func (p *postgresImpl) queryDeleteByFilter(spaceId int64, args domaincontractsrepository.DeleteRolePermissionArgs) (string, []any, error) {
	selectors := squirrel.Or{}
	if args.RoleId != nil {
		selectors = append(selectors, squirrel.Eq{"role_id": *args.RoleId})
	}

	if args.PermissionId != nil {
		selectors = append(selectors, squirrel.Eq{"permission_id": *args.PermissionId})
	}

	return p.SqrD.Delete("role_permission").Where(squirrel.Eq{"space_id": spaceId}).Where(selectors).ToSql()
}
