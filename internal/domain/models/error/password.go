package domainmodelserror

import "errors"

var (
	PasswordMismatch      = errors.New("PASSWORD_MISMATCH")
	PasswordHashFailed    = errors.New("PASSWORD_HASH_FAILED")
	PasswordCompareFailed = errors.New("PASSWORD_COMPARE_FAILED")
)
