package misc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	applicationmisctracer "github.com/MateWorkspace/lockmate/internal/application/misc/tracer"
	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructureutilitytracer "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/tracer"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
	"github.com/google/uuid"
)

func TestTracerGeneratesEventAndPreservesContext(t *testing.T) {
	tracer := infrastructureutilitytracer.NewContextImpl()
	logger := &testshared.Logger{}
	u := applicationmisctracer.NewUsecaseImpl(tracer, logger)
	parent, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	parent = tracer.Inject(parent, domainmodelsapp.Trace{EventId: uuid.New(), ClientSlug: "previous"})
	executionId := uuid.New()
	first, err := u.Inject(parent, "web", executionId)
	if err != nil {
		t.Fatal(err)
	}
	second, err := u.Inject(parent, "", uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	a, err := tracer.Extract(first)
	if err != nil {
		t.Fatal(err)
	}
	b, err := tracer.Extract(second)
	if err != nil {
		t.Fatal(err)
	}
	if a.EventId == uuid.Nil || a.EventId == b.EventId || a.EventId.Version() != 7 || a.ClientSlug != "web" || a.ClientExecId != executionId || b.ClientSlug != "" || b.ClientExecId != uuid.Nil {
		t.Fatalf("wrong trace injection: %v %v", a, b)
	}
	original, _ := tracer.Extract(parent)
	if original.ClientSlug != "previous" {
		t.Fatal("parent trace mutated")
	}
	deadline, _ := parent.Deadline()
	got, _ := first.Deadline()
	if !got.Equal(deadline) {
		t.Fatal("deadline changed")
	}
	cancel()
	if !errors.Is(first.Err(), context.Canceled) {
		t.Fatal("parent cancellation lost")
	}
	if len(logger.Entries) != 0 {
		t.Fatal("tracer success logged")
	}
}
func TestTracerGenerationFailureDoesNotInject(t *testing.T) {
	source := errors.New("private entropy error")
	want := domainmodelserror.NewError("failed to generate event id", domainmodelserror.TraceEventIdFailed, source)
	tracer := &testshared.Tracer{GenerateEventIdFn: func() (uuid.UUID, error) { return uuid.Nil, want }}
	logger := &testshared.Logger{}
	result, err := applicationmisctracer.NewUsecaseImpl(tracer, logger).Inject(context.Background(), "web", uuid.New())
	if result != nil || err != want || !errors.Is(err, source) || len(logger.Entries) != 1 || logger.Entries[0].Level != "error" {
		t.Fatal("generation failure did not propagate safely")
	}
}
