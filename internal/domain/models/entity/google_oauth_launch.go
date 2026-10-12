package domainmodelsentity

import "time"

// GoogleOAuthLaunch is an internal expiring record; never serialize it as an API response.
type GoogleOAuthLaunch struct {
	TicketHash  string    `db:"ticket_hash" json:"-"`
	ClientId    int64     `db:"client_id" json:"-"`
	SpaceId     int64     `db:"space_id" json:"-"`
	RedirectUri string    `db:"redirect_uri" json:"-"`
	ClientState string    `db:"client_state" json:"-"`
	ExpiresAt   time.Time `db:"expires_at" json:"-"`
}
