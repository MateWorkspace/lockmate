package misc_test

import (
	"context"
	"errors"
	applicationmiscauthflow "github.com/MateWorkspace/lockmate/internal/application/misc/auth_flow"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
	"testing"
)

type purgeLaunch struct {
	domaincontractsrepository.GoogleOAuthLaunch
	calls int
	fail  error
}

func (s *purgeLaunch) PurgeExpired(_ context.Context, limit int64) (int64, error) {
	if limit != 1000 {
		panic("unbounded cleanup")
	}
	s.calls++
	if s.fail != nil {
		return 0, s.fail
	}
	if s.calls == 1 {
		return 1000, nil
	}
	return 1, nil
}

type purgeAttempt struct {
	domaincontractsrepository.GoogleOAuthAttempt
	calls int
	fail  error
}

func (s *purgeAttempt) PurgeExpired(_ context.Context, limit int64) (int64, error) {
	if limit != 1000 {
		panic("unbounded cleanup")
	}
	s.calls++
	return 1, s.fail
}

type purgeHandoff struct {
	domaincontractsrepository.AuthHandoff
	calls int
	fail  error
}

func (s *purgeHandoff) PurgeExpired(_ context.Context, limit int64) (int64, error) {
	if limit != 1000 {
		panic("unbounded cleanup")
	}
	s.calls++
	return 1, s.fail
}

func TestAuthFlowCleanup(t *testing.T) {
	for _, stage := range []string{"success", "launch", "attempt", "handoff"} {
		t.Run(stage, func(t *testing.T) {
			launch := &purgeLaunch{}
			attempt := &purgeAttempt{}
			handoff := &purgeHandoff{}
			logger := &testshared.Logger{}
			failure := errors.New("cleanup unavailable")
			switch stage {
			case "launch":
				launch.fail = failure
			case "attempt":
				attempt.fail = failure
			case "handoff":
				handoff.fail = failure
			}
			err := applicationmiscauthflow.NewUsecaseImpl(launch, attempt, handoff, logger).PurgeExpired(context.Background())
			if stage == "success" {
				if err != nil || launch.calls != 2 || attempt.calls != 1 || handoff.calls != 1 || len(logger.Entries) != 0 {
					t.Fatal("cleanup did not finish bounded batches")
				}
			} else if err != failure || len(logger.Entries) != 1 || logger.Entries[0].Level != "warn" {
				t.Fatal("cleanup failure was not warned")
			}
		})
	}
}
