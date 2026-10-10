package repository_test

import (
	"errors"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"testing"
)

func TestRolePermissionRepository(t *testing.T) {
	f := newFixture(t)
	permissionId, err := f.Permission.Create(f.ctx, f.spaceId, domaincontractsrepository.CreatePermissionArgs{Slug: "user:get", Name: "Read Users"})
	require(t, err)
	id, err := f.RolePermission.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateRolePermissionArgs{RoleId: f.roleId, PermissionId: permissionId, By: ptr(int64(7))})
	require(t, err)
	item, err := f.RolePermission.ReadById(f.ctx, f.spaceId, id)
	require(t, err)
	if item.Role.Id != f.roleId || item.Permission.Id != permissionId || item.RolePermission.AuditCreate.By == nil {
		t.Fatal("incorrect assignment join")
	}
	pair, err := f.RolePermission.ReadByRoleIdAndPermissionId(f.ctx, f.spaceId, f.roleId, permissionId)
	require(t, err)
	if pair.RolePermission.Id != id {
		t.Fatal("pair lookup failed")
	}
	byRole, err := f.RolePermission.ReadByRoleId(f.ctx, f.spaceId, f.roleId)
	require(t, err)
	byPermission, err := f.RolePermission.ReadByPermissionId(f.ctx, f.spaceId, permissionId)
	require(t, err)
	if len(byRole) != 1 || len(byPermission) != 1 {
		t.Fatal("assignment lists incorrect")
	}
	if _, err := f.RolePermission.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateRolePermissionArgs{RoleId: f.roleId, PermissionId: permissionId}); !errors.Is(err, domainmodelserror.RolePermissionConflict) {
		t.Fatal("duplicate assignment accepted")
	}
	other, err := f.Permission.Create(f.ctx, f.otherSpaceId, domaincontractsrepository.CreatePermissionArgs{Slug: "user:get", Name: "Other"})
	require(t, err)
	if _, err := f.RolePermission.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateRolePermissionArgs{RoleId: f.roleId, PermissionId: other}); !errors.Is(err, domainmodelserror.PermissionNotFound) {
		t.Fatalf("cross-space assignment accepted: %v", err)
	}
	if err := f.RolePermission.DeleteById(f.ctx, f.otherSpaceId, id); !errors.Is(err, domainmodelserror.RolePermissionNotFound) {
		t.Fatal("cross-space deletion accepted")
	}
	if err := f.RolePermission.DeleteByRoleIdOrPermissionId(f.ctx, f.spaceId, domaincontractsrepository.DeleteRolePermissionArgs{}); !errors.Is(err, domainmodelserror.BadArgs) {
		t.Fatal("unbounded delete accepted")
	}
	require(t, f.RolePermission.DeleteByRoleIdOrPermissionId(f.ctx, f.spaceId, domaincontractsrepository.DeleteRolePermissionArgs{RoleId: ptr(int64(999)), PermissionId: &permissionId}))
	if _, err := f.RolePermission.ReadById(f.ctx, f.spaceId, id); !errors.Is(err, domainmodelserror.RolePermissionNotFound) {
		t.Fatal("OR selector failed")
	}
	id, err = f.RolePermission.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateRolePermissionArgs{RoleId: f.roleId, PermissionId: permissionId})
	require(t, err)
	require(t, f.RolePermission.DeleteById(f.ctx, f.spaceId, id))
	empty, err := f.RolePermission.ReadByRoleId(f.ctx, f.spaceId, f.roleId)
	require(t, err)
	if empty == nil || len(empty) != 0 {
		t.Fatal("empty assignment list must be allocated")
	}
	require(t, f.RolePermission.DeleteByRoleIdOrPermissionId(f.ctx, f.spaceId, domaincontractsrepository.DeleteRolePermissionArgs{RoleId: &f.roleId}))
}

