package seeder_test

import (
	"context"
	"errors"
	"testing"

	applicationseeder "github.com/MateWorkspace/lockmate/internal/application/seeder"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructureutilitykeygen "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/keygen"
	infrastructureutilityvalidator "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/validator"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
)

func TestSeedFailuresStopWithoutSuccessLogs(t *testing.T) {
	unavailable := errors.New("database unavailable")
	for _, test := range []struct {
		name     string
		failure  error
		attempts int
		create   bool
		levels   []string
	}{
		{"read failure", unavailable, 1, false, []string{"error", "error"}},
		{"exhausted conflict", domainmodelserror.SpaceSlugConflict, 2, true, []string{"warn", "warn", "error", "error"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			data := definitions(t)
			attempts := 0
			space := &testshared.Space{
				ReadBySlugFn: func(context.Context, string) (*domainmodelsentity.Space, error) {
					if test.create {
						return nil, domainmodelserror.SpaceNotFound
					}

					return nil, test.failure
				},
				CreateFn: func(context.Context, domaincontractsrepository.CreateSpaceArgs) (int64, error) {
					return 0, test.failure
				},
			}
			transactor := &testshared.Transactor{
				WithTxFn: func(ctx context.Context, fn func(context.Context) error) error {
					attempts++
					return fn(ctx)
				},
			}
			logger := &testshared.Logger{}
			usecase := applicationseeder.NewUsecaseImpl(
				space,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				infrastructureutilitykeygen.NewGeneratorImpl(),
				infrastructureutilityvalidator.NewValidatorImpl(),
				transactor,
				logger,
				&data,
			)
			if err := usecase.Run(context.Background()); !errors.Is(err, test.failure) {
				t.Fatalf("expected original failure, got %v", err)
			}

			if attempts != test.attempts || len(logger.Entries) != len(test.levels) {
				t.Fatalf("attempts=%d logs=%v", attempts, logger.Entries)
			}

			for index, level := range test.levels {
				if logger.Entries[index].Level != level {
					t.Fatalf("log %d: expected %s, got %s", index, level, logger.Entries[index].Level)
				}
			}
		})
	}
}
