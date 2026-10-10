package infrastructureutilitylogger

import (
	"context"
	"log/slog"

	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
)

type slogImpl struct {
	logger   *slog.Logger
	tracer   domaincontractsutility.Tracer
	level    domainmodelslogger.Level
	useTrace bool
}

func NewSlogImpl(
	logger *slog.Logger,
	tracer domaincontractsutility.Tracer,
	level domainmodelslogger.Level,
	useTrace bool,
) domaincontractsutility.Logger {
	return &slogImpl{
		logger:   logger,
		tracer:   tracer,
		level:    level,
		useTrace: useTrace,
	}
}

func (s *slogImpl) Error(ctx context.Context, tag string, message string, meta domainmodelslogger.Meta) {
	if s.level.Order() < domainmodelslogger.LevelError.Order() {
		return
	}

	s.logger.LogAttrs(ctx, slog.LevelError, message, s.buildAttrs(ctx, tag, meta)...)
}

func (s *slogImpl) Warn(ctx context.Context, tag string, message string, meta domainmodelslogger.Meta) {
	if s.level.Order() < domainmodelslogger.LevelWarn.Order() {
		return
	}

	s.logger.LogAttrs(ctx, slog.LevelWarn, message, s.buildAttrs(ctx, tag, meta)...)
}

func (s *slogImpl) Info(ctx context.Context, tag string, message string, meta domainmodelslogger.Meta) {
	if s.level.Order() < domainmodelslogger.LevelInfo.Order() {
		return
	}

	s.logger.LogAttrs(ctx, slog.LevelInfo, message, s.buildAttrs(ctx, tag, meta)...)
}

func (s *slogImpl) Debug(ctx context.Context, tag string, message string, meta domainmodelslogger.Meta) {
	if s.level.Order() < domainmodelslogger.LevelDebug.Order() {
		return
	}

	s.logger.LogAttrs(ctx, slog.LevelDebug, message, s.buildAttrs(ctx, tag, meta)...)
}

func (s *slogImpl) buildAttrs(ctx context.Context, tag string, meta domainmodelslogger.Meta) []slog.Attr {
	attrs := []slog.Attr{
		slog.String("tag", tag),
		slog.Any("meta", normalizeMeta(meta)),
	}

	if s.useTrace {
		trace, err := s.tracer.Extract(ctx)
		if err != nil || trace == nil {
			trace = &domainmodelsapp.Trace{}
		}

		attrs = append(attrs, slog.Any("trace", *trace))
	}

	return attrs
}
