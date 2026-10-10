package domaincontractsutility

type Validator interface {
	SpaceSlug(val string) error
	SpaceName(val string) error
	SpaceDesc(val string) error
	PermissionSlug(val string) error
	PermissionName(val string) error
	PermissionDesc(val string) error
	RoleSlug(val string) error
	RoleName(val string) error
	RoleDesc(val string) error
	UserName(val string) error
	UserBio(val string) error
	UserUsername(val string) error
	UserEmail(val string) error
	UserPhone(val string) error
	UserPassword(val string) error
	ClientSlug(val string) error
	ClientName(val string) error
	ClientDesc(val string) error
}
