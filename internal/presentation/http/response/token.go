package presentationhttpresponse

import (
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
)

// TokenResponse carries only issued session tokens.
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJ..."`
	RefreshToken string `json:"refresh_token" example:"eyJ..."`
	IsNewUser    bool   `json:"is_new_user" example:"false"`
}

func Token(value domainusecasesauth.TokenResult) TokenResponse {
	return TokenResponse{AccessToken: value.AccessToken, RefreshToken: value.RefreshToken, IsNewUser: value.IsNewUser}
}
