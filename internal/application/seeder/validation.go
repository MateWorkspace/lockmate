package applicationseeder

import domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"

// validate checks every definition before the first database write. Error messages
// identify fields and references without including credential values.
func (u *usecase) validate() error {
	if u.data == nil {
		return domainmodelserror.NewError("seed definitions are missing", domainmodelserror.BadArgs, nil)
	}

	if err := u.validator.SpaceSlug(u.data.Space.Slug); err != nil {
		return err
	}

	if err := u.validator.SpaceName(u.data.Space.Name); err != nil {
		return err
	}

	if err := u.validator.SpaceDesc(u.data.Space.Description); err != nil {
		return err
	}

	permissions := make(map[string]bool, len(u.data.Permissions))
	for _, value := range u.data.Permissions {
		if err := u.validator.PermissionSlug(value.Slug); err != nil {
			return err
		}

		if err := u.validator.PermissionName(value.Name); err != nil {
			return err
		}

		if err := u.validator.PermissionDesc(value.Description); err != nil {
			return err
		}

		if permissions[value.Slug] {
			return domainmodelserror.NewError("duplicate seed permission slug", domainmodelserror.BadArgs, nil)
		}

		permissions[value.Slug] = true
	}

	roles := make(map[string]bool, len(u.data.Roles))
	defaults := 0
	for _, value := range u.data.Roles {
		if err := u.validator.RoleSlug(value.Slug); err != nil {
			return err
		}

		if err := u.validator.RoleName(value.Name); err != nil {
			return err
		}

		if err := u.validator.RoleDesc(value.Description); err != nil {
			return err
		}

		if roles[value.Slug] {
			return domainmodelserror.NewError("duplicate seed role slug", domainmodelserror.BadArgs, nil)
		}

		roles[value.Slug] = true
		if value.IsDefault {
			defaults++
		}

		assigned := make(map[string]bool, len(value.Permissions))
		for _, slug := range value.Permissions {
			if !permissions[slug] {
				return domainmodelserror.NewError("seed role references an undefined permission", domainmodelserror.BadArgs, nil)
			}

			if assigned[slug] {
				return domainmodelserror.NewError("duplicate seed role permission", domainmodelserror.BadArgs, nil)
			}

			assigned[slug] = true
		}
	}

	if defaults != 1 {
		return domainmodelserror.NewError("seed definitions require exactly one default role", domainmodelserror.BadArgs, nil)
	}

	usernames := make(map[string]bool, len(u.data.Users))
	for _, value := range u.data.Users {
		if !roles[value.RoleSlug] {
			return domainmodelserror.NewError("seed user references an undefined role", domainmodelserror.BadArgs, nil)
		}

		if err := u.validator.UserName(value.Name); err != nil {
			return err
		}

		if err := u.validator.UserUsername(value.Username); err != nil {
			return err
		}

		if err := u.validator.UserPassword(value.Password); err != nil {
			return err
		}

		if value.Bio != nil {
			if err := u.validator.UserBio(*value.Bio); err != nil {
				return err
			}
		}

		if usernames[value.Username] {
			return domainmodelserror.NewError("duplicate seed username", domainmodelserror.BadArgs, nil)
		}

		usernames[value.Username] = true
	}

	clients := make(map[string]bool, len(u.data.Clients))
	for _, value := range u.data.Clients {
		if err := u.validator.ClientSlug(value.Slug); err != nil {
			return err
		}

		if err := u.validator.ClientName(value.Name); err != nil {
			return err
		}

		if err := u.validator.ClientDesc(value.Description); err != nil {
			return err
		}

		if err := u.keygen.Validate(value.SecretKey); err != nil {
			return err
		}

		if value.AuthRedirectUri != nil {
			if err := u.validator.ClientAuthRedirectUri(*value.AuthRedirectUri); err != nil {
				return err
			}
		}

		if clients[value.Slug] {
			return domainmodelserror.NewError("duplicate seed client slug", domainmodelserror.BadArgs, nil)
		}

		clients[value.Slug] = true
		origins := make(map[string]bool, len(value.Origins))
		for _, origin := range value.Origins {
			if err := u.validator.ClientOrigin(origin); err != nil {
				return err
			}

			if origins[origin] {
				return domainmodelserror.NewError("duplicate seed client origin", domainmodelserror.BadArgs, nil)
			}

			origins[origin] = true
		}
	}

	return nil
}
