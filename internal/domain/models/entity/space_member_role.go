package domainmodelsentity

type SpaceMemberRole struct {
	Id            int64 `db:"id" json:"id"`
	SpaceId       int64 `db:"space_id" json:"space_id"`
	SpaceMemberId int64 `db:"space_member_id" json:"space_member_id"`
	RoleId        int64 `db:"role_id" json:"role_id"`

	AuditCreate
}
