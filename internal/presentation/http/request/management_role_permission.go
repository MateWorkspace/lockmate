package presentationhttprequest

type ManagementCreateRolePermissionRequest struct {
	RoleId       int64 `json:"role_id" example:"1"`
	PermissionId int64 `json:"permission_id" example:"1"`
}
