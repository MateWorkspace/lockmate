package infrastructureutilityvalidator

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
)

const (
	nameMinLength     = 3
	nameMaxLength     = 100
	textMaxLength     = 1000
	usernameMinLength = 3
	usernameMaxLength = 32
	emailMinLength    = 3
	emailMaxLength    = 254
	phoneMinLength    = 8
	phoneMaxLength    = 15
	passwordMinLength = 8
	passwordMaxLength = 72
	originMaxLength   = 2048
)

type validatorImpl struct {
	permissionSlugRegex *regexp.Regexp
	slugRegex           *regexp.Regexp
	nameRegex           *regexp.Regexp
	usernameRegex       *regexp.Regexp
	emailRegex          *regexp.Regexp
	phoneRegex          *regexp.Regexp
}

func NewValidatorImpl() domaincontractsutility.Validator {
	return &validatorImpl{
		permissionSlugRegex: regexp.MustCompile(`^[a-z][a-z0-9_.:-]*$`),
		slugRegex:           regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`),
		nameRegex:           regexp.MustCompile(`^[\p{L}\p{N}][\p{L}\p{M}\p{N} ._'-]*$`),
		usernameRegex:       regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`),
		emailRegex:          regexp.MustCompile("^[A-Za-z0-9.!#$%&'*+/=?^_`{|}~-]+@[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?(?:\\.[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?)+$"),
		phoneRegex:          regexp.MustCompile(`^\+?[0-9]+$`),
	}
}

func (v *validatorImpl) PermissionName(val string) (err error) {
	return validatePattern(
		"permission name", val,
		nameMinLength, nameMaxLength,
		v.nameRegex,
		domainmodelserror.PermissionNameInvalid,
		domainmodelserror.PermissionNameTooShort,
		domainmodelserror.PermissionNameTooLong,
	)
}

func (v *validatorImpl) PermissionDesc(val string) error {
	return validateText(
		"permission description", val,
		domainmodelserror.PermissionDescInvalid,
		domainmodelserror.PermissionDescTooLong,
	)
}

func (v *validatorImpl) RoleName(val string) (err error) {
	return validatePattern(
		"role name", val,
		nameMinLength, nameMaxLength,
		v.nameRegex,
		domainmodelserror.RoleNameInvalid,
		domainmodelserror.RoleNameTooShort,
		domainmodelserror.RoleNameTooLong,
	)
}

func (v *validatorImpl) RoleDesc(val string) error {
	return validateText(
		"role description", val,
		domainmodelserror.RoleDescInvalid,
		domainmodelserror.RoleDescTooLong,
	)
}

func (v *validatorImpl) UserName(val string) (err error) {
	return validatePattern(
		"user name", val,
		nameMinLength, nameMaxLength,
		v.nameRegex,
		domainmodelserror.UserNameInvalid,
		domainmodelserror.UserNameTooShort,
		domainmodelserror.UserNameTooLong,
	)
}

func (v *validatorImpl) UserBio(val string) error {
	return validateText("user bio", val, domainmodelserror.UserBioInvalid, domainmodelserror.UserBioTooLong)
}

func (v *validatorImpl) UserUsername(val string) (err error) {
	return validatePattern(
		"username", val,
		usernameMinLength, usernameMaxLength,
		v.usernameRegex,
		domainmodelserror.UserUsernameInvalid,
		domainmodelserror.UserUsernameTooShort,
		domainmodelserror.UserUsernameTooLong,
	)
}

func (v *validatorImpl) UserEmail(val string) (err error) {
	if err := validatePattern(
		"email", val,
		emailMinLength, emailMaxLength,
		v.emailRegex,
		domainmodelserror.UserEmailInvalid,
		domainmodelserror.UserEmailTooShort,
		domainmodelserror.UserEmailTooLong,
	); err != nil {
		return err
	}

	address, err := mail.ParseAddress(val)
	if err != nil || address.Address != val {
		return domainmodelserror.NewError("email has an invalid format", domainmodelserror.UserEmailInvalid, nil)
	}
	return nil
}

func (v *validatorImpl) UserPhone(val string) (err error) {
	if err := validateLength(
		"phone", strings.TrimPrefix(val, "+"),
		phoneMinLength, phoneMaxLength,
		domainmodelserror.UserPhoneTooShort,
		domainmodelserror.UserPhoneTooLong,
	); err != nil {
		return err
	}

	if !v.phoneRegex.MatchString(val) {
		return domainmodelserror.NewError("phone must contain digits with an optional leading '+'", domainmodelserror.UserPhoneInvalid, nil)
	}
	return nil
}

func (v *validatorImpl) UserPassword(val string) (err error) {
	if !utf8.ValidString(val) {
		return domainmodelserror.NewError("password must be valid UTF-8", domainmodelserror.UserPasswordInvalid, nil)
	}
	if err := validateLength(
		"password", val,
		passwordMinLength, passwordMaxLength,
		domainmodelserror.UserPasswordTooShort,
		domainmodelserror.UserPasswordTooLong,
	); err != nil {
		return err
	}

	if len(val) > passwordMaxLength {
		return domainmodelserror.NewError("password must not exceed 72 bytes", domainmodelserror.UserPasswordTooLong, nil)
	}

	for _, char := range val {
		if unicode.IsControl(char) {
			return domainmodelserror.NewError("password must not contain control characters", domainmodelserror.UserPasswordInvalid, nil)
		}
	}
	return nil
}

func (v *validatorImpl) SpaceSlug(val string) error {
	return validatePattern(
		"space slug", val,
		3, 100,
		v.slugRegex,
		domainmodelserror.SpaceSlugInvalid,
		domainmodelserror.SpaceSlugTooShort,
		domainmodelserror.SpaceSlugTooLong,
	)
}

func (v *validatorImpl) SpaceName(val string) error {
	return validatePattern(
		"space name", val,
		3, 100,
		v.nameRegex,
		domainmodelserror.SpaceNameInvalid,
		domainmodelserror.SpaceNameTooShort,
		domainmodelserror.SpaceNameTooLong,
	)
}

func (v *validatorImpl) SpaceDesc(val string) error {
	return validateText(
		"space desc", val,
		domainmodelserror.SpaceDescInvalid,
		domainmodelserror.SpaceDescTooLong,
	)
}

func (v *validatorImpl) PermissionSlug(val string) error {
	return validatePattern(
		"permission slug", val,
		3, 100,
		v.permissionSlugRegex,
		domainmodelserror.PermissionSlugInvalid,
		domainmodelserror.PermissionSlugTooShort,
		domainmodelserror.PermissionSlugTooLong,
	)
}

func (v *validatorImpl) RoleSlug(val string) error {
	return validatePattern(
		"role slug", val,
		3, 100,
		v.slugRegex,
		domainmodelserror.RoleSlugInvalid,
		domainmodelserror.RoleSlugTooShort,
		domainmodelserror.RoleSlugTooLong,
	)
}

func (v *validatorImpl) ClientSlug(val string) error {
	return validatePattern(
		"client slug", val,
		3, 100,
		v.slugRegex,
		domainmodelserror.ClientSlugInvalid,
		domainmodelserror.ClientSlugTooShort,
		domainmodelserror.ClientSlugTooLong,
	)
}

func (v *validatorImpl) ClientName(val string) error {
	return validatePattern(
		"client name", val,
		3, 100,
		v.nameRegex,
		domainmodelserror.ClientNameInvalid,
		domainmodelserror.ClientNameTooShort,
		domainmodelserror.ClientNameTooLong,
	)
}

func (v *validatorImpl) ClientDesc(val string) error {
	return validateText(
		"client desc", val,
		domainmodelserror.ClientDescInvalid,
		domainmodelserror.ClientDescTooLong,
	)
}

func (v *validatorImpl) ClientOrigin(val string) error {
	if len(val) > originMaxLength {
		return domainmodelserror.NewError("client origin must not exceed 2048 bytes", domainmodelserror.ClientOriginTooLong, nil)
	}

	if !validClientOrigin(val) {
		return domainmodelserror.NewError("client origin must be a canonical HTTPS or local HTTP origin", domainmodelserror.ClientOriginInvalid, nil)
	}

	return nil
}

func (v *validatorImpl) ClientAuthRedirectUri(val string) error {
	if len(val) > 2048 {
		return domainmodelserror.NewError("authentication redirect URI is too long", domainmodelserror.ClientAuthRedirectUriTooLong, nil)
	}
	if !validAuthRedirectUri(val) {
		return domainmodelserror.NewError("authentication redirect URI has an invalid format", domainmodelserror.ClientAuthRedirectUriInvalid, nil)
	}
	return nil
}
