package domainmodelsentity

import "time"

// AuthHandoff is an internal expiring record; never serialize it as an API response.
type AuthHandoff struct {
	CodeHash      string    `db:"code_hash" json:"-"`
	ClientId      int64     `db:"client_id" json:"-"`
	SpaceId       int64     `db:"space_id" json:"-"`
	UserId        int64     `db:"user_id" json:"-"`
	SpaceMemberId int64     `db:"space_member_id" json:"-"`
	RedirectUri   string    `db:"redirect_uri" json:"-"`
	IsNewUser     bool      `db:"is_new_user" json:"-"`
	ExpiresAt     time.Time `db:"expires_at" json:"-"`
}
