package management_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	applicationmanagementpermission "github.com/MateWorkspace/lockmate/internal/application/management/permission"
	applicationmanagementrole "github.com/MateWorkspace/lockmate/internal/application/management/role"
	applicationmanagementspace "github.com/MateWorkspace/lockmate/internal/application/management/space"
	applicationmanagementspacemember "github.com/MateWorkspace/lockmate/internal/application/management/space_member"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
)

func TestUpdateSuccessLoggingPolicy(t *testing.T) {
	cases := []struct {
		name string
		call func(*testshared.Fixture, bool) error
	}{
		{"space", func(f *testshared.Fixture, important bool) error {
			name, active := "Updated space", false
			request := domainusecasesmanagement.UpdateSpaceRequest{Name: &name}
			if important {
				request.IsActive = &active
			}

			f.Space.UpdateByIdFn = func(ctx context.Context, id int64, args domaincontractsrepository.UpdateSpaceArgs) error {
				testshared.RequireTx(t, ctx)
				if len(f.Logger.Entries) != 0 {
					t.Fatal("logged before commit")
				}
				return nil
			}
			u := applicationmanagementspace.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Guard, f.Validator, f.Logger, f.Transactor, "")
			return u.UpdateById(context.Background(), f.Caller, 10, request)
		}},
		{"role", func(f *testshared.Fixture, important bool) error {
			name, defaultRole := "Updated role", false
			request := domainusecasesmanagement.UpdateRoleRequest{Name: &name}
			if important {
				request.IsDefault = &defaultRole
			}
			f.Role.UpdateByIdFn = func(ctx context.Context, spaceId, id int64, args domaincontractsrepository.UpdateRoleArgs) error {
				testshared.RequireTx(t, ctx)
				if len(f.Logger.Entries) != 0 {
					t.Fatal("logged before commit")
				}
				return nil
			}
			u := applicationmanagementrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Role, f.Guard, f.Validator, f.Logger, f.Transactor, "")
			return u.UpdateById(context.Background(), f.Caller, 10, 60, request)
		}},
		{"client", func(f *testshared.Fixture, important bool) error {
			name, active := "Updated client", false
			request := domainusecasesmanagement.UpdateClientRequest{Name: &name}
			if important {
				request.IsActive = &active
			}
			f.Client.UpdateByIdFn = func(ctx context.Context, spaceId, id int64, args domaincontractsrepository.UpdateClientArgs) error {
				testshared.RequireTx(t, ctx)
				if len(f.Logger.Entries) != 0 {
					t.Fatal("logged before commit")
				}
				return nil
			}
			return client(f).UpdateById(context.Background(), f.Caller, 10, 30, request)
		}},
		{"membership", func(f *testshared.Fixture, important bool) error {
			active := false
			request := domainusecasesmanagement.UpdateSpaceMemberRequest{}
			if important {
				request.IsActive = &active
			}
			f.SpaceMember.UpdateByIdFn = func(ctx context.Context, spaceId, id int64, args domaincontractsrepository.UpdateSpaceMemberArgs) error {
				testshared.RequireTx(t, ctx)
				if len(f.Logger.Entries) != 0 {
					t.Fatal("logged before commit")
				}
				return nil
			}
			u := applicationmanagementspacemember.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Logger, f.Transactor, "")
			return u.UpdateById(context.Background(), f.Caller, 10, 20, request)
		}},
	}

	for _, test := range cases {
		for _, important := range []bool{false, true} {
			for _, failCommit := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/important=%t/commit_failure=%t", test.name, important, failCommit), func(t *testing.T) {
					f := testshared.NewFixture()
					if failCommit {
						f.CommitErr = domainmodelserror.NewError("private-message", domainmodelserror.Failure, errors.New("private-source"))
					}
					err := test.call(f, important)
					if failCommit {
						if err != f.CommitErr {
							t.Fatalf("commit error changed: %v", err)
						}
					} else if err != nil || !f.Committed {
						t.Fatalf("update failed: %v", err)
					}

					expectedLevel := ""
					if failCommit {
						expectedLevel = "error"
					} else if important {
						expectedLevel = "info"
					}
					if expectedLevel == "" {
						if len(f.Logger.Entries) != 0 {
							t.Fatalf("cosmetic update logged: %v", f.Logger.Entries)
						}
					} else if len(f.Logger.Entries) != 1 || f.Logger.Entries[0].Level != expectedLevel {
						t.Fatalf("unexpected logs: %v", f.Logger.Entries)
					}
					if failCommit && f.Logger.Entries[0].Meta["err"] != f.CommitErr {
						t.Fatal("commit error missing from metadata")
					}
				})
			}
		}
	}
}

