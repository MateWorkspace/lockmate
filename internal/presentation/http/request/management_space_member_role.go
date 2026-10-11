package presentationhttprequest

type ManagementCreateSpaceMemberRoleRequest struct {
	SpaceMemberId int64 `json:"space_member_id" example:"1"`
	RoleId        int64 `json:"role_id" example:"1"`
}
