package domainmodelsentity

import "encoding/json"

type Role struct {
	Id          int64           `db:"id" json:"id"`
	SpaceId     int64           `db:"space_id" json:"space_id"`
	Slug        string          `db:"slug" json:"slug"`
	Name        string          `db:"name" json:"name"`
	Description string          `db:"description" json:"description"`
	IsDefault   bool            `db:"is_default" json:"is_default"`
	Preferences json.RawMessage `db:"preferences" json:"preferences"`

	AuditCreateUpdateDelete
}
