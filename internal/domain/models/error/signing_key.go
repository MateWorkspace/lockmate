package domainmodelserror

import "errors"

var (
	SigningKeyNotFound = errors.New("SIGNING_KEY_NOT_FOUND")
	SigningKeyConflict = errors.New("SIGNING_KEY_CONFLICT")
)
