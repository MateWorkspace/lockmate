package presentationgrpcservermanagementspace

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
	lockmatev1.UnimplementedSpaceServiceServer
	space domainusecasesmanagement.Space
}

func NewServer(space domainusecasesmanagement.Space) lockmatev1.SpaceServiceServer {
	return &server{space: space}
}

func (s *server) Create(ctx context.Context, req *lockmatev1.SpaceServiceCreateRequest) (*lockmatev1.SpaceServiceCreateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	id, err := s.space.Create(ctx, caller, domainusecasesmanagement.CreateSpaceRequest{
		Slug:        req.GetSlug(),
		Name:        req.GetName(),
		Description: req.Description,
		IsActive:    req.IsActive,
		Preferences: presentationgrpcmapper.OptionalPreferences(req.Preferences),
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.SpaceServiceCreateResponse{Id: id}, nil
}

func (s *server) ReadById(ctx context.Context, req *lockmatev1.SpaceServiceReadByIdRequest) (*lockmatev1.SpaceServiceReadByIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.space.ReadById(ctx, caller, req.GetId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.SpaceServiceReadByIdResponse{Space: presentationgrpcmapper.Space(*result)}, nil
}

func (s *server) ReadBySlug(ctx context.Context, req *lockmatev1.SpaceServiceReadBySlugRequest) (*lockmatev1.SpaceServiceReadBySlugResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.space.ReadBySlug(ctx, caller, req.GetSlug())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.SpaceServiceReadBySlugResponse{Space: presentationgrpcmapper.Space(*result)}, nil
}

func (s *server) List(ctx context.Context, req *lockmatev1.SpaceServiceListRequest) (*lockmatev1.SpaceServiceListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	page, err := s.space.ReadByFilter(ctx, caller, domainusecasesmanagement.FilterSpaceRequest{
		Page:     req.GetPage(),
		Limit:    req.GetLimit(),
		Search:   req.Search,
		IsActive: req.IsActive,
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.SpaceServiceListResponse{
		Data: presentationgrpcmapper.Spaces(page.Items),
		Page: &lockmatev1.PageResponse{Page: req.GetPage(), Limit: req.GetLimit(), TotalItems: page.Total},
	}, nil
}

func (s *server) Update(ctx context.Context, req *lockmatev1.SpaceServiceUpdateRequest) (*lockmatev1.SpaceServiceUpdateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.space.UpdateById(ctx, caller, req.GetId(), domainusecasesmanagement.UpdateSpaceRequest{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
		Preferences: presentationgrpcmapper.OptionalPreferences(req.Preferences),
	}); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.SpaceServiceUpdateResponse{}, nil
}

func (s *server) Delete(ctx context.Context, req *lockmatev1.SpaceServiceDeleteRequest) (*lockmatev1.SpaceServiceDeleteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.space.DeleteById(ctx, caller, req.GetId()); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.SpaceServiceDeleteResponse{}, nil
}
