package domainmodelserror

import "errors"

var (
	KeygenFailed = errors.New("KEYGEN_FAILED")
	KeyInvalid   = errors.New("KEY_INVALID")
	KeyMismatch  = errors.New("KEY_MISMATCH")
)
