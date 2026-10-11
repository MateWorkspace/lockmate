package presentationhttprequest

type AuthSignInRequest struct {
	Username string `json:"username" example:"alex"`
	Password string `json:"password" example:"SecurePass123!"`
}

type AuthRegisterRequest struct {
	Name     string  `json:"name" example:"Alex Doe"`
	Username string  `json:"username" example:"alex"`
	Password string  `json:"password" example:"SecurePass123!"`
	Bio      *string `json:"bio" example:"Example profile"`
	Email    *string `json:"email" example:"alex@example.com"`
	Phone    *string `json:"phone" example:"+628123456789"`
}

type AuthRefreshRequest struct {
	RefreshToken string `json:"refresh_token" example:"eyJ..."`
}
