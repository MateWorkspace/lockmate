package domainmodelsentity

import "time"

type AuditCreate struct {
	At time.Time `db:"created_at" json:"created_at"`
	By *int64    `db:"created_by" json:"created_by,omitempty"`
}

type AuditUpdate struct {
	At *time.Time `db:"updated_at" json:"updated_at,omitempty"`
	By *int64     `db:"updated_by" json:"updated_by,omitempty"`
}

type AuditDelete struct {
	At *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	By *int64     `db:"deleted_by" json:"deleted_by,omitempty"`
}

type AuditCreateUpdateDelete struct {
	AuditCreate
	AuditUpdate
	AuditDelete
}
