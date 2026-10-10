package repository_test

import (
	"context"
	"errors"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"sync"
	"testing"
	"time"
)

func TestDefaultRoleAndMembershipAtomicity(t *testing.T) {
	f := newFixture(t)
	if _, err := f.SpaceMember.Create(f.ctx, f.otherSpaceId, domaincontractsrepository.CreateSpaceMemberArgs{UserId: f.userId}); !errors.Is(err, domainmodelserror.DefaultRoleNotFound) {
		t.Fatalf("missing default accepted: %v", err)
	}
	var count int
	require(t, f.pool.QueryRow(f.ctx, "SELECT count(*) FROM space_members WHERE space_id=$1", f.otherSpaceId).Scan(&count))
	if count != 0 {
		t.Fatal("membership retained after missing-default failure")
	}
	if err := f.Role.UpdateById(f.ctx, f.spaceId, 999, domaincontractsrepository.UpdateRoleArgs{IsDefault: ptr(true)}); !errors.Is(err, domainmodelserror.RoleNotFound) {
		t.Fatal("missing default target accepted")
	}
	current, err := f.Role.ReadDefault(f.ctx, f.spaceId)
	require(t, err)
	if current.Id != f.roleId {
		t.Fatal("failed default update changed the existing default")
	}
	replacement, err := f.Role.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateRoleArgs{Slug: "replacement", Name: "Replacement", IsDefault: ptr(true)})
	require(t, err)
	member, err := f.SpaceMember.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateSpaceMemberArgs{UserId: f.userId})
	require(t, err)
	roles, err := f.SpaceMemberRole.ReadBySpaceMemberId(f.ctx, f.spaceId, member)
	require(t, err)
	if len(roles) != 1 || roles[0].Role.Id != replacement {
		t.Fatal("membership did not receive current default")
	}
	require(t, f.Role.DeleteById(f.ctx, f.spaceId, replacement, nil))
	if _, err := f.Role.ReadDefault(f.ctx, f.spaceId); !errors.Is(err, domainmodelserror.DefaultRoleNotFound) {
		t.Fatal("deleted role remained default")
	}
}

func TestConcurrentDefaultChangesAndMembershipCreation(t *testing.T) {
	f := newFixture(t)
	ctx, cancel := context.WithTimeout(f.ctx, 10*time.Second)
	defer cancel()
	second, err := f.Role.Create(ctx, f.spaceId, domaincontractsrepository.CreateRoleArgs{Slug: "second-role", Name: "Second"})
	require(t, err)
	users := make([]int64, 6)
	for i := range users {
		users[i], err = f.User.Create(ctx, domaincontractsrepository.CreateUserArgs{Name: "Concurrent User"})
		require(t, err)
	}
	var group sync.WaitGroup
	failures := make(chan error, 18)
	for _, userId := range users {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := f.SpaceMember.Create(ctx, f.spaceId, domaincontractsrepository.CreateSpaceMemberArgs{UserId: userId})
			failures <- err
		}()
	}
	for i := range 12 {
		roleId := f.roleId
		if i%2 == 0 {
			roleId = second
		}
		group.Add(1)
		go func() {
			defer group.Done()
			failures <- f.Role.UpdateById(ctx, f.spaceId, roleId, domaincontractsrepository.UpdateRoleArgs{IsDefault: ptr(true)})
		}()
	}
	group.Wait()
	close(failures)
	for err := range failures {
		require(t, err)
	}
	var defaults, badAssignments int
	require(t, f.pool.QueryRow(ctx, "SELECT count(*) FROM roles WHERE space_id=$1 AND is_default AND deleted_at IS NULL", f.spaceId).Scan(&defaults))
	require(t, f.pool.QueryRow(ctx, "SELECT count(*) FROM space_members m WHERE m.space_id=$1 AND (SELECT count(*) FROM space_member_role mr WHERE mr.space_member_id=m.id)!=1", f.spaceId).Scan(&badAssignments))
	if defaults != 1 || badAssignments != 0 {
		t.Fatalf("concurrent invariants failed: defaults=%d bad memberships=%d", defaults, badAssignments)
	}
}