func TestPermissionMutationLogging(t *testing.T) {
	for _, operation := range []string{"create", "update", "delete"} {
		for _, failCommit := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/commit_failure=%t", operation, failCommit), func(t *testing.T) {
				f := testshared.NewFixture()
				if failCommit {
					f.CommitErr = testshared.NotFound(domainmodelserror.Failure)
				}
				check := func(ctx context.Context) {
					testshared.RequireTx(t, ctx)
					if len(f.Logger.Entries) != 0 {
						t.Fatal("logged before commit")
					}
				}
				f.Permission.CreateFn = func(ctx context.Context, spaceId int64, args domaincontractsrepository.CreatePermissionArgs) (int64, error) {
					check(ctx)
					return 70, nil
				}
				f.Permission.UpdateByIdFn = func(ctx context.Context, spaceId, id int64, args domaincontractsrepository.UpdatePermissionArgs) error {
					check(ctx)
					return nil
				}
				f.Permission.DeleteByIdFn = func(ctx context.Context, spaceId, id int64, by *int64) error { check(ctx); return nil }
				u := applicationmanagementpermission.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Permission, f.Guard, f.Validator, f.Logger, f.Transactor, "")
				name := "Readable permission"
				var err error
				switch operation {
				case "create":
					_, err = u.Create(context.Background(), f.Caller, 10, domainusecasesmanagement.CreatePermissionRequest{Slug: "test:get", Name: name})
				case "update":
					err = u.UpdateById(context.Background(), f.Caller, 10, 70, domainusecasesmanagement.UpdatePermissionRequest{Name: &name})
				case "delete":
					err = u.DeleteById(context.Background(), f.Caller, 10, 70)
				}
				if failCommit {
					testshared.RequireError(t, err, domainmodelserror.Failure)
					if len(f.Logger.Entries) != 1 || f.Logger.Entries[0].Level != "error" {
						t.Fatalf("wrong failure logging: %v", f.Logger.Entries)
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					if operation == "update" {
						if len(f.Logger.Entries) != 0 {
							t.Fatal("cosmetic update logged")
						}
					} else if len(f.Logger.Entries) != 1 || f.Logger.Entries[0].Level != "info" {
						t.Fatalf("missing mutation log: %v", f.Logger.Entries)
					}
				}
			})
		}
	}
}

func TestRoleCreationLogsSpecificFailureAndTransactionContext(t *testing.T) {
	for _, stage := range []string{"slug", "repository", "begin", "commit"} {
		t.Run(stage, func(t *testing.T) {
			f := testshared.NewFixture()
			request := domainusecasesmanagement.CreateRoleRequest{Slug: "staff", Name: "Staff"}
			failure := domainmodelserror.NewError("role persistence failed", domainmodelserror.Failure, errors.New("database detail"))
			f.Role.CreateFn = func(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateRoleArgs) (int64, error) {
				testshared.RequireTx(t, ctx)
				if len(f.Logger.Entries) != 0 {
					t.Fatal("logged success before commit")
				}
				if stage == "repository" {
					return 0, failure
				}
				return 60, nil
			}
			expectedMessage := ""
			expectedCount := 1
			switch stage {
			case "slug":
				request.Slug = ""
				expectedMessage = "invalid role slug"
				expectedCount = 2
			case "repository":
				expectedMessage = "failed to create role"
				expectedCount = 2
			case "begin":
				f.BeginErr = failure
			case "commit":
				f.CommitErr = failure
			}
			u := applicationmanagementrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Role, f.Guard, f.Validator, f.Logger, f.Transactor, "")
			result, err := u.Create(context.Background(), f.Caller, 10, request)
			if err == nil || result != 0 || f.Committed {
				t.Fatal("failed creation returned result or committed")
			}
			if len(f.Logger.Entries) != expectedCount {
				t.Fatalf("wrong logs: %v", f.Logger.Entries)
			}
			if expectedMessage != "" && f.Logger.Entries[0].Message != expectedMessage {
				t.Fatalf("missing step message: %v", f.Logger.Entries)
			}
			if f.Logger.Entries[expectedCount-1].Message != "failed to complete role creation transaction" {
				t.Fatal("missing transaction message")
			}
			for _, entry := range f.Logger.Entries {
				if entry.Level != "error" || entry.Meta["err"] != err || entry.Meta["space_id"] != int64(10) || entry.Meta["caller_user_id"] != int64(1) {
					t.Fatalf("incorrect contextual error: %v", entry)
				}
			}
		})
	}
}
