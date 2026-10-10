package repository_test

import (
	"encoding/json"
	"errors"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"testing"
)

func TestScopedLookupsAndFilters(t *testing.T) {
	f := newFixture(t)
	_, err := f.Space.ReadBySlug(f.ctx, "first")
	require(t, err)
	_, err = f.Role.ReadBySlug(f.ctx, f.spaceId, "default")
	require(t, err)
	permissionId, err := f.Permission.Create(f.ctx, f.spaceId, domaincontractsrepository.CreatePermissionArgs{Slug: "user:manage", Name: "Manage Users", Description: ptr("Identity management")})
	require(t, err)
	found, err := f.Permission.ReadBySlug(f.ctx, f.spaceId, "user:manage")
	require(t, err)
	if found.Id != permissionId {
		t.Fatal("slug lookup returned wrong permission")
	}
	if _, err := f.Permission.ReadBySlug(f.ctx, f.otherSpaceId, "user:manage"); !errors.Is(err, domainmodelserror.PermissionNotFound) {
		t.Fatal("cross-space slug lookup accepted")
	}
	permissions, total, err := f.Permission.ReadByFilter(f.ctx, f.spaceId, domaincontractsrepository.PermissionFilter{Page: 1, Limit: 10, Search: ptr("Identity")})
	require(t, err)
	if len(permissions) != 1 || total != 1 {
		t.Fatal("permission search missed description")
	}
	roles, total, err := f.Role.ReadByFilter(f.ctx, f.spaceId, domaincontractsrepository.RoleFilter{Page: 1, Limit: 10, IsDefault: ptr(true)})
	require(t, err)
	if len(roles) != 1 || total != 1 {
		t.Fatal("default filter failed")
	}
	_, err = f.ClassicAuth.Create(f.ctx, domaincontractsrepository.CreateClassicAuthArgs{UserId: f.userId, Username: "filter-user", PasswordHash: "PRIVATE_HASH", Email: ptr("filter@example.test"), IsEmailVerified: ptr(true)})
	require(t, err)
	_, err = f.GoogleAuth.Create(f.ctx, domaincontractsrepository.CreateGoogleAuthArgs{UserId: f.userId, Subject: "filter-subject", Email: "filter@example.test", IsEmailVerified: ptr(true)})
	require(t, err)
	memberId, err := f.SpaceMember.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateSpaceMemberArgs{UserId: f.userId})
	require(t, err)
	member, err := f.SpaceMember.ReadByUserId(f.ctx, f.spaceId, f.userId)
	require(t, err)
	if member.Id != memberId {
		t.Fatal("membership user lookup failed")
	}
	members, total, err := f.SpaceMember.ReadByFilter(f.ctx, f.spaceId, domaincontractsrepository.SpaceMemberFilter{Page: 1, Limit: 10, UserId: &f.userId, RoleId: &f.roleId, Search: ptr("filter@example"), IsActive: ptr(true)})
	require(t, err)
	if len(members) != 1 || total != 1 || members[0].User.Id != f.userId {
		t.Fatal("joined contact search duplicated or lost membership")
	}
	auths, total, err := f.ClassicAuth.ReadByFilter(f.ctx, domaincontractsrepository.ClassicAuthFilter{Page: 1, Limit: 10, UserId: &f.userId, IsEmailVerified: ptr(true), IsPhoneVerified: ptr(false), Search: ptr("filter-user")})
	require(t, err)
	if len(auths) != 1 || total != 1 {
		t.Fatal("credential filters failed")
	}
	google, total, err := f.GoogleAuth.ReadByFilter(f.ctx, domaincontractsrepository.GoogleAuthFilter{Page: 1, Limit: 10, UserId: &f.userId, IsEmailVerified: ptr(true), Search: ptr("filter@example")})
	require(t, err)
	if len(google) != 1 || total != 1 {
		t.Fatal("Google filters failed")
	}
	require(t, f.User.UpdateById(f.ctx, f.userId, domaincontractsrepository.UpdateUserArgs{AvatarPath: ptr("avatars/user.png")}))
	user, err := f.User.ReadById(f.ctx, f.userId)
	require(t, err)
	if user.AvatarPath == nil {
		t.Fatal("avatar update failed")
	}
	require(t, f.User.UpdateById(f.ctx, f.userId, domaincontractsrepository.UpdateUserArgs{AvatarPathSetNull: true}))
	user, err = f.User.ReadById(f.ctx, f.userId)
	require(t, err)
	if user.AvatarPath != nil {
		t.Fatal("avatar clear failed")
	}
	require(t, f.Space.UpdateById(f.ctx, f.spaceId, domaincontractsrepository.UpdateSpaceArgs{IsActive: ptr(false)}))
	spaces, total, err := f.Space.ReadByFilter(f.ctx, domaincontractsrepository.SpaceFilter{Page: 1, Limit: 10, IsActive: ptr(false)})
	require(t, err)
	if len(spaces) != 1 || total != 1 {
		t.Fatal("inactive-space filter failed")
	}
	if _, err := f.SpaceMemberRole.ReadEffectiveRolesBySpaceMemberId(f.ctx, f.spaceId, memberId); !errors.Is(err, domainmodelserror.Forbidden) {
		t.Fatal("inactive space granted access")
	}
}

func TestAtomicDeletionAudit(t *testing.T) {
	f := newFixture(t)
	_, err := f.ClassicAuth.Create(f.ctx, domaincontractsrepository.CreateClassicAuthArgs{UserId: f.userId, Username: "audit-user", PasswordHash: "hash"})
	require(t, err)
	_, err = f.GoogleAuth.Create(f.ctx, domaincontractsrepository.CreateGoogleAuthArgs{UserId: f.userId, Subject: "audit-subject", Email: "audit@example.test"})
	require(t, err)
	_, err = f.SpaceMember.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateSpaceMemberArgs{UserId: f.userId})
	require(t, err)
	require(t, f.User.DeleteById(f.ctx, f.userId, ptr(int64(9))))
	var matches bool
	require(t, f.pool.QueryRow(f.ctx, "SELECT u.deleted_at=c.deleted_at AND u.deleted_at=g.deleted_at AND u.deleted_at=m.deleted_at AND u.deleted_by=c.deleted_by AND u.deleted_by=g.deleted_by AND u.deleted_by=m.deleted_by FROM users u JOIN classic_auths c ON c.user_id=u.id JOIN google_auths g ON g.user_id=u.id JOIN space_members m ON m.user_id=u.id WHERE u.id=$1", f.userId).Scan(&matches))
	if !matches {
		t.Fatal("dependent deletion audit differs")
	}
}

func TestZeroLimitAndJSONReplacement(t *testing.T) {
	f := newFixture(t)
	items, total, err := f.User.ReadByFilter(f.ctx, domaincontractsrepository.UserFilter{Page: -2, Limit: -1})
	require(t, err)
	if items == nil || len(items) != 0 || total != 1 {
		t.Fatal("nonpositive pagination behavior changed")
	}
	require(t, f.User.UpdateById(f.ctx, f.userId, domaincontractsrepository.UpdateUserArgs{Preferences: ptr(json.RawMessage(`{}`))}))
	user, err := f.User.ReadById(f.ctx, f.userId)
	require(t, err)
	if string(user.Preferences) != "{}" {
		t.Fatalf("preference replacement failed: %s", user.Preferences)
	}
}
