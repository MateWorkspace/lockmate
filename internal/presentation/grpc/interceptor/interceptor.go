// Package presentationgrpcinterceptor mirrors the HTTP request pipeline as a
// gRPC unary interceptor chain: recovery, trace, client authentication, caller
// IP allowlist, and bearer authentication.
package presentationgrpcinterceptor

import (
	"context"
	"net/netip"
	"strings"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
	presentationgrpcmapper "github.com/MateWorkspace/lockmate/internal/presentation/grpc/mapper"
	presentationshared "github.com/MateWorkspace/lockmate/internal/presentation/shared"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	clientSlugKey    = "x-client-slug"
	clientSecretKey  = "x-client-secret"
	clientExecIdKey  = "x-client-exec-id"
	authorizationKey = "authorization"
)

// Args carries the interceptor dependencies.
type Args struct {
	Client         domainusecasesauth.Client
	Session        domainusecasesauth.Session
	Tracer         domainusecasesmisc.Tracer
	IpAllowlist    domainusecasesmisc.IpAllowlist
	TrustedProxies []netip.Prefix
	ClientIpHeader string
}

// Unary returns the chained unary server option.
func Unary(args Args) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		recovery(),
		trace(args.Tracer),
		client(args.Client),
		ipAllowlist(args.IpAllowlist, args.TrustedProxies, args.ClientIpHeader),
		bearer(args.Session),
	)
}

func recovery() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ any, err error) {
		defer func() {
			if recover() != nil {
				err = presentationgrpcmapper.Error(domainmodelserror.NewError("request failed", domainmodelserror.Failure, nil))
			}
		}()

		return handler(ctx, req)
	}
}

func trace(tracer domainusecasesmisc.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		slug := firstValue(ctx, clientSlugKey)
		execId, _ := uuid.Parse(firstValue(ctx, clientExecIdKey))

		newCtx, err := tracer.Inject(ctx, slug, execId)
		if err != nil {
			return nil, presentationgrpcmapper.Error(err)
		}
		if newCtx == nil {
			return nil, presentationgrpcmapper.Error(domainmodelserror.NewError("trace context is missing", domainmodelserror.Failure, nil))
		}

		return handler(newCtx, req)
	}
}

func client(auth domainusecasesauth.Client) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if isPublic(info.FullMethod) {
			return handler(ctx, req)
		}

		slug, err := singleValue(ctx, clientSlugKey)
		if err != nil {
			return nil, presentationgrpcmapper.Error(err)
		}
		secret, err := singleValue(ctx, clientSecretKey)
		if err != nil {
			return nil, presentationgrpcmapper.Error(err)
		}
		if _, err := execIdValue(ctx); err != nil {
			return nil, presentationgrpcmapper.Error(err)
		}

		identity, err := auth.Authenticate(ctx, domainusecasesauth.AuthenticateClientRequest{Slug: slug, Secret: secret})
		if err != nil {
			return nil, presentationgrpcmapper.Error(err)
		}
		if identity == nil || identity.Id <= 0 || identity.SpaceId <= 0 {
			return nil, presentationgrpcmapper.Error(domainmodelserror.NewError("client identity is missing", domainmodelserror.Failure, nil))
		}

		return handler(presentationshared.InjectClientIdentity(ctx, identity), req)
	}
}

func ipAllowlist(allowlist domainusecasesmisc.IpAllowlist, trusted []netip.Prefix, header string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if isPublic(info.FullMethod) {
			return handler(ctx, req)
		}

		identity := presentationshared.ClientIdentity(ctx)
		if identity == nil {
			return nil, presentationgrpcmapper.Error(domainmodelserror.NewError("client authentication is required", domainmodelserror.Unauthorized, nil))
		}

		address := presentationshared.Resolve(peerAddr(ctx), forwardedValues(ctx, header), trusted)
		if !address.IsValid() {
			return nil, presentationgrpcmapper.Error(domainmodelserror.NewError("caller IP address is unavailable", domainmodelserror.Forbidden, nil))
		}

		allowed, err := allowlist.IsAllowedIpForClient(ctx, identity.SpaceId, identity.Id, address.String())
		if err != nil {
			return nil, presentationgrpcmapper.Error(err)
		}
		if !allowed {
			return nil, presentationgrpcmapper.Error(domainmodelserror.NewError("caller IP address is not allowed for this client", domainmodelserror.Forbidden, nil))
		}

		return handler(ctx, req)
	}
}

