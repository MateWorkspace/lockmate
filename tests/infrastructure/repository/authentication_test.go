package repository_test

import (
	"context"
	"errors"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"github.com/jackc/pgx/v5/pgconn"
	"testing"
)

func TestAuthenticationContactsAndConflicts(t *testing.T) {
	f := newFixture(t)
	id, err := f.ClassicAuth.Create(f.ctx, domaincontractsrepository.CreateClassicAuthArgs{UserId: f.userId, Username: "username", PasswordHash: "hash", Email: ptr("one@example.test"), Phone: ptr("1234567890"), IsEmailVerified: ptr(true), IsPhoneVerified: ptr(true)})
	require(t, err)
	for _, read := range []func() error{
		func() error { _, err := f.ClassicAuth.ReadByUserId(f.ctx, f.userId); return err },
		func() error { _, err := f.ClassicAuth.ReadByUsername(f.ctx, "username"); return err },
		func() error { _, err := f.ClassicAuth.ReadByEmail(f.ctx, "one@example.test"); return err },
		func() error { _, err := f.ClassicAuth.ReadByPhone(f.ctx, "1234567890"); return err },
	} {
		require(t, read())
	}
	require(t, f.ClassicAuth.UpdateById(f.ctx, id, domaincontractsrepository.UpdateClassicAuthArgs{Email: ptr("one@example.test")}))
	item, err := f.ClassicAuth.ReadById(f.ctx, id)
	require(t, err)
	if !item.IsEmailVerified {
		t.Fatal("unchanged contact reset verification")
	}
	require(t, f.ClassicAuth.UpdateById(f.ctx, id, domaincontractsrepository.UpdateClassicAuthArgs{Email: ptr("two@example.test"), PhoneSetNull: true}))
	item, err = f.ClassicAuth.ReadById(f.ctx, id)
	require(t, err)
	if item.IsEmailVerified || item.IsPhoneVerified || item.Phone != nil {
		t.Fatal("contact reset failed")
	}
	require(t, f.ClassicAuth.UpdateById(f.ctx, id, domaincontractsrepository.UpdateClassicAuthArgs{Email: ptr("three@example.test"), IsEmailVerified: ptr(true)}))
	if err := f.ClassicAuth.UpdateById(f.ctx, id, domaincontractsrepository.UpdateClassicAuthArgs{Email: ptr("bad"), EmailSetNull: true}); !errors.Is(err, domainmodelserror.BadArgs) {
		t.Fatal("contradictory clear accepted")
	}
	secondUser, err := f.User.Create(f.ctx, domaincontractsrepository.CreateUserArgs{Name: "Second User"})
	require(t, err)
	_, err = f.ClassicAuth.Create(f.ctx, domaincontractsrepository.CreateClassicAuthArgs{UserId: secondUser, Username: "username", PasswordHash: "hash"})
	var pgError *pgconn.PgError
	if !errors.Is(err, domainmodelserror.ClassicAuthUsernameConflict) || !errors.As(err, &pgError) || pgError.ConstraintName != "uq_classic_auths_username" {
		t.Fatalf("lost conflict cause: %v", err)
	}
	_, err = f.ClassicAuth.Create(f.ctx, domaincontractsrepository.CreateClassicAuthArgs{UserId: secondUser, Username: "other", PasswordHash: "hash", Email: ptr("three@example.test")})
	if !errors.Is(err, domainmodelserror.ClassicAuthEmailConflict) {
		t.Fatalf("wrong email conflict: %v", err)
	}
	_, err = f.ClassicAuth.Create(f.ctx, domaincontractsrepository.CreateClassicAuthArgs{UserId: f.userId, Username: "other", PasswordHash: "hash"})
	if !errors.Is(err, domainmodelserror.ClassicAuthUserConflict) {
		t.Fatalf("wrong user conflict: %v", err)
	}
	googleId, err := f.GoogleAuth.Create(f.ctx, domaincontractsrepository.CreateGoogleAuthArgs{UserId: f.userId, Subject: "subject", Email: "google@example.test", IsEmailVerified: ptr(true)})
	require(t, err)
	_, err = f.GoogleAuth.ReadByUserId(f.ctx, f.userId)
	require(t, err)
	_, err = f.GoogleAuth.ReadBySubject(f.ctx, "subject")
	require(t, err)
	require(t, f.GoogleAuth.UpdateById(f.ctx, googleId, domaincontractsrepository.UpdateGoogleAuthArgs{Email: ptr("changed@example.test")}))
	google, err := f.GoogleAuth.ReadById(f.ctx, googleId)
	require(t, err)
	if google.IsEmailVerified {
		t.Fatal("changed Google email retained verification")
	}
	_, err = f.GoogleAuth.Create(f.ctx, domaincontractsrepository.CreateGoogleAuthArgs{UserId: secondUser, Subject: "subject", Email: "changed@example.test"})
	if !errors.Is(err, domainmodelserror.GoogleAuthSubjectConflict) {
		t.Fatal("duplicate subject accepted")
	}
	_, err = f.GoogleAuth.Create(f.ctx, domaincontractsrepository.CreateGoogleAuthArgs{UserId: secondUser, Subject: "other-subject", Email: "changed@example.test"})
	require(t, err)
}

