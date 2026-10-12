package management_test

import (
	"context"
	"errors"
	"testing"

	applicationmanagementclient "github.com/MateWorkspace/lockmate/internal/application/management/client"
	applicationmiscipallowlist "github.com/MateWorkspace/lockmate/internal/application/misc/ip_allowlist"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	infrastructureutilitypermissionguard "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/permission_guard"
	infrastructureutilityvalidator "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/validator"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
)

func TestPostgresClientIpAllowlistManagementAndAllowlist(t *testing.T) {
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
	userId, err := f.User.Create(ctx, domaincontractsrepository.CreateUserArgs{Name: "Allowlist Admin"})
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
		Permissions: []string{"client_ip_allowlist:add", "client_ip_allowlist:get", "client_ip_allowlist:set", "client_ip_allowlist:remove"},
	}
	validator := infrastructureutilityvalidator.NewValidatorImpl(true)
	guard := infrastructureutilitypermissionguard.NewClaimsImpl()
	logger := &testshared.Logger{}
	newManagement := func(transactor domaincontractsutility.Transactor) domainusecasesmanagement.Client {
		return applicationmanagementclient.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, f.Client, f.ClientIpAllowlist, guard, validator, logger, transactor, &testshared.Keygen{}, "")
	}
	u := newManagement(f.Transactor)
	allowed := applicationmiscipallowlist.NewUsecaseImpl(f.ClientIpAllowlist, logger)

	const originCidr, originIp = "203.0.113.0/24", "203.0.113.9"
	const replacementIp = "198.51.100.9"
	replacementCidr := "198.51.100.0/24"
	const rollbackCidr, rollbackIp = "192.0.2.0/24", "192.0.2.9"
	const outsideIp = "198.18.0.9"

	id, err := u.CreateIpAllowlist(ctx, caller, spaceId, clientId, domainusecasesmanagement.CreateClientIpAllowlistRequest{Cidr: originCidr})
	if err != nil {
		t.Fatal(err)
	}
	assertAllowed := func(ip string, expected bool) {
		t.Helper()
		scoped, err := allowed.IsAllowedIpForClient(ctx, spaceId, clientId, ip)
		if err != nil || scoped != expected {
			t.Fatalf("incorrect client allowance for %s: %v %v", ip, scoped, err)
		}
	}
	assertAllowed(originIp, true)
	assertAllowed(outsideIp, false)

	item, err := u.ReadIpAllowlistById(ctx, caller, spaceId, clientId, id)
	if err != nil || item.ClientId != clientId || item.AuditCreate.By == nil || *item.AuditCreate.By != userId {
		t.Fatalf("incorrect created entry: %+v %v", item, err)
	}
	byCidr, err := u.ReadIpAllowlistByCidr(ctx, caller, spaceId, clientId, originCidr)
	if err != nil || byCidr.Id != id {
		t.Fatalf("CIDR lookup failed: %v", err)
	}
	items, err := u.ReadIpAllowlistsByClientId(ctx, caller, spaceId, clientId)
	if err != nil || len(items) != 1 {
		t.Fatalf("allowlist list failed: %v", err)
	}
	page, err := u.ReadIpAllowlistsByFilter(ctx, caller, spaceId, clientId, domainusecasesmanagement.FilterClientIpAllowlistRequest{Page: 1, Limit: 10})
	if err != nil || len(page.Items) != 1 || page.Total != 1 {
		t.Fatalf("allowlist page failed: %v", err)
	}
	_, err = u.ReadIpAllowlistById(ctx, caller, spaceId, otherClientId, id)
	testshared.RequireError(t, err, domainmodelserror.ClientIpAllowlistNotFound)
	err = u.UpdateIpAllowlistById(ctx, caller, spaceId, otherClientId, id, domainusecasesmanagement.UpdateClientIpAllowlistRequest{IsActive: &no})
	testshared.RequireError(t, err, domainmodelserror.ClientIpAllowlistNotFound)
	err = u.DeleteIpAllowlistById(ctx, caller, spaceId, otherClientId, id)
	testshared.RequireError(t, err, domainmodelserror.ClientIpAllowlistNotFound)
	assertAllowed(originIp, true)

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
	created, err := failing.CreateIpAllowlist(ctx, caller, spaceId, clientId, domainusecasesmanagement.CreateClientIpAllowlistRequest{Cidr: rollbackCidr})
	if created != 0 || !errors.Is(err, rollback) {
		t.Fatalf("rollback returned a creation result: %d %v", created, err)
	}
	assertAllowed(rollbackIp, false)
	err = failing.UpdateIpAllowlistById(ctx, caller, spaceId, clientId, id, domainusecasesmanagement.UpdateClientIpAllowlistRequest{IsActive: &no})
	testshared.RequireError(t, err, rollback)
	assertAllowed(originIp, true)
	err = failing.DeleteIpAllowlistById(ctx, caller, spaceId, clientId, id)
	testshared.RequireError(t, err, rollback)
	assertAllowed(originIp, true)

	err = u.UpdateIpAllowlistById(ctx, caller, spaceId, clientId, id, domainusecasesmanagement.UpdateClientIpAllowlistRequest{IsActive: &no})
	if err != nil {
		t.Fatal(err)
	}
	assertAllowed(originIp, false)
	err = u.UpdateIpAllowlistById(ctx, caller, spaceId, clientId, id, domainusecasesmanagement.UpdateClientIpAllowlistRequest{Cidr: &replacementCidr, IsActive: &yes})
	if err != nil {
		t.Fatal(err)
	}
	assertAllowed(originIp, false)
	assertAllowed(replacementIp, true)

	err = f.Client.UpdateById(ctx, spaceId, clientId, domaincontractsrepository.UpdateClientArgs{IsActive: &no})
	if err != nil {
		t.Fatal(err)
	}
	assertAllowed(replacementIp, false)
	_, err = u.ReadIpAllowlistById(ctx, caller, spaceId, clientId, id)
	if err != nil {
		t.Fatal("inactive client became unmanageable", err)
	}
	err = f.Client.UpdateById(ctx, spaceId, clientId, domaincontractsrepository.UpdateClientArgs{IsActive: &yes})
	if err != nil {
		t.Fatal(err)
	}
	assertAllowed(replacementIp, true)

	err = u.DeleteIpAllowlistById(ctx, caller, spaceId, clientId, id)
	if err != nil {
		t.Fatal(err)
	}
	assertAllowed(replacementIp, false)
	_, err = u.ReadIpAllowlistById(ctx, caller, spaceId, clientId, id)
	testshared.RequireError(t, err, domainmodelserror.ClientIpAllowlistNotFound)
	page, err = u.ReadIpAllowlistsByFilter(ctx, caller, spaceId, clientId, domainusecasesmanagement.FilterClientIpAllowlistRequest{Limit: 10})
	if err != nil || page.Items == nil || len(page.Items) != 0 || page.Total != 0 {
		t.Fatalf("deleted entry remained in management list: %+v %v", page, err)
	}

	// Existing client permissions do not grant allowlist administration.
	caller.Permissions = []string{"client:add", "client:get", "client:set", "client:remove"}
	_, err = u.CreateIpAllowlist(ctx, caller, spaceId, clientId, domainusecasesmanagement.CreateClientIpAllowlistRequest{Cidr: originCidr})
	testshared.RequireError(t, err, domainmodelserror.Forbidden)
}
