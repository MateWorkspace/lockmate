package domainmodelsentity

type RolePermission struct {
	Id           int64 `db:"id" json:"id"`
	SpaceId      int64 `db:"space_id" json:"space_id"`
	RoleId       int64 `db:"role_id" json:"role_id"`
	PermissionId int64 `db:"permission_id" json:"permission_id"`

	AuditCreate
}
