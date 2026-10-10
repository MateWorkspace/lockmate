package management_test

import (
	"context"
	"errors"
	"testing"

	applicationmanagementclient "github.com/MateWorkspace/lockmate/internal/application/management/client"
	applicationmiscallowedorigin "github.com/MateWorkspace/lockmate/internal/application/misc/allowed_origin"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	infrastructureutilitypermissionguard "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/permission_guard"
	infrastructureutilityvalidator "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/validator"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
)

func TestPostgresClientOriginManagementAndAllowlist(t *testing.T) {
	f := testshared.NewPostgresFixture(t)
	ctx := context.Background()
	spaceId, err := f.Space.Create(ctx, domaincontractsrepository.CreateSpaceArgs{Slug: "lockmate", Name: "Lockmate"})
	if err != nil {
		t.Fatal(err)
	}
	yes, no := true, false
	_, err = f.Role.Create(ctx, spaceId, domaincontractsrepository.CreateRoleArgs{Slug: "admin", Name: "Admin", IsDefault: &yes})
	if err != nil {
		t.Fatal(err)
	}
	userId, err := f.User.Create(ctx, domaincontractsrepository.CreateUserArgs{Name: "Origin Admin"})
	if err != nil {
		t.Fatal(err)
	}
	memberId, err := f.SpaceMember.Create(ctx, spaceId, domaincontractsrepository.CreateSpaceMemberArgs{UserId: userId})
	if err != nil {
		t.Fatal(err)
	}
	clientId, err := f.Client.Create(ctx, spaceId, domaincontractsrepository.CreateClientArgs{Slug: "web", Name: "Web", SecretKeyHash: "hash", SecretKeyRedacted: "masked"})
	if err != nil {
		t.Fatal(err)
	}
	otherClientId, err := f.Client.Create(ctx, spaceId, domaincontractsrepository.CreateClientArgs{Slug: "other-web", Name: "Other Web", SecretKeyHash: "hash", SecretKeyRedacted: "masked"})
	if err != nil {
		t.Fatal(err)
	}
	caller := &domainmodelsentity.AccessClaims{
		UserId: userId, SpaceId: spaceId, SpaceMemberId: memberId,
		Permissions: []string{"client_origin:add", "client_origin:get", "client_origin:set", "client_origin:remove"},
	}
	validator := infrastructureutilityvalidator.NewValidatorImpl()
	guard := infrastructureutilitypermissionguard.NewClaimsImpl()
	logger := &testshared.Logger{}
	newManagement := func(transactor domaincontractsutility.Transactor) domainusecasesmanagement.Client {
		return applicationmanagementclient.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Client, f.ClientOrigin, guard, validator, logger, transactor, &testshared.Keygen{}, "")
	}
	u := newManagement(f.Transactor)
	allowed := applicationmiscallowedorigin.NewUsecaseImpl(f.ClientOrigin, validator, logger)
	origin := "https://app.example.com"
	id, err := u.CreateOrigin(ctx, caller, spaceId, clientId, domainusecasesmanagement.CreateClientOriginRequest{Origin: origin})
	if err != nil {
		t.Fatal(err)
	}
	assertAllowed := func(value string, expected bool) {
		t.Helper()
		global, err := allowed.IsAllowedOrigin(ctx, value)
		if err != nil || global != expected {
			t.Fatalf("incorrect global allowance: %v %v", global, err)
		}
		scoped, err := allowed.IsAllowedOriginForClient(ctx, spaceId, clientId, value)
		if err != nil || scoped != expected {
			t.Fatalf("incorrect client allowance: %v %v", scoped, err)
		}
	}
	assertAllowed(origin, true)
	item, err := u.ReadOriginById(ctx, caller, spaceId, clientId, id)
	if err != nil || item.ClientId != clientId || item.AuditCreate.By == nil || *item.AuditCreate.By != userId {
		t.Fatalf("incorrect created origin: %+v %v", item, err)
	}
	byValue, err := u.ReadOriginByValue(ctx, caller, spaceId, clientId, origin)
	if err != nil || byValue.Id != id {
		t.Fatalf("origin lookup failed: %v", err)
	}
	items, err := u.ReadOriginsByClientId(ctx, caller, spaceId, clientId)
	if err != nil || len(items) != 1 {
		t.Fatalf("origin list failed: %v", err)
	}
	page, err := u.ReadOriginsByFilter(ctx, caller, spaceId, clientId, domainusecasesmanagement.FilterClientOriginRequest{Page: 1, Limit: 10})
	if err != nil || len(page.Items) != 1 || page.Total != 1 {
		t.Fatalf("origin page failed: %v", err)
	}
	_, err = u.ReadOriginById(ctx, caller, spaceId, otherClientId, id)
	testshared.RequireError(t, err, domainmodelserror.ClientOriginNotFound)
	err = u.UpdateOriginById(ctx, caller, spaceId, otherClientId, id, domainusecasesmanagement.UpdateClientOriginRequest{IsActive: &no})
	testshared.RequireError(t, err, domainmodelserror.ClientOriginNotFound)
	err = u.DeleteOriginById(ctx, caller, spaceId, otherClientId, id)
	testshared.RequireError(t, err, domainmodelserror.ClientOriginNotFound)
	assertAllowed(origin, true)

	// Force failure after real repository writes, before committing the transaction.
	rollback := errors.New("application operation aborted")
	failing := newManagement(&testshared.Transactor{WithTxFn: func(ctx context.Context, fn func(context.Context) error) error {
		return f.Transactor.WithTx(ctx, func(txCtx context.Context) error {
			if err := fn(txCtx); err != nil {
				return err
			}
			return rollback
		})
	}})
	created, err := failing.CreateOrigin(ctx, caller, spaceId, clientId, domainusecasesmanagement.CreateClientOriginRequest{Origin: "https://rollback.example.com"})
	if created != 0 || !errors.Is(err, rollback) {
		t.Fatalf("rollback returned a creation result: %d %v", created, err)
	}
	assertAllowed("https://rollback.example.com", false)
	err = failing.UpdateOriginById(ctx, caller, spaceId, clientId, id, domainusecasesmanagement.UpdateClientOriginRequest{IsActive: &no})
	testshared.RequireError(t, err, rollback)
	assertAllowed(origin, true)
	err = failing.DeleteOriginById(ctx, caller, spaceId, clientId, id)
	testshared.RequireError(t, err, rollback)
	assertAllowed(origin, true)

	err = u.UpdateOriginById(ctx, caller, spaceId, clientId, id, domainusecasesmanagement.UpdateClientOriginRequest{IsActive: &no})
	if err != nil {
		t.Fatal(err)
	}
	assertAllowed(origin, false)
	replacement := "https://preview.example.com"
	err = u.UpdateOriginById(ctx, caller, spaceId, clientId, id, domainusecasesmanagement.UpdateClientOriginRequest{Origin: &replacement, IsActive: &yes})
	if err != nil {
		t.Fatal(err)
	}
	assertAllowed(origin, false)
	assertAllowed(replacement, true)

	err = f.Client.UpdateById(ctx, spaceId, clientId, domaincontractsrepository.UpdateClientArgs{IsActive: &no})
	if err != nil {
		t.Fatal(err)
	}
	assertAllowed(replacement, false)
	_, err = u.ReadOriginById(ctx, caller, spaceId, clientId, id)
	if err != nil {
		t.Fatal("inactive client became unmanageable", err)
	}
	err = f.Client.UpdateById(ctx, spaceId, clientId, domaincontractsrepository.UpdateClientArgs{IsActive: &yes})
	if err != nil {
		t.Fatal(err)
	}
	assertAllowed(replacement, true)

	err = u.DeleteOriginById(ctx, caller, spaceId, clientId, id)
	if err != nil {
		t.Fatal(err)
	}
	assertAllowed(replacement, false)
	_, err = u.ReadOriginById(ctx, caller, spaceId, clientId, id)
	testshared.RequireError(t, err, domainmodelserror.ClientOriginNotFound)
	page, err = u.ReadOriginsByFilter(ctx, caller, spaceId, clientId, domainusecasesmanagement.FilterClientOriginRequest{Limit: 10})
	if err != nil || page.Items == nil || len(page.Items) != 0 || page.Total != 0 {
		t.Fatalf("deleted origin remained in management list: %+v %v", page, err)
	}

	// Existing client permissions do not grant origin administration.
	caller.Permissions = []string{"client:add", "client:get", "client:set", "client:remove"}
	_, err = u.CreateOrigin(ctx, caller, spaceId, clientId, domainusecasesmanagement.CreateClientOriginRequest{Origin: origin})
	testshared.RequireError(t, err, domainmodelserror.Forbidden)
}
