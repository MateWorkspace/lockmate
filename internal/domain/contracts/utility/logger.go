package domaincontractsutility

import (
	"context"

	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
)

type Logger interface {
	Error(ctx context.Context, tag string, message string, meta domainmodelslogger.Meta)
	Warn(ctx context.Context, tag string, message string, meta domainmodelslogger.Meta)
	Info(ctx context.Context, tag string, message string, meta domainmodelslogger.Meta)
	Debug(ctx context.Context, tag string, message string, meta domainmodelslogger.Meta)
}
