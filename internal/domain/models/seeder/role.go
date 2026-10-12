package domainmodelsseeder

type Role struct {
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	IsDefault   bool     `json:"is_default"`
	Permissions []string `json:"permissions"`
}
