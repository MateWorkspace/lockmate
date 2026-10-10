package domainmodelsentity

import "encoding/json"

type ClassicAuth struct {
	Id              int64           `db:"id" json:"id"`
	UserId          int64           `db:"user_id" json:"user_id"`
	Username        string          `db:"username" json:"username"`
	Email           *string         `db:"email" json:"email,omitempty"`
	Phone           *string         `db:"phone" json:"phone,omitempty"`
	PasswordHash    string          `db:"password_hash" json:"password_hash"`
	IsEmailVerified bool            `db:"is_email_verified" json:"is_email_verified"`
	IsPhoneVerified bool            `db:"is_phone_verified" json:"is_phone_verified"`
	Preferences     json.RawMessage `db:"preferences" json:"preferences"`

	AuditCreateUpdateDelete
}
