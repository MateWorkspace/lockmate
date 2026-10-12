package presentationgrpcservermanagementgoogleauth

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
	lockmatev1.UnimplementedGoogleAuthServiceServer
	googleAuth domainusecasesmanagement.GoogleAuth
}

func NewServer(googleAuth domainusecasesmanagement.GoogleAuth) lockmatev1.GoogleAuthServiceServer {
	return &server{googleAuth: googleAuth}
}

func (s *server) ReadById(ctx context.Context, req *lockmatev1.GoogleAuthServiceReadByIdRequest) (*lockmatev1.GoogleAuthServiceReadByIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.googleAuth.ReadById(ctx, caller, req.GetId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.GoogleAuthServiceReadByIdResponse{GoogleAuth: presentationgrpcmapper.GoogleAuth(*result)}, nil
}

func (s *server) ReadByUserId(ctx context.Context, req *lockmatev1.GoogleAuthServiceReadByUserIdRequest) (*lockmatev1.GoogleAuthServiceReadByUserIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.googleAuth.ReadByUserId(ctx, caller, req.GetUserId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.GoogleAuthServiceReadByUserIdResponse{GoogleAuth: presentationgrpcmapper.GoogleAuth(*result)}, nil
}

func (s *server) ReadBySubject(ctx context.Context, req *lockmatev1.GoogleAuthServiceReadBySubjectRequest) (*lockmatev1.GoogleAuthServiceReadBySubjectResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.googleAuth.ReadBySubject(ctx, caller, req.GetSubject())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.GoogleAuthServiceReadBySubjectResponse{GoogleAuth: presentationgrpcmapper.GoogleAuth(*result)}, nil
}

func (s *server) List(ctx context.Context, req *lockmatev1.GoogleAuthServiceListRequest) (*lockmatev1.GoogleAuthServiceListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	page, err := s.googleAuth.ReadByFilter(ctx, caller, domainusecasesmanagement.FilterGoogleAuthRequest{
		Page:            req.GetPage(),
		Limit:           req.GetLimit(),
		Search:          req.Search,
		UserId:          req.UserId,
		IsEmailVerified: req.IsEmailVerified,
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.GoogleAuthServiceListResponse{
		Data: presentationgrpcmapper.GoogleAuths(page.Items),
		Page: &lockmatev1.PageResponse{Page: req.GetPage(), Limit: req.GetLimit(), TotalItems: page.Total},
	}, nil
}
