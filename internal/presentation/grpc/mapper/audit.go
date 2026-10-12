package presentationgrpcmapper

import (
	"time"

	lockmatev1 "github.com/MateWorkspace/lockmate-proto/gen/go/lockmate/v1"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Timestamp(value time.Time) *timestamppb.Timestamp {
	if value.IsZero() {
		return nil
	}

	return timestamppb.New(value)
}

func TimestampPointer(value *time.Time) *timestamppb.Timestamp {
	if value == nil {
		return nil
	}

	return timestamppb.New(*value)
}

func Audit(value domainmodelsentity.AuditCreateUpdateDelete) *lockmatev1.AuditFields {
	return &lockmatev1.AuditFields{
		CreatedAt: Timestamp(value.AuditCreate.At),
		UpdatedAt: TimestampPointer(value.AuditUpdate.At),
		DeletedAt: TimestampPointer(value.AuditDelete.At),
		CreatedBy: value.AuditCreate.By,
		UpdatedBy: value.AuditUpdate.By,
		DeletedBy: value.AuditDelete.By,
	}
}

func AuditCreateOnly(value domainmodelsentity.AuditCreate) *lockmatev1.AuditFields {
	return &lockmatev1.AuditFields{
		CreatedAt: Timestamp(value.At),
		CreatedBy: value.By,
	}
}
