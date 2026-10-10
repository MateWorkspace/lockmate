package infrastructurerepositoryspacememberrole

import (
	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	infrastructurerepositoryshared "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/shared"
)

var assignmentColumns = []string{
	"id",
	"space_id",
	"space_member_id",
	"role_id",
	"created_at",
	"created_by",
}

var leftColumns = []string{
	"id",
	"space_id",
	"user_id",
	"is_active",
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
	"is_default",
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
		From("space_member_role e").
		Join("space_members l ON l.id=e.space_member_id AND l.space_id=e.space_id AND l.deleted_at IS NULL").
		Join("roles r ON r.id=e.role_id AND r.space_id=e.space_id AND r.deleted_at IS NULL").
		Join("spaces s ON s.id=e.space_id AND s.deleted_at IS NULL").
		Where(squirrel.Eq{"e.space_id": spaceId})
	q = q.Join("users u ON u.id=l.user_id AND u.deleted_at IS NULL")
	return q
}

func (p *postgresImpl) queryCreate(spaceId int64, args domaincontractsrepository.CreateSpaceMemberRoleArgs) (string, []any, error) {
	return p.SqrD.Insert("space_member_role").
		Columns("space_id", "space_member_id", "role_id", "created_by").
		Values(spaceId, args.SpaceMemberId, args.RoleId, args.By).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryReadById(spaceId, id int64) (string, []any, error) {
	return p.querySelect(spaceId, "details").Where(squirrel.Eq{"e.id": id}).ToSql()
}

func (p *postgresImpl) queryReadPair(spaceId, leftId, rightId int64) (string, []any, error) {
	return p.querySelect(spaceId, "details").Where(squirrel.Eq{"e.space_member_id": leftId, "e.role_id": rightId}).ToSql()
}

func (p *postgresImpl) queryReadLeft(spaceId, id int64) (string, []any, error) {
	return p.querySelect(spaceId, "right").Where(squirrel.Eq{"e.space_member_id": id}).OrderBy("r.created_at DESC", "r.id ASC", "e.id ASC").ToSql()
}

func (p *postgresImpl) queryReadRight(spaceId, id int64) (string, []any, error) {
	return p.querySelect(spaceId, "left").Where(squirrel.Eq{"e.role_id": id}).OrderBy("l.created_at DESC", "l.id ASC", "e.id ASC").ToSql()
}

func (p *postgresImpl) queryDeleteById(spaceId, id int64) (string, []any, error) {
	return p.SqrD.Delete("space_member_role").Where(squirrel.Eq{"space_id": spaceId, "id": id}).ToSql()
}

func (p *postgresImpl) queryDeleteByFilter(spaceId int64, args domaincontractsrepository.DeleteSpaceMemberRoleArgs) (string, []any, error) {
	selectors := squirrel.Or{}
	if args.SpaceMemberId != nil {
		selectors = append(selectors, squirrel.Eq{"space_member_id": *args.SpaceMemberId})
	}

	if args.RoleId != nil {
		selectors = append(selectors, squirrel.Eq{"role_id": *args.RoleId})
	}

	return p.SqrD.Delete("space_member_role").Where(squirrel.Eq{"space_id": spaceId}).Where(selectors).ToSql()
}

func (p *postgresImpl) queryMemberUser(spaceId, id int64) (string, []any, error) {
	return p.SqrD.Select("user_id").From("space_members").Where(squirrel.Eq{"space_id": spaceId, "id": id}).Where("deleted_at IS NULL").ToSql()
}

func (p *postgresImpl) queryEligibility(spaceId, id int64) (string, []any, error) {
	return p.SqrD.Select("s.is_active", "m.is_active").
		From("space_members m").
		Join("spaces s ON s.id=m.space_id AND s.deleted_at IS NULL").
		Join("users u ON u.id=m.user_id AND u.deleted_at IS NULL").
		Where(squirrel.Eq{"m.space_id": spaceId, "m.id": id}).
		Where("m.deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryEffectiveRoles(spaceId, id int64) (string, []any, error) {
	return p.querySelect(spaceId, "right").
		RemoveColumns().
		Columns(infrastructurerepositoryshared.QualifiedColumns("r", rightColumns)...).
		Distinct().
		Where(squirrel.Eq{"e.space_member_id": id, "s.is_active": true, "l.is_active": true}).
		OrderBy("r.slug ASC", "r.id ASC").
		ToSql()
}

var permissionColumns = []string{
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

func (p *postgresImpl) queryEffectivePermissions(spaceId, id int64) (string, []any, error) {
	return p.querySelect(spaceId, "right").
		RemoveColumns().
		Columns(infrastructurerepositoryshared.QualifiedColumns("p", permissionColumns)...).
		Join("role_permission rp ON rp.role_id=r.id AND rp.space_id=e.space_id").
		Join("permissions p ON p.id=rp.permission_id AND p.space_id=rp.space_id AND p.deleted_at IS NULL").
		Distinct().
		Where(squirrel.Eq{"e.space_member_id": id, "s.is_active": true, "l.is_active": true}).
		OrderBy("p.slug ASC", "p.id ASC").
		ToSql()
}
