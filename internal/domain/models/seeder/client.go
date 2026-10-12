package domainmodelsseeder

type Client struct {
	Slug            string   `json:"slug"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	IsActive        bool     `json:"is_active"`
	AuthRedirectUri *string  `json:"auth_redirect_uri"`
	Origins         []string `json:"origins"`

	// SecretKey is supplied at runtime and is never embedded or serialized.
	SecretKey string `json:"-"`
}
