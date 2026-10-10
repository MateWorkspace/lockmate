package domainmodelsentity

import "encoding/json"

type GoogleAuth struct {
	Id              int64           `db:"id" json:"id"`
	UserId          int64           `db:"user_id" json:"user_id"`
	Subject         string          `db:"subject" json:"subject"`
	Email           string          `db:"email" json:"email"`
	IsEmailVerified bool            `db:"is_email_verified" json:"is_email_verified"`
	Preferences     json.RawMessage `db:"preferences" json:"preferences"`

	AuditCreateUpdateDelete
}
