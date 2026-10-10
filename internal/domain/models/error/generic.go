package domainmodelserror

import (
	"errors"
)

var (
	NotFound      = errors.New("NOT_FOUND")
	BadArgs       = errors.New("BAD_ARGS")
	Conflict      = errors.New("CONFLICT")
	BadState      = errors.New("BAD_STATE")
	Forbidden     = errors.New("FORBIDDEN")
	Unauthorized  = errors.New("UNAUTHORIZED")
	Timeout       = errors.New("TIMEOUT")
	Unimplemented = errors.New("UNIMPLEMENTED")
	Failure       = errors.New("FAILURE")
	Unknown       = errors.New("UNKNOWN")
)
