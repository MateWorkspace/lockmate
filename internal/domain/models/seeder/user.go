package domainmodelsseeder

type User struct {
	RoleSlug string  `json:"role_slug"`
	Name     string  `json:"name"`
	Bio      *string `json:"bio,omitempty"`
	Username string  `json:"username"`

	// Password is supplied at runtime and is never embedded or serialized.
	Password string `json:"-"`
}
