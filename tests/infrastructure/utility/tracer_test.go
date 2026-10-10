package utility_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructureutilitytracer "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/tracer"
	"github.com/google/uuid"
)

func TestTracerMissing(t *testing.T) {
	tracer := infrastructureutilitytracer.NewContextImpl()
	trace, err := tracer.Extract(context.Background())
	if trace != nil || !errors.Is(err, domainmodelserror.TraceNotFound) {
		t.Fatalf("expected missing trace error, got trace=%v err=%v", trace, err)
	}

	var model *domainmodelserror.Error
	if !errors.As(err, &model) {
		t.Fatalf("expected domain error, got %T", err)
	}
}

func TestTracerContextAndCopy(t *testing.T) {
	tracer := infrastructureutilitytracer.NewContextImpl()
	deadline := time.Now().Add(time.Minute)
	parent, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	want := domainmodelsapp.Trace{
		EventId:      uuid.New(),
		ClientExecId: uuid.New(),
		ClientSlug:   "taskmate-web",
	}
	ctx := tracer.Inject(parent, want)
	want.ClientSlug = "modified"
	trace, err := tracer.Extract(ctx)
	if err != nil || trace.ClientSlug != "taskmate-web" || trace.EventId != want.EventId || trace.ClientExecId != want.ClientExecId {
		t.Fatalf("trace was not preserved: trace=%v err=%v", trace, err)
	}
	trace.ClientSlug = "changed-copy"
	stored, err := tracer.Extract(ctx)
	if err != nil || stored.ClientSlug != "taskmate-web" {
		t.Fatalf("extract mutated stored trace: trace=%v err=%v", stored, err)
	}
	if got, ok := ctx.Deadline(); !ok || !got.Equal(deadline) {
		t.Fatal("parent deadline was not preserved")
	}
	if _, err := tracer.Extract(parent); !errors.Is(err, domainmodelserror.TraceNotFound) {
		t.Fatal("inject mutated the parent context")
	}
	cancel()
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatal("parent cancellation was not preserved")
	}
}

func TestTracerReplacementAndZeroValue(t *testing.T) {
	tracer := infrastructureutilitytracer.NewContextImpl()
	ctx := tracer.Inject(context.Background(), domainmodelsapp.Trace{ClientSlug: "original"})
	replaced := tracer.Inject(ctx, domainmodelsapp.Trace{})
	trace, err := tracer.Extract(replaced)
	if err != nil || *trace != (domainmodelsapp.Trace{}) {
		t.Fatalf("expected present zero-value trace, got trace=%v err=%v", trace, err)
	}
	original, err := tracer.Extract(ctx)
	if err != nil || original.ClientSlug != "original" {
		t.Fatal("replacement mutated the parent trace")
	}
}

func TestTracerGeneratesFreshV7EventIds(t *testing.T) {
	tracer := infrastructureutilitytracer.NewContextImpl()
	first, err := tracer.GenerateEventId()
	if err != nil {
		t.Fatal(err)
	}
	second, err := tracer.GenerateEventId()
	if err != nil {
		t.Fatal(err)
	}
	if first == uuid.Nil || first == second || first.Version() != 7 || second.Version() != 7 {
		t.Fatal("expected distinct nonzero UUIDv7 event identifiers")
	}
}
