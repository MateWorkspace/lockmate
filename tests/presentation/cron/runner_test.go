package cron_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	presentationcronrunner "github.com/MateWorkspace/lockmate/internal/presentation/cron/runner"
)

func TestRunnerRunsStartupThenPeriodicallyAndStops(t *testing.T) {
	var startup, ticks atomic.Int64
	stop, err := presentationcronrunner.Run(context.Background(), presentationcronrunner.Args{
		Timeout: time.Second,
		Jobs: []presentationcronrunner.Schedule{
			{Name: "job", Interval: time.Second, Run: func(context.Context) error {
				if startup.Load() == 0 {
					startup.Add(1)

					return nil
				}

				ticks.Add(1)

				return nil
			}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if startup.Load() != 1 {
		t.Fatal("startup run missing")
	}

	deadline := time.Now().Add(2 * time.Second)
	for ticks.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	stop()
	if ticks.Load() == 0 {
		t.Fatal("periodic run missing")
	}
}

func TestRunnerRequiredStartupFailureAborts(t *testing.T) {
	_, err := presentationcronrunner.Run(context.Background(), presentationcronrunner.Args{
		Timeout: time.Second,
		Jobs: []presentationcronrunner.Schedule{
			{Name: "job", Interval: time.Hour, Run: func(context.Context) error { return errors.New("boom") }},
		},
	})
	if err == nil {
		t.Fatal("required startup failure ignored")
	}
}

func TestRunnerOptionalStartupFailureContinues(t *testing.T) {
	stop, err := presentationcronrunner.Run(context.Background(), presentationcronrunner.Args{
		Timeout: time.Second,
		Jobs: []presentationcronrunner.Schedule{
			{Name: "job", Interval: time.Hour, Optional: true, Run: func(context.Context) error { return errors.New("boom") }},
		},
	})
	if err != nil {
		t.Fatalf("optional failure aborted startup: %v", err)
	}
	stop()
}
