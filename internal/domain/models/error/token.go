package domainmodelserror

import "errors"

var (
	TokenFailed  = errors.New("TOKEN_FAILED")
	TokenExpired = errors.New("TOKEN_EXPIRED")
	TokenInvalid = errors.New("TOKEN_INVALID")
)
