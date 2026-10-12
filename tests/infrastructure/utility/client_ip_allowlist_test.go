package utility_test

import (
	"errors"
	"strings"
	"testing"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructureutilityvalidator "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/validator"
)

func TestClientIpAllowlistValidation(t *testing.T) {
	validator := infrastructureutilityvalidator.NewValidatorImpl(true)
	for _, value := range []string{
		"203.0.113.9", "203.0.113.0/24", "0.0.0.0/0", "255.255.255.255",
		"127.0.0.1", "10.0.0.0/8", "::1", "2001:db8::1", "2001:db8::/32", "::/0",
		"::ffff:203.0.113.9", "::ffff:203.0.113.0/120",
	} {
		t.Run(value, func(t *testing.T) {
			if err := validator.ClientIpAllowlist(value); err != nil {
				t.Fatalf("valid IP or CIDR rejected: %v", err)
			}
		})
	}

	for _, value := range []string{
		"", "null", "*", "not-an-ip", "app.example.com", "localhost",
		"300.0.0.0/8", "203.0.113.9/24", "203.0.113.0/33", "10.0.0.0/8/8",
		"2001:db8::/129", "2001:db8::1/64", "https://app.example.com", "http://127.0.0.1",
		"1.2.3.4:80", "1.2.3.4/24 ", " 1.2.3.4/24", "1.2.3.4/24\n",
		"fe80::1%eth0", "127.000.000.001", "::ffff:192.0.2.1/120",
	} {
		t.Run(value, func(t *testing.T) {
			if err := validator.ClientIpAllowlist(value); !errors.Is(err, domainmodelserror.ClientIpAllowlistInvalid) {
				t.Fatalf("invalid IP or CIDR accepted or incorrectly classified: %v", err)
			}
		})
	}

	if err := validator.ClientIpAllowlist(strings.Repeat("a", 129)); !errors.Is(err, domainmodelserror.ClientIpAllowlistTooLong) {
		t.Fatalf("byte limit not enforced: %v", err)
	}
}

func TestIPAllowlistIgnoresRedirectHTTPAllowance(t *testing.T) {
	v := infrastructureutilityvalidator.NewValidatorImpl(true)
	for _, value := range []string{"http://100.92.75.2:10703", "https://app.example.com", "localhost"} {
		if err := v.ClientIpAllowlist(value); !errors.Is(err, domainmodelserror.ClientIpAllowlistInvalid) {
			t.Fatalf("hostname or URL accepted as IP allowlist entry: %q", value)
		}
	}
}
