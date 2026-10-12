package presentationgrpcservermanagementclient

import (
	"context"

	lockmatev1 "github.com/MateWorkspace/lockmate-proto/gen/go/lockmate/v1"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	presentationgrpcmapper "github.com/MateWorkspace/lockmate/internal/presentation/grpc/mapper"
	presentationgrpcservershared "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/shared"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	lockmatev1.UnimplementedClientServiceServer
	client domainusecasesmanagement.Client
}

func NewServer(client domainusecasesmanagement.Client) lockmatev1.ClientServiceServer {
	return &server{client: client}
}

func (s *server) Create(ctx context.Context, req *lockmatev1.ClientServiceCreateRequest) (*lockmatev1.ClientServiceCreateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.client.Create(ctx, caller, req.GetSpaceId(), domainusecasesmanagement.CreateClientRequest{
		AuthRedirectUri: req.AuthRedirectUri,
		Slug:            req.GetSlug(),
		Name:            req.GetName(),
		Description:     req.Description,
		IsActive:        req.IsActive,
		Preferences:     presentationgrpcmapper.OptionalPreferences(req.Preferences),
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.ClientServiceCreateResponse{ClientSecret: clientSecret(*result)}, nil
}

func (s *server) ReadById(ctx context.Context, req *lockmatev1.ClientServiceReadByIdRequest) (*lockmatev1.ClientServiceReadByIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.client.ReadById(ctx, caller, req.GetSpaceId(), req.GetId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.ClientServiceReadByIdResponse{Client: presentationgrpcmapper.Client(*result)}, nil
}

func (s *server) ReadBySlug(ctx context.Context, req *lockmatev1.ClientServiceReadBySlugRequest) (*lockmatev1.ClientServiceReadBySlugResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.client.ReadBySlug(ctx, caller, req.GetSpaceId(), req.GetSlug())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.ClientServiceReadBySlugResponse{Client: presentationgrpcmapper.Client(*result)}, nil
}

func (s *server) List(ctx context.Context, req *lockmatev1.ClientServiceListRequest) (*lockmatev1.ClientServiceListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	page, err := s.client.ReadByFilter(ctx, caller, req.GetSpaceId(), domainusecasesmanagement.FilterClientRequest{
		Page:     req.GetPage(),
		Limit:    req.GetLimit(),
		Search:   req.Search,
		IsActive: req.IsActive,
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.ClientServiceListResponse{
		Data: presentationgrpcmapper.Clients(page.Items),
		Page: &lockmatev1.PageResponse{Page: req.GetPage(), Limit: req.GetLimit(), TotalItems: page.Total},
	}, nil
}

func (s *server) Update(ctx context.Context, req *lockmatev1.ClientServiceUpdateRequest) (*lockmatev1.ClientServiceUpdateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.client.UpdateById(ctx, caller, req.GetSpaceId(), req.GetId(), domainusecasesmanagement.UpdateClientRequest{
		AuthRedirectUri:        req.AuthRedirectUri,
		SetAuthRedirectUriNull: req.GetAuthRedirectUriSetNull(),
		Name:                   req.Name,
		Description:            req.Description,
		IsActive:               req.IsActive,
		Preferences:            presentationgrpcmapper.OptionalPreferences(req.Preferences),
	}); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.ClientServiceUpdateResponse{}, nil
}

func (s *server) RotateSecret(ctx context.Context, req *lockmatev1.ClientServiceRotateSecretRequest) (*lockmatev1.ClientServiceRotateSecretResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.client.RotateSecret(ctx, caller, req.GetSpaceId(), req.GetId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.ClientServiceRotateSecretResponse{ClientSecret: clientSecret(*result)}, nil
}

func (s *server) Delete(ctx context.Context, req *lockmatev1.ClientServiceDeleteRequest) (*lockmatev1.ClientServiceDeleteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.client.DeleteById(ctx, caller, req.GetSpaceId(), req.GetId()); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.ClientServiceDeleteResponse{}, nil
}

func (s *server) CreateIpAllowlist(ctx context.Context, req *lockmatev1.ClientServiceCreateIpAllowlistRequest) (*lockmatev1.ClientServiceCreateIpAllowlistResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	id, err := s.client.CreateIpAllowlist(ctx, caller, req.GetSpaceId(), req.GetClientId(), domainusecasesmanagement.CreateClientIpAllowlistRequest{
		Cidr:     req.GetCidr(),
		IsActive: req.IsActive,
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.ClientServiceCreateIpAllowlistResponse{Id: id}, nil
}

func (s *server) ReadIpAllowlistById(ctx context.Context, req *lockmatev1.ClientServiceReadIpAllowlistByIdRequest) (*lockmatev1.ClientServiceReadIpAllowlistByIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.client.ReadIpAllowlistById(ctx, caller, req.GetSpaceId(), req.GetClientId(), req.GetId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.ClientServiceReadIpAllowlistByIdResponse{Entry: presentationgrpcmapper.ClientIpAllowlist(*result)}, nil
}

func (s *server) ReadIpAllowlistByCidr(ctx context.Context, req *lockmatev1.ClientServiceReadIpAllowlistByCidrRequest) (*lockmatev1.ClientServiceReadIpAllowlistByCidrResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.client.ReadIpAllowlistByCidr(ctx, caller, req.GetSpaceId(), req.GetClientId(), req.GetCidr())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.ClientServiceReadIpAllowlistByCidrResponse{Entry: presentationgrpcmapper.ClientIpAllowlist(*result)}, nil
}

func (s *server) ListIpAllowlists(ctx context.Context, req *lockmatev1.ClientServiceListIpAllowlistsRequest) (*lockmatev1.ClientServiceListIpAllowlistsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.client.ReadIpAllowlistsByClientId(ctx, caller, req.GetSpaceId(), req.GetClientId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.ClientServiceListIpAllowlistsResponse{Data: presentationgrpcmapper.ClientIpAllowlists(result)}, nil
}

func (s *server) FilterIpAllowlists(ctx context.Context, req *lockmatev1.ClientServiceFilterIpAllowlistsRequest) (*lockmatev1.ClientServiceFilterIpAllowlistsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	page, err := s.client.ReadIpAllowlistsByFilter(ctx, caller, req.GetSpaceId(), req.GetClientId(), domainusecasesmanagement.FilterClientIpAllowlistRequest{
		Page:     req.GetPage(),
		Limit:    req.GetLimit(),
		Search:   req.Search,
		IsActive: req.IsActive,
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.ClientServiceFilterIpAllowlistsResponse{
		Data: presentationgrpcmapper.ClientIpAllowlists(page.Items),
		Page: &lockmatev1.PageResponse{Page: req.GetPage(), Limit: req.GetLimit(), TotalItems: page.Total},
	}, nil
}

func (s *server) UpdateIpAllowlist(ctx context.Context, req *lockmatev1.ClientServiceUpdateIpAllowlistRequest) (*lockmatev1.ClientServiceUpdateIpAllowlistResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.client.UpdateIpAllowlistById(ctx, caller, req.GetSpaceId(), req.GetClientId(), req.GetId(), domainusecasesmanagement.UpdateClientIpAllowlistRequest{
		Cidr:     req.Cidr,
		IsActive: req.IsActive,
	}); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.ClientServiceUpdateIpAllowlistResponse{}, nil
}

func (s *server) DeleteIpAllowlist(ctx context.Context, req *lockmatev1.ClientServiceDeleteIpAllowlistRequest) (*lockmatev1.ClientServiceDeleteIpAllowlistResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.client.DeleteIpAllowlistById(ctx, caller, req.GetSpaceId(), req.GetClientId(), req.GetId()); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.ClientServiceDeleteIpAllowlistResponse{}, nil
}

func clientSecret(value domainusecasesmanagement.ClientSecretResult) *lockmatev1.ClientSecret {
	return &lockmatev1.ClientSecret{Client: presentationgrpcmapper.Client(value.Client), Secret: value.Secret}
}