func TestConcurrentUserDeletionAndDependentCreation(t *testing.T) {
	for _, kind := range []string{"classic", "google", "membership"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			ctx, cancel := context.WithTimeout(f.ctx, 5*time.Second)
			defer cancel()
			completed := make(chan error, 1)
			require(t, f.dt.WithTx(ctx, func(txCtx context.Context) error {
				if err := f.User.DeleteById(txCtx, f.userId, nil); err != nil {
					return err
				}
				started := make(chan struct{})
				go func() {
					close(started)
					var err error
					switch kind {
					case "classic":
						_, err = f.ClassicAuth.Create(ctx, domaincontractsrepository.CreateClassicAuthArgs{UserId: f.userId, Username: "concurrent", PasswordHash: "hash"})
					case "google":
						_, err = f.GoogleAuth.Create(ctx, domaincontractsrepository.CreateGoogleAuthArgs{UserId: f.userId, Subject: "concurrent-subject", Email: "concurrent@example.test"})
					case "membership":
						_, err = f.SpaceMember.Create(ctx, f.spaceId, domaincontractsrepository.CreateSpaceMemberArgs{UserId: f.userId})
					}
					completed <- err
				}()
				<-started
				return nil
			}))
			if err := <-completed; !errors.Is(err, domainmodelserror.UserNotFound) {
				t.Fatalf("dependent creation survived deletion: %v", err)
			}
			var count int
			require(t, f.pool.QueryRow(ctx, "SELECT (SELECT count(*) FROM classic_auths WHERE user_id=$1 AND deleted_at IS NULL)+(SELECT count(*) FROM google_auths WHERE user_id=$1 AND deleted_at IS NULL)+(SELECT count(*) FROM space_members WHERE user_id=$1 AND deleted_at IS NULL)", f.userId).Scan(&count))
			if count != 0 {
				t.Fatal("live dependent attached to deleted user")
			}
		})
	}
}

func TestClientSecretsAndDeletedParent(t *testing.T) {
	f := newFixture(t)
	id, err := f.Client.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientArgs{Slug: "web", Name: "Web", SecretKeyHash: "first-hash", SecretKeyRedacted: "abcd****wxyz"})
	require(t, err)
	_, err = f.Client.ReadBySlug(f.ctx, "web")
	require(t, err)
	if err := f.Client.UpdateById(f.ctx, f.spaceId, id, domaincontractsrepository.UpdateClientArgs{SecretKeyHash: ptr("second-hash")}); !errors.Is(err, domainmodelserror.BadArgs) {
		t.Fatal("partial secret rotation accepted")
	}
	require(t, f.Client.UpdateById(f.ctx, f.spaceId, id, domaincontractsrepository.UpdateClientArgs{SecretKeyHash: ptr("second-hash"), SecretKeyRedacted: ptr("efgh****mnop"), IsActive: ptr(false)}))
	client, err := f.Client.ReadById(f.ctx, f.spaceId, id)
	require(t, err)
	if client.SecretKeyHash != "second-hash" || client.IsActive {
		t.Fatal("secret rotation or inactive management failed")
	}
	require(t, f.Space.DeleteById(f.ctx, f.spaceId, nil))
	if _, err := f.Client.ReadBySlug(f.ctx, "web"); !errors.Is(err, domainmodelserror.ClientNotFound) {
		t.Fatal("deleted parent still exposed client")
	}
	if _, err := f.Client.Create(f.ctx, f.otherSpaceId, domaincontractsrepository.CreateClientArgs{Slug: "web", Name: "Other", SecretKeyHash: "hash", SecretKeyRedacted: "abcd****wxyz"}); !errors.Is(err, domainmodelserror.ClientSlugConflict) {
		t.Fatal("deleted parent released client slug")
	}
}
