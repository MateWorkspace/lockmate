package domainmodelsapp

import "github.com/google/uuid"

type Trace struct {
	EventId      uuid.UUID `json:"event_id"`
	ClientExecId uuid.UUID `json:"client_exec_id"`
	ClientSlug   string    `json:"client_slug"`
}
