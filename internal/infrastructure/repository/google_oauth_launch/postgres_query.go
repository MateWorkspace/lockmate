package infrastructurerepositorygoogleoauthlaunch

import (
	"github.com/Masterminds/squirrel"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

var columns = []string{
	"ticket_hash",
	"client_id",
	"space_id",
	"redirect_uri",
	"client_state",
	"expires_at",
}

func (p *postgresImpl) queryCreate(value domainmodelsentity.GoogleOAuthLaunch) (string, []any, error) {
	return p.sqr.Insert("google_oauth_launches").Columns(columns...).Values(
		value.TicketHash,
		value.ClientId,
		value.SpaceId,
		value.RedirectUri,
		value.ClientState,
		value.ExpiresAt,
	).ToSql()
}

func (p *postgresImpl) queryConsume(ticketHash string) (string, []any, error) {
	q := p.sqr.Delete("google_oauth_launches").Where(squirrel.Eq{"ticket_hash": ticketHash})
	q = q.Where("expires_at > CURRENT_TIMESTAMP")
	return q.Suffix("RETURNING ticket_hash, client_id, space_id, redirect_uri, client_state, expires_at").ToSql()
}

func (p *postgresImpl) queryPurge(limit int64) (string, []any, error) {
	expired := p.sqr.Select("ticket_hash").From("google_oauth_launches").Where("expires_at <= CURRENT_TIMESTAMP").OrderBy("expires_at").Limit(uint64(limit))
	return p.sqr.Delete("google_oauth_launches").Where(squirrel.Expr("ticket_hash IN (?)", expired)).ToSql()
}
