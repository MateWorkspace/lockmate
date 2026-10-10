package applicationsharedtest

import (
	"context"
	"errors"
	"testing"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructureutilityvalidator "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/validator"
)

type transactionKey struct{}

type Fixture struct {
	User            *User
	Space           *Space
	SpaceMember     *SpaceMember
	Client          *Client
	ClassicAuth     *ClassicAuth
	GoogleAuth      *GoogleAuth
	SpaceMemberRole *SpaceMemberRole
	Role            *Role
	Permission      *Permission
	RolePermission  *RolePermission
	Guard           *PermissionGuard
	Validator       domaincontractsutility.Validator
	Password        *Password
	Token           *Token
	Keygen          *Keygen
	Transactor      *Transactor
	Logger          *Logger
	Users           map[int64]*domainmodelsentity.User
	Members         map[int64]*domainmodelsentity.SpaceMember
	Classics        map[int64]*domainmodelsentity.ClassicAuth
	Googles         map[int64]*domainmodelsentity.GoogleAuth
	Caller          *domainmodelsentity.AccessClaims
	CommitErr       error
	BeginErr        error
	TokenErr        error
	MissingDefault  bool
	Transactions    int
	Committed       bool
}

func NewFixture() *Fixture {
	f := &Fixture{
		User: &User{}, Space: &Space{}, SpaceMember: &SpaceMember{}, Client: &Client{},
		ClassicAuth: &ClassicAuth{}, GoogleAuth: &GoogleAuth{}, SpaceMemberRole: &SpaceMemberRole{},
		Role: &Role{}, Permission: &Permission{}, RolePermission: &RolePermission{},
		Guard: &PermissionGuard{}, Password: &Password{}, Token: &Token{}, Keygen: &Keygen{}, Transactor: &Transactor{}, Logger: &Logger{},
		Validator: infrastructureutilityvalidator.NewValidatorImpl(),
		Users:     map[int64]*domainmodelsentity.User{1: {Id: 1, Name: "Current name"}},
		Members:   map[int64]*domainmodelsentity.SpaceMember{20: {Id: 20, SpaceId: 10, UserId: 1, IsActive: true}},
		Classics:  map[int64]*domainmodelsentity.ClassicAuth{1: {Id: 40, UserId: 1, Username: "user", PasswordHash: "hash:password"}},
		Googles:   map[int64]*domainmodelsentity.GoogleAuth{},
		Caller:    &domainmodelsentity.AccessClaims{UserId: 1, SpaceId: 10, SpaceMemberId: 20, Permissions: []string{"profile:get"}},
	}
	f.User.ReadByIdFn = func(ctx context.Context, id int64) (*domainmodelsentity.User, error) {
		if v := f.Users[id]; v != nil {
			return v, nil
		}
		return nil, NotFound(domainmodelserror.UserNotFound)
	}
	f.User.CreateFn = func(ctx context.Context, args domaincontractsrepository.CreateUserArgs) (int64, error) {
		id := int64(len(f.Users) + 1)
		f.Users[id] = &domainmodelsentity.User{Id: id, Name: args.Name}
		return id, nil
	}
	f.User.UpdateByIdFn = func(ctx context.Context, id int64, args domaincontractsrepository.UpdateUserArgs) error {
		current := f.Users[id]
		if current == nil {
			return NotFound(domainmodelserror.UserNotFound)
		}
		value := *current
		if args.Name != nil {
			value.Name = *args.Name
		}
		f.Users[id] = &value
		return nil
	}
	f.Space.ReadByIdFn = func(ctx context.Context, id int64) (*domainmodelsentity.Space, error) {
		return &domainmodelsentity.Space{Id: id, Slug: "lockmate", IsActive: true}, nil
	}
	f.Space.ReadBySlugFn = func(ctx context.Context, slug string) (*domainmodelsentity.Space, error) {
		return &domainmodelsentity.Space{Id: 10, Slug: slug, IsActive: true}, nil
	}
	f.SpaceMember.ReadByIdFn = func(ctx context.Context, spaceId, id int64) (*domainmodelsentity.SpaceMember, error) {
		if v := f.Members[id]; v != nil && v.SpaceId == spaceId {
			return v, nil
		}
		return nil, NotFound(domainmodelserror.SpaceMemberNotFound)
	}
	f.SpaceMember.ReadByUserIdFn = func(ctx context.Context, spaceId, userId int64) (*domainmodelsentity.SpaceMember, error) {
		for _, v := range f.Members {
			if v.SpaceId == spaceId && v.UserId == userId {
				return v, nil
			}
		}
		return nil, NotFound(domainmodelserror.SpaceMemberNotFound)
	}
	f.SpaceMember.CreateFn = func(ctx context.Context, spaceId int64, args domaincontractsrepository.CreateSpaceMemberArgs) (int64, error) {
		if f.MissingDefault {
			return 0, NotFound(domainmodelserror.DefaultRoleNotFound)
		}
		id := int64(20 + len(f.Members))
		f.Members[id] = &domainmodelsentity.SpaceMember{Id: id, SpaceId: spaceId, UserId: args.UserId, IsActive: true}
		return id, nil
	}
	f.Client.ReadByIdFn = func(ctx context.Context, spaceId, id int64) (*domainmodelsentity.Client, error) {
		return &domainmodelsentity.Client{Id: id, SpaceId: spaceId, Slug: "web", IsActive: true, SecretKeyHash: "hash:secret"}, nil
	}
	f.Client.ReadBySlugFn = func(ctx context.Context, slug string) (*domainmodelsentity.Client, error) {
		return &domainmodelsentity.Client{Id: 30, SpaceId: 10, Slug: slug, IsActive: true, SecretKeyHash: "hash:secret"}, nil
	}
	f.ClassicAuth.ReadByUsernameFn = func(ctx context.Context, username string) (*domainmodelsentity.ClassicAuth, error) {
		for _, v := range f.Classics {
			if v.Username == username {
				return v, nil
			}
		}
		return nil, NotFound(domainmodelserror.ClassicAuthNotFound)
	}
	f.ClassicAuth.ReadByUserIdFn = func(ctx context.Context, userId int64) (*domainmodelsentity.ClassicAuth, error) {
		if v := f.Classics[userId]; v != nil {
			return v, nil
		}
		return nil, NotFound(domainmodelserror.ClassicAuthNotFound)
	}
	f.ClassicAuth.ReadByIdFn = func(ctx context.Context, id int64) (*domainmodelsentity.ClassicAuth, error) {
		for _, v := range f.Classics {
			if v.Id == id {
				return v, nil
			}
		}
		return nil, NotFound(domainmodelserror.ClassicAuthNotFound)
	}
	f.ClassicAuth.CreateFn = func(ctx context.Context, args domaincontractsrepository.CreateClassicAuthArgs) (int64, error) {
		id := int64(40 + len(f.Classics))
		f.Classics[args.UserId] = &domainmodelsentity.ClassicAuth{Id: id, UserId: args.UserId, Username: args.Username, PasswordHash: args.PasswordHash}
		return id, nil
	}
	f.ClassicAuth.UpdateByIdFn = func(ctx context.Context, id int64, args domaincontractsrepository.UpdateClassicAuthArgs) error {
		for userId, current := range f.Classics {
			if current.Id == id {
				v := *current
				if args.PasswordHash != nil {
					v.PasswordHash = *args.PasswordHash
				}
				f.Classics[userId] = &v
				return nil
			}
		}
		return NotFound(domainmodelserror.ClassicAuthNotFound)
	}
	f.ClassicAuth.DeleteByIdFn = func(ctx context.Context, id int64, by *int64) error {
		for userId, v := range f.Classics {
			if v.Id == id {
				delete(f.Classics, userId)
				return nil
			}
		}
		return NotFound(domainmodelserror.ClassicAuthNotFound)
	}
	f.GoogleAuth.ReadBySubjectFn = func(ctx context.Context, subject string) (*domainmodelsentity.GoogleAuth, error) {
		for _, v := range f.Googles {
			if v.Subject == subject {
				return v, nil
			}
		}
		return nil, NotFound(domainmodelserror.GoogleAuthNotFound)
	}
	f.GoogleAuth.ReadByUserIdFn = func(ctx context.Context, id int64) (*domainmodelsentity.GoogleAuth, error) {
		if v := f.Googles[id]; v != nil {
			return v, nil
		}
		return nil, NotFound(domainmodelserror.GoogleAuthNotFound)
	}
	f.GoogleAuth.CreateFn = func(ctx context.Context, args domaincontractsrepository.CreateGoogleAuthArgs) (int64, error) {
		id := int64(50 + len(f.Googles))
		verified := args.IsEmailVerified != nil && *args.IsEmailVerified
		f.Googles[args.UserId] = &domainmodelsentity.GoogleAuth{Id: id, UserId: args.UserId, Subject: args.Subject, Email: args.Email, IsEmailVerified: verified}
		return id, nil
	}
	f.GoogleAuth.UpdateByIdFn = func(ctx context.Context, id int64, args domaincontractsrepository.UpdateGoogleAuthArgs) error {
		for userId, current := range f.Googles {
			if current.Id == id {
				v := *current
				if args.Email != nil {
					v.Email = *args.Email
				}
				if args.IsEmailVerified != nil {
					v.IsEmailVerified = *args.IsEmailVerified
				}
				f.Googles[userId] = &v
				return nil
			}
		}
		return NotFound(domainmodelserror.GoogleAuthNotFound)
	}
	f.SpaceMemberRole.ReadEffectiveRolesBySpaceMemberIdFn = func(ctx context.Context, spaceId, id int64) ([]domainmodelsentity.Role, error) {
		return []domainmodelsentity.Role{{Id: 60, SpaceId: spaceId, Slug: "user"}}, nil
	}
	f.SpaceMemberRole.ReadEffectivePermissionsBySpaceMemberIdFn = func(ctx context.Context, spaceId, id int64) ([]domainmodelsentity.Permission, error) {
		return []domainmodelsentity.Permission{{Id: 70, SpaceId: spaceId, Slug: "profile:get"}}, nil
	}
	f.Guard.CheckFn = func(caller *domainmodelsentity.AccessClaims, permissions []string) error {
		if caller == nil {
			return NotFound(domainmodelserror.BadArgs)
		}
		return nil
	}
	f.Password.HashFn = func(value string) (string, error) { return "hash:" + value, nil }
	f.Password.CompareFn = func(hash, value string) error {
		if hash != "hash:"+value {
			return NotFound(domainmodelserror.PasswordMismatch)
		}
		return nil
	}
	f.Token.GenerateAccessFn = func(claims *domainmodelsentity.AccessClaims) (string, error) { return "access-token", f.TokenErr }
	f.Token.GenerateRefreshFn = func(claims *domainmodelsentity.RefreshClaims) (string, error) { return "refresh-token", f.TokenErr }
	f.Token.ValidateAccessFn = func(spaceId int64, value string) (*domainmodelsentity.AccessClaims, error) {
		v := *f.Caller
		v.Name = "stale"
		v.Permissions = []string{"stale"}
		return &v, nil
	}
	f.Token.ValidateRefreshFn = func(spaceId int64, value string) (*domainmodelsentity.RefreshClaims, error) {
		return &domainmodelsentity.RefreshClaims{UserId: 1, SpaceId: spaceId, SpaceMemberId: 20}, nil
	}
	f.Keygen.GenerateFn = func() (string, string, string, error) { return "raw-secret", "hash:secret", "reda****cted", nil }
	f.Keygen.CompareFn = func(hash, raw string) error {
		if hash != "hash:"+raw {
			return NotFound(domainmodelserror.KeyMismatch)
		}
		return nil
	}
	f.Transactor.WithTxFn = func(ctx context.Context, fn func(context.Context) error) error {
		if ctx.Value(transactionKey{}) != nil {
			return NotFound(domainmodelserror.BadState)
		}
		if f.BeginErr != nil {
			return f.BeginErr
		}
		users, members, classics, googles := clone(f.Users), clone(f.Members), clone(f.Classics), clone(f.Googles)
		f.Transactions++
		f.Committed = false
		err := fn(context.WithValue(ctx, transactionKey{}, true))
		if err == nil {
			err = f.CommitErr
		}
		if err != nil {
			f.Users, f.Members, f.Classics, f.Googles = users, members, classics, googles
			return err
		}
		f.Committed = true
		return nil
	}
	return f
}

func clone[T any](values map[int64]*T) map[int64]*T {
	out := make(map[int64]*T, len(values))
	for k, v := range values {
		out[k] = v
	}
	return out
}
func NotFound(kind error) error {
	return domainmodelserror.NewError("test dependency failure", kind, nil)
}
func RequireTx(t *testing.T, ctx context.Context) {
	t.Helper()
	if ctx.Value(transactionKey{}) == nil {
		t.Fatal("missing callback transaction context")
	}
}
func InTransaction() context.Context {
	return context.WithValue(context.Background(), transactionKey{}, true)
}
func RequireError(t *testing.T, err, kind error) {
	t.Helper()
	if !errors.Is(err, kind) {
		t.Fatalf("got %v, want %v", err, kind)
	}
}
