package presentationhttpresponse

import (
	"encoding/json"
	"time"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type IdResponse struct {
	Id int64 `json:"id" example:"1"`
}

type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid Format"`
	Message string `json:"message" example:"The request is invalid."`
}

type PageResponse struct {
	Page       int64 `json:"page" example:"1"`
	Limit      int64 `json:"limit" example:"10"`
	TotalItems int64 `json:"total_items" example:"87"`
}

type PageDataResponse[T any] struct {
	Data []T          `json:"data"`
	Page PageResponse `json:"page"`
}

type AuditCreateResponse struct {
	CreatedAt time.Time `json:"created_at" example:"2026-10-10T09:30:00Z"`
	CreatedBy *int64    `json:"created_by,omitempty" example:"1"`
}

type AuditResponse struct {
	AuditCreateResponse
	UpdatedAt *time.Time `json:"updated_at,omitempty" example:"2026-10-10T09:30:00Z"`
	UpdatedBy *int64     `json:"updated_by,omitempty" example:"1"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" example:"2026-10-10T09:30:00Z"`
	DeletedBy *int64     `json:"deleted_by,omitempty" example:"1"`
}

func AuditCreate(value domainmodelsentity.AuditCreate) AuditCreateResponse {
	return AuditCreateResponse{CreatedAt: value.At, CreatedBy: value.By}
}

func Audit(value domainmodelsentity.AuditCreateUpdateDelete) AuditResponse {
	return AuditResponse{
		AuditCreateResponse: AuditCreate(value.AuditCreate),
		UpdatedAt:           value.AuditUpdate.At, UpdatedBy: value.AuditUpdate.By,
		DeletedAt: value.AuditDelete.At, DeletedBy: value.AuditDelete.By,
	}
}

func NormalizeJSON(value json.RawMessage) json.RawMessage {
	if len(value) == 0 {
		return json.RawMessage(`{}`)
	}
	return value
}
