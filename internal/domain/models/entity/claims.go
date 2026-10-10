package domainmodelsentity

type AccessClaims struct {
	UserId        int64    `json:"user_id"`
	SpaceId       int64    `json:"space_id"`
	SpaceMemberId int64    `json:"space_member_id"`
	Name          string   `json:"name"`
	Roles         []string `json:"roles"`
	Permissions   []string `json:"permissions"`
}

type RefreshClaims struct {
	UserId        int64  `json:"user_id"`
	SpaceId       int64  `json:"space_id"`
	SpaceMemberId int64  `json:"space_member_id"`
	Name          string `json:"name"`
}
