package management_test

import (
	"context"
	"encoding/json"
	"math"
	"testing"

	applicationmanagementrole "github.com/MateWorkspace/lockmate/internal/application/management/role"
	applicationmanagementspacememberrole "github.com/MateWorkspace/lockmate/internal/application/management/space_member_role"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
)

func TestFiltersPreserveEmptyPageTotalAndRejectOverflow(t *testing.T) {
	f := testshared.NewFixture()
	called := false
	f.Role.ReadByFilterFn = func(ctx context.Context, spaceId int64, filter domaincontractsrepository.RoleFilter) ([]domainmodelsentity.Role, int64, error) {
		called = true
		if spaceId != 10 || filter.Page != 0 || filter.Limit != 0 {
			t.Fatal("pagination normalized incorrectly")
		}
		return nil, 16, nil
	}
	u := applicationmanagementrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Role, f.Guard, f.Validator, f.Logger, f.Transactor, "")
	result, err := u.ReadByFilter(context.Background(), f.Caller, 10, domainusecasesmanagement.FilterRoleRequest{})
	if err != nil || !called || result.Items == nil || result.Total != 16 || len(f.Logger.Entries) != 0 {
		t.Fatalf("page semantics: %v", err)
	}
	called = false
	_, err = u.ReadByFilter(context.Background(), f.Caller, 10, domainusecasesmanagement.FilterRoleRequest{Page: math.MaxInt64, Limit: 2})
	testshared.RequireError(t, err, domainmodelserror.BadArgs)
	if called {
		t.Fatal("overflow reached repository")
	}
}
func TestInvalidPreferencesDoNotReachWrite(t *testing.T) {
	f := testshared.NewFixture()
	raw := json.RawMessage("invalid-json")
	u := applicationmanagementrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Role, f.Guard, f.Validator, f.Logger, f.Transactor, "")
	err := u.UpdateById(context.Background(), f.Caller, 10, 60, domainusecasesmanagement.UpdateRoleRequest{Preferences: &raw})
	testshared.RequireError(t, err, domainmodelserror.BadArgs)
	if f.Committed || len(f.Logger.Entries) != 2 || f.Logger.Entries[0].Level != "error" {
		t.Fatal("invalid write committed or logged success")
	}
}
func TestBulkSelectorsAndAuditContext(t *testing.T) {
	f := testshared.NewFixture()
	called := false
	f.SpaceMemberRole.DeleteBySpaceMemberIdOrRoleIdFn = func(ctx context.Context, spaceId int64, args domaincontractsrepository.DeleteSpaceMemberRoleArgs) error {
		testshared.RequireTx(t, ctx)
		called = true
		if spaceId != 10 || args.SpaceMemberId == nil || *args.SpaceMemberId != 20 || args.RoleId == nil || *args.RoleId != 60 {
			t.Fatal("bulk selectors lost")
		}
		return nil
	}
	u := applicationmanagementspacememberrole.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.SpaceMemberRole, f.Guard, f.Logger, f.Transactor, "")
	err := u.DeleteBySpaceMemberIdOrRoleId(context.Background(), f.Caller, 10, domainusecasesmanagement.DeleteSpaceMemberRoleRequest{})
	testshared.RequireError(t, err, domainmodelserror.BadArgs)
	if called {
		t.Fatal("empty selector reached repository")
	}
	memberId, roleId := int64(20), int64(60)
	err = u.DeleteBySpaceMemberIdOrRoleId(context.Background(), f.Caller, 10, domainusecasesmanagement.DeleteSpaceMemberRoleRequest{SpaceMemberId: &memberId, RoleId: &roleId})
	if err != nil || !called || !f.Committed {
		t.Fatalf("bulk deletion: %v", err)
	}
}