func bearer(session domainusecasesauth.Session) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !requiresBearer(info.FullMethod) {
			return handler(ctx, req)
		}

		identity := presentationshared.ClientIdentity(ctx)
		if identity == nil {
			return nil, presentationgrpcmapper.Error(domainmodelserror.NewError("client authentication is required", domainmodelserror.Unauthorized, nil))
		}

		token, err := bearerToken(ctx)
		if err != nil {
			return nil, presentationgrpcmapper.Error(err)
		}

		claims, err := session.AuthenticateAccessToken(ctx, identity, token)
		if err != nil {
			return nil, presentationgrpcmapper.Error(err)
		}
		if claims == nil || claims.UserId <= 0 || claims.SpaceId <= 0 || claims.SpaceMemberId <= 0 || claims.SpaceId != identity.SpaceId {
			return nil, presentationgrpcmapper.Error(domainmodelserror.NewError("access claims are missing", domainmodelserror.Failure, nil))
		}

		return handler(presentationshared.InjectAccessClaims(ctx, claims), req)
	}
}

func isPublic(method string) bool {
	switch method {
	case "/lockmate.v1.VersionService/Get", "/grpc.health.v1.Health/Check", "/grpc.health.v1.Health/Watch":
		return true
	default:
		return false
	}
}

func requiresBearer(method string) bool {
	switch serviceName(method) {
	case "AccountService", "SecurityService", "SpaceService", "PermissionService", "RoleService",
		"UserService", "ClassicAuthService", "GoogleAuthService", "ClientService", "SpaceMemberService",
		"RolePermissionService", "SpaceMemberRoleService":
		return true
	default:
		return false
	}
}

func serviceName(method string) string {
	trimmed := strings.TrimPrefix(method, "/")
	if index := strings.LastIndex(trimmed, "/"); index >= 0 {
		trimmed = trimmed[:index]
	}
	if index := strings.LastIndex(trimmed, "."); index >= 0 {
		return trimmed[index+1:]
	}

	return trimmed
}

func firstValue(ctx context.Context, key string) string {
	values := metadataValues(ctx, key)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func singleValue(ctx context.Context, key string) (string, error) {
	values := metadataValues(ctx, key)
	if len(values) != 1 || strings.TrimSpace(values[0]) == "" {
		return "", domainmodelserror.NewError(key+" must occur once", domainmodelserror.BadArgs, nil)
	}

	return values[0], nil
}

func execIdValue(ctx context.Context) (uuid.UUID, error) {
	value, err := singleValue(ctx, clientExecIdKey)
	if err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, domainmodelserror.NewError(clientExecIdKey+" must be a nonzero UUID", domainmodelserror.BadArgs, nil)
	}

	return id, nil
}

func bearerToken(ctx context.Context) (string, error) {
	value, err := singleValue(ctx, authorizationKey)
	if err != nil {
		return "", domainmodelserror.NewError("bearer authorization is required", domainmodelserror.Unauthorized, nil)
	}

	parts := strings.Fields(value)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", domainmodelserror.NewError("bearer authorization is required", domainmodelserror.Unauthorized, nil)
	}

	return parts[1], nil
}

func metadataValues(ctx context.Context, key string) []string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	return md.Get(key)
}

func forwardedValues(ctx context.Context, header string) []string {
	header = strings.ToLower(strings.TrimSpace(header))
	if header == "" {
		return nil
	}

	return metadataValues(ctx, header)
}

func peerAddr(ctx context.Context) netip.Addr {
	value, ok := peer.FromContext(ctx)
	if !ok || value == nil {
		return netip.Addr{}
	}

	return presentationshared.FromAddr(value.Addr)
}
