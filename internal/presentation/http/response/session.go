package presentationhttpresponse

import (
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
)

type SessionResponse struct {
	User         UserResponse         `json:"user"`
	Space        SpaceResponse        `json:"space"`
	SpaceMember  SpaceMemberResponse  `json:"space_member"`
	Roles        []RoleResponse       `json:"roles"`
	Permissions  []PermissionResponse `json:"permissions"`
	AccessToken  string               `json:"access_token" example:"eyJ..."`
	RefreshToken string               `json:"refresh_token" example:"eyJ..."`
	IsNewUser    bool                 `json:"is_new_user" example:"false"`
}

func Session(value domainusecasesauth.SessionResult) SessionResponse {
	return SessionResponse{User: User(value.User), Space: Space(value.Space), SpaceMember: SpaceMember(value.SpaceMember), Roles: RoleList(value.Roles), Permissions: PermissionList(value.Permissions), AccessToken: value.AccessToken, RefreshToken: value.RefreshToken, IsNewUser: value.IsNewUser}
}