func TestUserDeletionAtomicity(t *testing.T) {
	f := newFixture(t)
	classicId, err := f.ClassicAuth.Create(f.ctx, domaincontractsrepository.CreateClassicAuthArgs{UserId: f.userId, Username: "delete-user", PasswordHash: "hash"})
	require(t, err)
	googleId, err := f.GoogleAuth.Create(f.ctx, domaincontractsrepository.CreateGoogleAuthArgs{UserId: f.userId, Subject: "delete-subject", Email: "delete@example.test"})
	require(t, err)
	memberId, err := f.SpaceMember.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateSpaceMemberArgs{UserId: f.userId})
	require(t, err)
	callbackError := errors.New("rollback deletion")
	err = f.dt.WithTx(f.ctx, func(ctx context.Context) error {
		if err := f.User.DeleteById(ctx, f.userId, ptr(int64(7))); err != nil {
			return err
		}
		return callbackError
	})
	if err != callbackError {
		t.Fatalf("transaction error changed: %v", err)
	}
	_, err = f.User.ReadById(f.ctx, f.userId)
	require(t, err)
	_, err = f.ClassicAuth.ReadById(f.ctx, classicId)
	require(t, err)
	_, err = f.GoogleAuth.ReadById(f.ctx, googleId)
	require(t, err)
	_, err = f.SpaceMember.ReadById(f.ctx, f.spaceId, memberId)
	require(t, err)
	require(t, f.User.DeleteById(f.ctx, f.userId, ptr(int64(9))))
	var authDeleted, googleDeleted, memberDeleted bool
	err = f.pool.QueryRow(f.ctx, "SELECT (SELECT deleted_at IS NOT NULL FROM classic_auths WHERE id=$1),(SELECT deleted_at IS NOT NULL FROM google_auths WHERE id=$2),(SELECT deleted_at IS NOT NULL FROM space_members WHERE id=$3)", classicId, googleId, memberId).Scan(&authDeleted, &googleDeleted, &memberDeleted)
	require(t, err)
	if !authDeleted || !googleDeleted || !memberDeleted {
		t.Fatal("dependent records not deleted")
	}
	var count int
	require(t, f.pool.QueryRow(f.ctx, "SELECT count(*) FROM space_member_role WHERE space_member_id=$1", memberId).Scan(&count))
	if count != 0 {
		t.Fatal("user membership assignments survived deletion")
	}
	replacement, err := f.User.Create(f.ctx, domaincontractsrepository.CreateUserArgs{Name: "Replacement"})
	require(t, err)
	_, err = f.ClassicAuth.Create(f.ctx, domaincontractsrepository.CreateClassicAuthArgs{UserId: replacement, Username: "delete-user", PasswordHash: "hash"})
	require(t, err)
	_, err = f.GoogleAuth.Create(f.ctx, domaincontractsrepository.CreateGoogleAuthArgs{UserId: replacement, Subject: "delete-subject", Email: "delete@example.test"})
	require(t, err)
}