func TestSpaceMemberRoleRepository(t *testing.T) {
	f := newFixture(t)
	memberId, err := f.SpaceMember.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateSpaceMemberArgs{UserId: f.userId})
	require(t, err)
	secondRole, err := f.Role.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateRoleArgs{Slug: "admin", Name: "Admin"})
	require(t, err)
	id, err := f.SpaceMemberRole.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateSpaceMemberRoleArgs{SpaceMemberId: memberId, RoleId: secondRole})
	require(t, err)
	item, err := f.SpaceMemberRole.ReadById(f.ctx, f.spaceId, id)
	require(t, err)
	if item.SpaceMember.Id != memberId || item.Role.Id != secondRole {
		t.Fatal("incorrect assignment join")
	}
	pair, err := f.SpaceMemberRole.ReadBySpaceMemberIdAndRoleId(f.ctx, f.spaceId, memberId, secondRole)
	require(t, err)
	if pair.SpaceMemberRole.Id != id {
		t.Fatal("pair lookup failed")
	}
	byMember, err := f.SpaceMemberRole.ReadBySpaceMemberId(f.ctx, f.spaceId, memberId)
	require(t, err)
	byRole, err := f.SpaceMemberRole.ReadByRoleId(f.ctx, f.spaceId, secondRole)
	require(t, err)
	if len(byMember) != 2 || len(byRole) != 1 {
		t.Fatal("assignment lists incorrect")
	}
	if _, err := f.SpaceMemberRole.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateSpaceMemberRoleArgs{SpaceMemberId: memberId, RoleId: secondRole}); !errors.Is(err, domainmodelserror.SpaceMemberRoleConflict) {
		t.Fatal("duplicate assignment accepted")
	}
	other, err := f.Role.Create(f.ctx, f.otherSpaceId, domaincontractsrepository.CreateRoleArgs{Slug: "other", Name: "Other"})
	require(t, err)
	if _, err := f.SpaceMemberRole.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateSpaceMemberRoleArgs{SpaceMemberId: memberId, RoleId: other}); !errors.Is(err, domainmodelserror.RoleNotFound) {
		t.Fatal("cross-space role accepted")
	}
	permissionId, err := f.Permission.Create(f.ctx, f.spaceId, domaincontractsrepository.CreatePermissionArgs{Slug: "profile:get", Name: "Read Profile"})
	require(t, err)
	for _, roleId := range []int64{f.roleId, secondRole} {
		_, err = f.RolePermission.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateRolePermissionArgs{RoleId: roleId, PermissionId: permissionId})
		require(t, err)
	}
	grants, err := f.SpaceMemberRole.ReadEffectivePermissionsBySpaceMemberId(f.ctx, f.spaceId, memberId)
	require(t, err)
	roles, err := f.SpaceMemberRole.ReadEffectiveRolesBySpaceMemberId(f.ctx, f.spaceId, memberId)
	require(t, err)
	if len(grants) != 1 || len(roles) != 2 || grants[0].Id != permissionId {
		t.Fatal("effective grants not deduplicated")
	}
	require(t, f.SpaceMember.UpdateById(f.ctx, f.spaceId, memberId, domaincontractsrepository.UpdateSpaceMemberArgs{IsActive: ptr(false)}))
	if _, err := f.SpaceMemberRole.ReadEffectivePermissionsBySpaceMemberId(f.ctx, f.spaceId, memberId); !errors.Is(err, domainmodelserror.Forbidden) {
		t.Fatal("inactive membership granted access")
	}
	require(t, f.SpaceMember.UpdateById(f.ctx, f.spaceId, memberId, domaincontractsrepository.UpdateSpaceMemberArgs{IsActive: ptr(true)}))
	require(t, f.Permission.DeleteById(f.ctx, f.spaceId, permissionId, nil))
	grants, err = f.SpaceMemberRole.ReadEffectivePermissionsBySpaceMemberId(f.ctx, f.spaceId, memberId)
	require(t, err)
	if grants == nil || len(grants) != 0 {
		t.Fatal("deleted permission granted access")
	}
	if err := f.SpaceMemberRole.DeleteBySpaceMemberIdOrRoleId(f.ctx, f.spaceId, domaincontractsrepository.DeleteSpaceMemberRoleArgs{}); !errors.Is(err, domainmodelserror.BadArgs) {
		t.Fatal("unbounded delete accepted")
	}
	require(t, f.SpaceMemberRole.DeleteById(f.ctx, f.spaceId, id))
	require(t, f.SpaceMemberRole.DeleteBySpaceMemberIdOrRoleId(f.ctx, f.spaceId, domaincontractsrepository.DeleteSpaceMemberRoleArgs{SpaceMemberId: &memberId, RoleId: ptr(int64(999))}))
	empty, err := f.SpaceMemberRole.ReadBySpaceMemberId(f.ctx, f.spaceId, memberId)
	require(t, err)
	if empty == nil || len(empty) != 0 {
		t.Fatal("OR bulk deletion failed")
	}
}
