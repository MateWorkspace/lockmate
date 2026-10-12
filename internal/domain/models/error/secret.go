package domainmodelserror

import "errors"

var (
	SecretKeyInvalid    = errors.New("SECRET_KEY_INVALID")
	SecretEncryptFailed = errors.New("SECRET_ENCRYPT_FAILED")
	SecretDecryptFailed = errors.New("SECRET_DECRYPT_FAILED")
)
