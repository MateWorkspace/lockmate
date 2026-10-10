package utility_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	infrastructureutilitylogger "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/logger"
	infrastructureutilitytracer "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/tracer"
	"github.com/google/uuid"
)

func TestLoggerJSONTraceAndMeta(t *testing.T) {
	for _, present := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing", true: "present"}[present], func(t *testing.T) {
			var output bytes.Buffer
			tracer := infrastructureutilitytracer.NewContextImpl()
			logger := infrastructureutilitylogger.NewSlogImpl(slog.New(slog.NewJSONHandler(&output, nil)), tracer, domainmodelslogger.LevelInfo, true)
			ctx := context.Background()
			want := domainmodelsapp.Trace{}
			if present {
				want = domainmodelsapp.Trace{EventId: uuid.New(), ClientExecId: uuid.New(), ClientSlug: "taskmate-web"}
				ctx = tracer.Inject(ctx, want)
			}
			failure := errors.New("failure message")
			meta := domainmodelslogger.Meta{"error": failure, "count": 3}
			logger.Info(ctx, "test/operation", "operation complete", meta)
			var record struct {
				Level   string            `json:"level"`
				Tag     string            `json:"tag"`
				Message string            `json:"msg"`
				Meta    map[string]any    `json:"meta"`
				Trace   map[string]string `json:"trace"`
			}
			if err := json.Unmarshal(output.Bytes(), &record); err != nil {
				t.Fatal(err)
			}
			if record.Level != "INFO" || record.Tag != "test/operation" || record.Message != "operation complete" {
				t.Fatalf("unexpected log: %s", output.String())
			}
			if record.Trace["event_id"] != want.EventId.String() || record.Trace["client_exec_id"] != want.ClientExecId.String() || record.Trace["client_slug"] != want.ClientSlug || len(record.Trace) != 3 {
				t.Fatalf("unexpected trace: %v", record.Trace)
			}
			if record.Meta["error"] != failure.Error() || record.Meta["count"] != float64(3) || meta["error"] != failure {
				t.Fatalf("unexpected metadata: %v", record.Meta)
			}
		})
	}
}

func TestLoggerTraceDisabledAndEmptyMeta(t *testing.T) {
	var output bytes.Buffer
	logger := infrastructureutilitylogger.NewSlogImpl(slog.New(slog.NewJSONHandler(&output, nil)), nil, domainmodelslogger.LevelInfo, false)
	logger.Info(context.Background(), "test", "message", nil)
	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if _, exists := record["trace"]; exists {
		t.Fatal("trace printed when disabled")
	}
	if meta, ok := record["meta"].(map[string]any); !ok || len(meta) != 0 {
		t.Fatalf("expected empty object metadata, got %v", record["meta"])
	}
}

func TestLoggerLevels(t *testing.T) {
	for _, level := range []domainmodelslogger.Level{domainmodelslogger.LevelNone, domainmodelslogger.LevelError, domainmodelslogger.LevelWarn, domainmodelslogger.LevelInfo, domainmodelslogger.LevelDebug} {
		t.Run(level.String(), func(t *testing.T) {
			var output bytes.Buffer
			logger := infrastructureutilitylogger.NewSlogImpl(slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug})), nil, level, false)
			ctx := context.Background()
			logger.Error(ctx, "test", "error", nil)
			logger.Warn(ctx, "test", "warn", nil)
			logger.Info(ctx, "test", "info", nil)
			logger.Debug(ctx, "test", "debug", nil)
			if count := bytes.Count(output.Bytes(), []byte("\n")); count != level.Order() {
				t.Fatalf("expected %d records, got %d", level.Order(), count)
			}
		})
	}
}

func TestLoggerPlain(t *testing.T) {
	var output bytes.Buffer
	tracer := infrastructureutilitytracer.NewContextImpl()
	logger := infrastructureutilitylogger.NewSlogImpl(slog.New(slog.NewTextHandler(&output, nil)), tracer, domainmodelslogger.LevelInfo, true)
	ctx := tracer.Inject(context.Background(), domainmodelsapp.Trace{ClientSlug: "taskmate-web"})
	logger.Info(ctx, "test/operation", "operation complete", nil)
	for _, part := range []string{"level=INFO", "tag=test/operation", "meta=map[]", "trace=", "taskmate-web"} {
		if !strings.Contains(output.String(), part) {
			t.Fatalf("missing %q in plain output: %s", part, output.String())
		}
	}
}
