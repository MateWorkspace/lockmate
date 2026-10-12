package utility_test

import (
	"errors"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructureutilityvalidator "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/validator"
	"strings"
	"testing"
)

func TestValidator(t *testing.T) {
	v := infrastructureutilityvalidator.NewValidatorImpl()
	for _, test := range []struct {
		name           string
		validate       func(string) error
		valid, invalid string
		expected       error
	}{
		{"space slug", v.SpaceSlug, "taskmate", "Taskmate", domainmodelserror.SpaceSlugInvalid},
		{"space name", v.SpaceName, "Task Mate", " Task Mate", domainmodelserror.SpaceNameInvalid},
		{"space description", v.SpaceDesc, "line one\nline two", "bad\x00", domainmodelserror.SpaceDescInvalid},
		{"permission slug", v.PermissionSlug, "user:get", "User:Get", domainmodelserror.PermissionSlugInvalid},
		{"permission name", v.PermissionName, "Read Users", " Read Users", domainmodelserror.PermissionNameInvalid},
		{"permission description", v.PermissionDesc, "", "\xff", domainmodelserror.PermissionDescInvalid},
		{"role slug", v.RoleSlug, "space-admin", "space_admin", domainmodelserror.RoleSlugInvalid},
		{"role name", v.RoleName, "Administrator", " Administrator", domainmodelserror.RoleNameInvalid},
		{"role description", v.RoleDesc, "", "\x00", domainmodelserror.RoleDescInvalid},
		{"user name", v.UserName, "José User", " José User", domainmodelserror.UserNameInvalid},
		{"user bio", v.UserBio, "Hello\nWorld", "\xff", domainmodelserror.UserBioInvalid},
		{"username", v.UserUsername, "jose.user", "jose user", domainmodelserror.UserUsernameInvalid},
		{"email", v.UserEmail, "user@example.com", "User <user@example.com>", domainmodelserror.UserEmailInvalid},
		{"phone", v.UserPhone, "+1234567890", "12345abc", domainmodelserror.UserPhoneInvalid},
		{"password", v.UserPassword, "password123", "password\n", domainmodelserror.UserPasswordInvalid},
		{"client slug", v.ClientSlug, "taskmate-web", "taskmate--web", domainmodelserror.ClientSlugInvalid},
		{"client name", v.ClientName, "Taskmate Web", "Taskmate@Web", domainmodelserror.ClientNameInvalid},
		{"client description", v.ClientDesc, "", "\x00", domainmodelserror.ClientDescInvalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := test.validate(test.valid); err != nil {
				t.Fatalf("valid input rejected: %v", err)
			}
			if err := test.validate(test.invalid); !errors.Is(err, test.expected) {
				t.Fatalf("expected %v got %v", test.expected, err)
			}
		})
	}
	for _, test := range []struct {
		validate func(string) error
		value    string
		expected error
	}{
		{v.SpaceSlug, "ab", domainmodelserror.SpaceSlugTooShort},
		{v.ClientName, strings.Repeat("a", 101), domainmodelserror.ClientNameTooLong},
		{v.UserPassword, "short", domainmodelserror.UserPasswordTooShort},
		{v.UserPassword, strings.Repeat("é", 37), domainmodelserror.UserPasswordTooLong},
		{v.UserPhone, "1234567", domainmodelserror.UserPhoneTooShort},
		{v.UserPhone, strings.Repeat("1", 16), domainmodelserror.UserPhoneTooLong},
		{v.UserBio, strings.Repeat("a", 1001), domainmodelserror.UserBioTooLong},
	} {
		if err := test.validate(test.value); !errors.Is(err, test.expected) {
			t.Fatalf("boundary: expected %v got %v", test.expected, err)
		}
	}
	if err := v.UserPassword(strings.Repeat("a", 72)); err != nil {
		t.Fatal(err)
	}
	if err := v.UserEmail("user..name@example.com"); !errors.Is(err, domainmodelserror.UserEmailInvalid) {
		t.Fatal("malformed local part accepted")
	}
}

func TestClientAuthRedirectUri(t *testing.T) {
	v := infrastructureutilityvalidator.NewValidatorImpl()
	for _, value := range []string{"https://app.example.com/auth/callback", "http://localhost:3000/callback", "http://127.0.0.1:3000/callback", "https://app.example.com/"} {
		if err := v.ClientAuthRedirectUri(value); err != nil {
			t.Fatalf("valid URI %q: %v", value, err)
		}
	}
	for _, value := range []string{"https://app.example.com", "https://app.example.com/callback?next=elsewhere", "https://app.example.com/callback#", "https://user:pass@app.example.com/callback", "http://app.example.com/callback", "https://APP.example.com/callback", "https://app.example.com/a/../callback", "https://app.example.com/%0a", "https://app.example.com/callback\\evil", "https://app.example.com:443/callback"} {
		if err := v.ClientAuthRedirectUri(value); !errors.Is(err, domainmodelserror.ClientAuthRedirectUriInvalid) {
			t.Fatalf("invalid URI %q: %v", value, err)
		}
	}
	if err := v.ClientAuthRedirectUri("https://app.example.com/" + strings.Repeat("a", 2048)); !errors.Is(err, domainmodelserror.ClientAuthRedirectUriTooLong) {
		t.Fatal(err)
	}
}
