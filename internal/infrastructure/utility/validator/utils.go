package infrastructureutilityvalidator

import (
	"fmt"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"regexp"
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
