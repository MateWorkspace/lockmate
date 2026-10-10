package domainmodelserror

import "errors"

var (
	GoogleOAuthStateInvalid = errors.New("GOOGLE_OAUTH_STATE_INVALID")
	GoogleOAuthDenied       = errors.New("GOOGLE_OAUTH_DENIED")
	GoogleOAuthInvalid      = errors.New("GOOGLE_OAUTH_INVALID")
	GoogleOAuthFailed       = errors.New("GOOGLE_OAUTH_FAILED")
)
