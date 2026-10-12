package presentationgrpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	lockmatev1 "github.com/MateWorkspace/lockmate-proto/gen/go/lockmate/v1"
	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
	domainusecasesprofile "github.com/MateWorkspace/lockmate/internal/domain/usecases/profile"
	presentationgrpcinterceptor "github.com/MateWorkspace/lockmate/internal/presentation/grpc/interceptor"
	presentationgrpcregister "github.com/MateWorkspace/lockmate/internal/presentation/grpc/register"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type clientStub struct{ domainusecasesauth.Client }

func (clientStub) Authenticate(_ context.Context, request domainusecasesauth.AuthenticateClientRequest) (*domainmodelsentity.ClientIdentity, error) {
	return &domainmodelsentity.ClientIdentity{Id: 1, SpaceId: 2, Slug: request.Slug}, nil
}

type sessionStub struct{ domainusecasesauth.Session }

func (sessionStub) SignIn(context.Context, *domainmodelsentity.ClientIdentity, domainusecasesauth.SignInRequest) (*domainusecasesauth.TokenResult, error) {
	return &domainusecasesauth.TokenResult{AccessToken: "access", RefreshToken: "refresh"}, nil
}

func (sessionStub) AuthenticateAccessToken(context.Context, *domainmodelsentity.ClientIdentity, string) (*domainmodelsentity.AccessClaims, error) {
	return &domainmodelsentity.AccessClaims{UserId: 3, SpaceId: 2, SpaceMemberId: 4}, nil
}

type tracerStub struct{ domainusecasesmisc.Tracer }

func (tracerStub) Inject(ctx context.Context, _ string, _ uuid.UUID) (context.Context, error) {
	return ctx, nil
}

type ipStub struct{ allowed bool }

func (s ipStub) IsAllowedIpForClient(context.Context, int64, int64, string) (bool, error) {
	return s.allowed, nil
}

type accountStub struct{ domainusecasesprofile.Account }

func (accountStub) GetProfile(context.Context, *domainmodelsentity.AccessClaims) (*domainusecasesprofile.ProfileResult, error) {
	return &domainusecasesprofile.ProfileResult{User: domainmodelsentity.User{Id: 3}, Space: domainmodelsentity.Space{Id: 2}, SpaceMember: domainmodelsentity.SpaceMember{Id: 4}}, nil
}

func TestGrpcPipeline(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	server := grpc.NewServer(presentationgrpcinterceptor.Unary(presentationgrpcinterceptor.Args{
		Client:      clientStub{},
		Session:     sessionStub{},
		Tracer:      tracerStub{},
		IpAllowlist: ipStub{allowed: true},
	}))
	presentationgrpcregister.RegisterAuth(server, presentationgrpcregister.AuthDependencies{Session: sessionStub{}})
	presentationgrpcregister.RegisterProfile(server, presentationgrpcregister.ProfileDependencies{Account: accountStub{}})
	presentationgrpcregister.RegisterMisc(server, presentationgrpcregister.MiscDependencies{Version: domainmodelsapp.Info{Name: "lockmate", Version: "test"}})
	go func() { _ = server.Serve(listener) }()
	defer server.Stop()

	conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	version := lockmatev1.NewVersionServiceClient(conn)
	session := lockmatev1.NewSessionServiceClient(conn)
	account := lockmatev1.NewAccountServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Public version requires no credentials.
	if response, err := version.Get(ctx, &lockmatev1.VersionServiceGetRequest{}); err != nil || response.GetVersion() != "test" {
		t.Fatalf("version: %v %v", response, err)
	}

	// Client-auth RPC without metadata is rejected.
	if _, err := session.SignIn(ctx, &lockmatev1.SessionServiceSignInRequest{Username: "u", Password: "p"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("missing metadata accepted: %v", err)
	}

	clientCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"x-client-slug", "lockmate",
		"x-client-secret", "secret",
		"x-client-exec-id", uuid.NewString(),
	))

	// Sign-in succeeds with client metadata.
	signedIn, err := session.SignIn(clientCtx, &lockmatev1.SessionServiceSignInRequest{Username: "u", Password: "p"})
	if err != nil || signedIn.GetToken().GetAccessToken() != "access" {
		t.Fatalf("sign-in: %v %v", signedIn, err)
	}

	// Protected RPC requires bearer authorization.
	if _, err := account.GetProfile(clientCtx, &lockmatev1.AccountServiceGetProfileRequest{}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("missing bearer accepted: %v", err)
	}

	bearerCtx := metadata.AppendToOutgoingContext(clientCtx, "authorization", "Bearer token")
	profile, err := account.GetProfile(bearerCtx, &lockmatev1.AccountServiceGetProfileRequest{})
	if err != nil || profile.GetProfile().GetUser().GetId() != 3 {
		t.Fatalf("profile: %v %v", profile, err)
	}
}

func TestGrpcPipelineDeniesUnlistedIP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	server := grpc.NewServer(presentationgrpcinterceptor.Unary(presentationgrpcinterceptor.Args{
		Client:      clientStub{},
		Session:     sessionStub{},
		Tracer:      tracerStub{},
		IpAllowlist: ipStub{allowed: false},
	}))
	presentationgrpcregister.RegisterAuth(server, presentationgrpcregister.AuthDependencies{Session: sessionStub{}})
	go func() { _ = server.Serve(listener) }()
	defer server.Stop()

	conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
		"x-client-slug", "lockmate",
		"x-client-secret", "secret",
		"x-client-exec-id", uuid.NewString(),
	))

	if _, err := lockmatev1.NewSessionServiceClient(conn).SignIn(ctx, &lockmatev1.SessionServiceSignInRequest{Username: "u", Password: "p"}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("unlisted IP accepted: %v", err)
	}
}
