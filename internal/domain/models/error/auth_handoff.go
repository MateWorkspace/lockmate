package domainmodelserror

import "errors"

var (
	AuthHandoffInvalid = errors.New("AUTH_HANDOFF_INVALID")
	AuthHandoffExpired = errors.New("AUTH_HANDOFF_EXPIRED")
)
