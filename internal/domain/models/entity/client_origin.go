package domainmodelsentity

type ClientOrigin struct {
	Id       int64  `db:"id" json:"id"`
	ClientId int64  `db:"client_id" json:"client_id"`
	Origin   string `db:"origin" json:"origin"`
	IsActive bool   `db:"is_active" json:"is_active"`

	AuditCreateUpdateDelete
}
