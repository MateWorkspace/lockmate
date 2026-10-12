package infrastructureutilityvalidator

import (
	"fmt"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func validateLength(
	field, val string,
	min, max int,
	tooShort,
	tooLong domainmodelserror.Type,
) error {
	length := utf8.RuneCountInString(val)
	if length < min {
		return domainmodelserror.NewError(fmt.Sprintf("%s must contain at least %d characters", field, min), tooShort, nil)
	}
	if length > max {
		return domainmodelserror.NewError(fmt.Sprintf("%s must contain at most %d characters", field, max), tooLong, nil)
	}
	return nil
}

func validatePattern(
	field, val string,
	min, max int,
	pattern *regexp.Regexp,
	invalid,
	tooShort,
	tooLong domainmodelserror.Type,
) error {
	if !utf8.ValidString(val) {
		return domainmodelserror.NewError(field+" must be valid UTF-8", invalid, nil)
	}
	if err := validateLength(field, val, min, max, tooShort, tooLong); err != nil {
		return err
	}
	if strings.TrimSpace(val) != val || !pattern.MatchString(val) {
		return domainmodelserror.NewError(field+" has an invalid format", invalid, nil)
	}
	return nil
}

func validateText(field, val string, invalid, tooLong domainmodelserror.Type) error {
	if !utf8.ValidString(val) {
		return domainmodelserror.NewError(field+" must be valid UTF-8", invalid, nil)
	}
	if err := validateLength(field, val, 0, textMaxLength, nil, tooLong); err != nil {
		return err
	}
	for _, c := range val {
		if unicode.IsControl(c) && c != '\n' && c != '\r' && c != '\t' {
			return domainmodelserror.NewError(field+" contains an invalid control character", invalid, nil)
		}
	}
	return nil
}

func validClientOrigin(val string) bool {
	return validOrigin(val, false)
}

func validOrigin(val string, allowHTTP bool) bool {
	for _, char := range val {
		if char <= ' ' || char > '~' {
			return false
		}
	}

	parsed, err := url.Parse(val)
	if err != nil || parsed.Opaque != "" || parsed.User != nil || parsed.Host == "" ||
		parsed.Path != "" || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return false
	}

	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return false
	}

	host := parsed.Hostname()
	local := host == "localhost"
	canonicalHost := host
	if address, err := netip.ParseAddr(host); err == nil {
		if address.Zone() != "" {
			return false
		}

		local = address.Unmap().IsLoopback()
		canonicalHost = address.String()
		if address.Is4In6() {
			// Browser origin serialization uses hexadecimal IPv6 groups even
			// for IPv4-mapped addresses; netip.String uses a dotted IPv4 tail.
			bytes := address.As16()
			canonicalHost = "::ffff:" + strconv.FormatUint(uint64(bytes[12])<<8|uint64(bytes[13]), 16) +
				":" + strconv.FormatUint(uint64(bytes[14])<<8|uint64(bytes[15]), 16)
		}

		if address.Is6() {
			canonicalHost = "[" + canonicalHost + "]"
		}
	} else if !validOriginHostname(host) {
		return false
	}

	if parsed.Scheme == "http" && !local && !allowHTTP {
		return false
	}

	canonical := parsed.Scheme + "://" + canonicalHost
	if port := parsed.Port(); port != "" {
		value, err := strconv.ParseUint(port, 10, 16)
		if err != nil || value == 0 || parsed.Scheme == "https" && value == 443 || parsed.Scheme == "http" && value == 80 {
			return false
		}

		canonical += ":" + strconv.FormatUint(value, 10)
	}

	if val != canonical {
		return false
	}

	return true
}

func validOriginHostname(host string) bool {
	if host == "" || len(host) > 253 || strings.ToLower(host) != host {
		return false
	}

	labels := strings.Split(host, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}

		for _, char := range label {
			if !(char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || char == '-') {
				return false
			}
		}
	}

	// Browsers interpret numeric final labels as IPv4 addresses, including
	// abbreviated and hexadecimal forms. Only canonical IPs are accepted above.
	last := labels[len(labels)-1]
	if strings.Trim(last, "0123456789") == "" {
		return false
	}

	if strings.HasPrefix(last, "0x") {
		if strings.Trim(last[2:], "0123456789abcdef") == "" {
			return false
		}
	}

	return true
}

func validAuthRedirectUri(val string, allowHTTP bool) bool {
	for _, char := range val {
		if char <= ' ' || char > '~' {
			return false
		}
	}
	parsed, err := url.Parse(val)
	if err != nil || parsed.Opaque != "" || parsed.User != nil || parsed.Path == "" || !strings.HasPrefix(parsed.Path, "/") || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" || strings.Contains(val, "#") {
		return false
	}
	if parsed.Scheme == "http" && !allowHTTP {
		return false
	}
	if !validOrigin(parsed.Scheme+"://"+parsed.Host, allowHTTP) {
		return false
	}
	for _, char := range parsed.Path {
		if char < ' ' || char == 127 {
			return false
		}
	}
	if parsed.String() != val || strings.Contains(parsed.Path, "\\") {
		return false
	}
	for _, part := range strings.Split(parsed.Path, "/") {
		if part == "." || part == ".." {
			return false
		}
	}
	return true
}
