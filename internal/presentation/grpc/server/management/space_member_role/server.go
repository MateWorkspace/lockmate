package presentationgrpcservermanagementspacememberrole

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
	lockmatev1.UnimplementedSpaceMemberRoleServiceServer
	spaceMemberRole domainusecasesmanagement.SpaceMemberRole
}

func NewServer(spaceMemberRole domainusecasesmanagement.SpaceMemberRole) lockmatev1.SpaceMemberRoleServiceServer {
	return &server{spaceMemberRole: spaceMemberRole}
}

func (s *server) Create(ctx context.Context, req *lockmatev1.SpaceMemberRoleServiceCreateRequest) (*lockmatev1.SpaceMemberRoleServiceCreateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	id, err := s.spaceMemberRole.Create(ctx, caller, req.GetSpaceId(), domainusecasesmanagement.CreateSpaceMemberRoleRequest{
		SpaceMemberId: req.GetSpaceMemberId(),
		RoleId:        req.GetRoleId(),
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.SpaceMemberRoleServiceCreateResponse{Id: id}, nil
}

func (s *server) ReadById(ctx context.Context, req *lockmatev1.SpaceMemberRoleServiceReadByIdRequest) (*lockmatev1.SpaceMemberRoleServiceReadByIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.spaceMemberRole.ReadById(ctx, caller, req.GetSpaceId(), req.GetId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.SpaceMemberRoleServiceReadByIdResponse{Details: presentationgrpcmapper.SpaceMemberRoleDetails(*result)}, nil
}

func (s *server) ReadByPair(ctx context.Context, req *lockmatev1.SpaceMemberRoleServiceReadByPairRequest) (*lockmatev1.SpaceMemberRoleServiceReadByPairResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.spaceMemberRole.ReadBySpaceMemberIdAndRoleId(ctx, caller, req.GetSpaceId(), req.GetSpaceMemberId(), req.GetRoleId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.SpaceMemberRoleServiceReadByPairResponse{Details: presentationgrpcmapper.SpaceMemberRoleDetails(*result)}, nil
}

func (s *server) ReadBySpaceMemberId(ctx context.Context, req *lockmatev1.SpaceMemberRoleServiceReadBySpaceMemberIdRequest) (*lockmatev1.SpaceMemberRoleServiceReadBySpaceMemberIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.spaceMemberRole.ReadBySpaceMemberId(ctx, caller, req.GetSpaceId(), req.GetSpaceMemberId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.SpaceMemberRoleServiceReadBySpaceMemberIdResponse{Data: presentationgrpcmapper.SpaceMemberRoleWithRoles(result)}, nil
}

func (s *server) ReadByRoleId(ctx context.Context, req *lockmatev1.SpaceMemberRoleServiceReadByRoleIdRequest) (*lockmatev1.SpaceMemberRoleServiceReadByRoleIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.spaceMemberRole.ReadByRoleId(ctx, caller, req.GetSpaceId(), req.GetRoleId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.SpaceMemberRoleServiceReadByRoleIdResponse{Data: presentationgrpcmapper.SpaceMemberRoleWithSpaceMembers(result)}, nil
}

func (s *server) Delete(ctx context.Context, req *lockmatev1.SpaceMemberRoleServiceDeleteRequest) (*lockmatev1.SpaceMemberRoleServiceDeleteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.spaceMemberRole.DeleteById(ctx, caller, req.GetSpaceId(), req.GetId()); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.SpaceMemberRoleServiceDeleteResponse{}, nil
}

func (s *server) DeleteBySelectors(ctx context.Context, req *lockmatev1.SpaceMemberRoleServiceDeleteBySelectorsRequest) (*lockmatev1.SpaceMemberRoleServiceDeleteBySelectorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.spaceMemberRole.DeleteBySpaceMemberIdOrRoleId(ctx, caller, req.GetSpaceId(), domainusecasesmanagement.DeleteSpaceMemberRoleRequest{
		SpaceMemberId: req.SpaceMemberId,
		RoleId:        req.RoleId,
	}); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.SpaceMemberRoleServiceDeleteBySelectorsResponse{}, nil
}
