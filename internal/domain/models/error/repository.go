package domainmodelserror

import "errors"

var (
	SpaceNotFound     = errors.New("SPACE_NOT_FOUND")
	SpaceSlugConflict = errors.New("SPACE_SLUG_CONFLICT")

	PermissionNotFound     = errors.New("PERMISSION_NOT_FOUND")
	PermissionSlugConflict = errors.New("PERMISSION_SLUG_CONFLICT")

	RoleNotFound        = errors.New("ROLE_NOT_FOUND")
	RoleSlugConflict    = errors.New("ROLE_SLUG_CONFLICT")
	RoleDefaultConflict = errors.New("ROLE_DEFAULT_CONFLICT")
	DefaultRoleNotFound = errors.New("DEFAULT_ROLE_NOT_FOUND")

	RolePermissionNotFound = errors.New("ROLE_PERMISSION_NOT_FOUND")
	RolePermissionConflict = errors.New("ROLE_PERMISSION_CONFLICT")

	UserNotFound = errors.New("USER_NOT_FOUND")

	ClassicAuthNotFound         = errors.New("CLASSIC_AUTH_NOT_FOUND")
	ClassicAuthUserConflict     = errors.New("CLASSIC_AUTH_USER_CONFLICT")
	ClassicAuthUsernameConflict = errors.New("CLASSIC_AUTH_USERNAME_CONFLICT")
	ClassicAuthEmailConflict    = errors.New("CLASSIC_AUTH_EMAIL_CONFLICT")
	ClassicAuthPhoneConflict    = errors.New("CLASSIC_AUTH_PHONE_CONFLICT")

	GoogleAuthNotFound        = errors.New("GOOGLE_AUTH_NOT_FOUND")
	GoogleAuthUserConflict    = errors.New("GOOGLE_AUTH_USER_CONFLICT")
	GoogleAuthSubjectConflict = errors.New("GOOGLE_AUTH_SUBJECT_CONFLICT")

	ClientNotFound     = errors.New("CLIENT_NOT_FOUND")
	ClientSlugConflict = errors.New("CLIENT_SLUG_CONFLICT")

	ClientOriginNotFound = errors.New("CLIENT_ORIGIN_NOT_FOUND")
	ClientOriginConflict = errors.New("CLIENT_ORIGIN_CONFLICT")

	SpaceMemberNotFound     = errors.New("SPACE_MEMBER_NOT_FOUND")
	SpaceMemberConflict     = errors.New("SPACE_MEMBER_CONFLICT")
	SpaceMemberRoleNotFound = errors.New("SPACE_MEMBER_ROLE_NOT_FOUND")
	SpaceMemberRoleConflict = errors.New("SPACE_MEMBER_ROLE_CONFLICT")
)
