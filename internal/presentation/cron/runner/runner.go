package presentationcronrunner

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
)

// Schedule pairs a job with its execution interval. Optional jobs never fail
// startup; their initial and periodic errors are the job's responsibility.
type Schedule struct {
	Name     string
	Interval time.Duration
	Optional bool
	Run      func(ctx context.Context) error
}

type Args struct {
	Timeout time.Duration
	Jobs    []Schedule
}

// Run executes every job once synchronously (surfacing a required failure), then
// schedules each job with the cron engine on its interval. The returned stop
// function halts the scheduler and waits for running jobs to finish.
func Run(ctx context.Context, args Args) (stop func(), err error) {
	for _, job := range args.Jobs {
		if err := runOnce(ctx, job, args.Timeout); err != nil && !job.Optional {
			return nil, err
		}
	}

	scheduler := cron.New()
	for _, job := range args.Jobs {
		if job.Interval < time.Second {
			continue
		}

		job := job
		if _, err := scheduler.AddFunc("@every "+job.Interval.String(), func() {
			_ = runOnce(ctx, job, args.Timeout)
		}); err != nil {
			return nil, err
		}
	}
	scheduler.Start()

	return func() {
		<-scheduler.Stop().Done()
	}, nil
}

func runOnce(ctx context.Context, job Schedule, timeout time.Duration) error {
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return job.Run(runCtx)
}
