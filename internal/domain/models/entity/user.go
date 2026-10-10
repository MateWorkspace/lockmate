package domainmodelsentity

import "encoding/json"

type User struct {
	Id          int64           `db:"id" json:"id"`
	Name        string          `db:"name" json:"name"`
	Bio         string          `db:"bio" json:"bio"`
	AvatarPath  *string         `db:"avatar_path" json:"avatar_path,omitempty"`
	Preferences json.RawMessage `db:"preferences" json:"preferences"`

	AuditCreateUpdateDelete
}
