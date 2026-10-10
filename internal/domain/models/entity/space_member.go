package domainmodelsentity

import "encoding/json"

type SpaceMember struct {
	Id          int64           `db:"id" json:"id"`
	SpaceId     int64           `db:"space_id" json:"space_id"`
	UserId      int64           `db:"user_id" json:"user_id"`
	IsActive    bool            `db:"is_active" json:"is_active"`
	Preferences json.RawMessage `db:"preferences" json:"preferences"`

	AuditCreateUpdateDelete
}
