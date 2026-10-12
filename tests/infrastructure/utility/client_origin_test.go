package utility_test

import (
	"errors"
	"strings"
	"testing"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructureutilityvalidator "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/validator"
)

func TestClientOriginValidation(t *testing.T) {
	validator := infrastructureutilityvalidator.NewValidatorImpl(true)
	for _, origin := range []string{
		"https://app.example.com", "https://staging.example.com:8443",
		"https://xn--bcher-kva.example", "https://192.0.2.1", "https://[2001:db8::1]:65535",
		"https://[::ffff:c000:201]", "http://[::ffff:7f00:1]",
		"https://localhost", "https://internal", "http://localhost:3000",
		"http://127.0.0.1", "http://127.1.2.3:8080", "http://[::1]:3000",
	} {
		t.Run(origin, func(t *testing.T) {
			if err := validator.ClientOrigin(origin); err != nil {
				t.Fatalf("valid origin rejected: %v", err)
			}
		})
	}

	for _, origin := range []string{
		"", "null", "*", "https://*.example.com", "app.example.com",
		"http://example.com", "http://localhost.example.com", "http://192.168.1.1", "http://[::2]",
		"HTTPS://example.com", "https://EXAMPLE.com", "https://example.com.",
		"https://example.com/", "https://example.com/path", "https://example.com?", "https://example.com?q=1",
		"https://example.com#", "https://example.com#fragment", "https://user:password@example.com",
		" https://example.com", "https://example.com ", "https://example.com\n", "https://example.com\x00",
		"ftp://example.com", "https:example.com", "https://example.com:443", "http://localhost:80",
		"https://example.com:0", "https://example.com:65536", "https://example.com:00444",
		"https://example.com:", "https://example.com:abc", "https://[::1", "https://::1",
		"https://[::ffff:192.0.2.1]", "http://[::ffff:192.0.2.1]",
		"https://[fe80::1%25eth0]", "https://[2001:0db8:0:0:0:0:0:1]", "https://[2001:DB8::1]",
		"https://bücher.example", "https://%65xample.com", "https://example_com", "https://-example.com",
		"https://example-.com", "https://example..com", "https://" + strings.Repeat("a", 64) + ".com",
		"https://999.0.0.1", "https://127.1", "https://2130706433", "https://0177.0.0.1", "https://0x7f000001",
		"https://" + strings.Repeat("9", 63), "https://example.0x", "https://[127.0.0.1]",
	} {
		t.Run(origin, func(t *testing.T) {
			if err := validator.ClientOrigin(origin); !errors.Is(err, domainmodelserror.ClientOriginInvalid) {
				t.Fatalf("invalid origin accepted or incorrectly classified: %v", err)
			}
		})
	}

	if err := validator.ClientOrigin(strings.Repeat("a", 2049)); !errors.Is(err, domainmodelserror.ClientOriginTooLong) {
		t.Fatalf("byte limit not enforced: %v", err)
	}
}
