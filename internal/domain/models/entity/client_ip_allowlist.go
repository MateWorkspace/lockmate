package domainmodelsentity

type ClientIpAllowlist struct {
	Id       int64  `db:"id" json:"id"`
	ClientId int64  `db:"client_id" json:"client_id"`
	Cidr     string `db:"cidr" json:"cidr"`
	IsActive bool   `db:"is_active" json:"is_active"`

	AuditCreateUpdateDelete
}
