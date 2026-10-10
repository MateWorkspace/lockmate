package domainmodelsentity

import "encoding/json"

type Space struct {
	Id          int64           `db:"id" json:"id"`
	Slug        string          `db:"slug" json:"slug"`
	Name        string          `db:"name" json:"name"`
	Description string          `db:"description" json:"description"`
	IsActive    bool            `db:"is_active" json:"is_active"`
	Preferences json.RawMessage `db:"preferences" json:"preferences"`

	AuditCreateUpdateDelete
}
