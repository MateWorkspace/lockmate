package presentationgrpcmapper

import (
	"context"
	"errors"
	"strings"

	lockmatev1 "github.com/MateWorkspace/lockmate-proto/gen/go/lockmate/v1"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Error maps a domain error to a gRPC status with a typed detail code. It never
// exposes the full diagnostic Error() source chain.
func Error(err error) error {
	if err == nil {
		return nil
	}

	var domainError *domainmodelserror.Error
	hasDomain := errors.As(err, &domainError)

	if !hasDomain {
		if errors.Is(err, context.Canceled) {
			return status.Error(codes.Canceled, "request canceled")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return status.Error(codes.DeadlineExceeded, "request deadline exceeded")
		}
		if _, ok := status.FromError(err); ok {
			return err
		}
	}

	var matched *errorMapping
	for i := range errorMappings {
		if errors.Is(err, errorMappings[i].err) {
			matched = &errorMappings[i]
			break
		}
	}
	if matched == nil {
		return status.Error(codes.Internal, "internal server error")
	}

	message := matched.err.Error()
	if hasDomain && strings.TrimSpace(domainError.Message) != "" {
		message = domainError.Message
	}

	stat := status.New(matched.grpcCode, message)
	detailed, detailErr := stat.WithDetails(&lockmatev1.Error{Code: matched.detail, Message: message})
	if detailErr != nil {
		return stat.Err()
	}

	return detailed.Err()
}
