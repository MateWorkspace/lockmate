package presentationhttprequest

type ProfileUpdateClassicAuthRequest struct {
	Username     *string `json:"username" example:"alex"`
	Email        *string `json:"email" example:"alex@example.com"`
	EmailSetNull bool    `json:"email_set_null" example:"true"`
	Phone        *string `json:"phone" example:"+628123456789"`
	PhoneSetNull bool    `json:"phone_set_null" example:"true"`
}

type ProfileChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" example:"SecurePass123!"`
	NewPassword     string `json:"new_password" example:"ReplacementPass123!"`
}
