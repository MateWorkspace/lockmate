package presentationgrpcservermanagementclassicauth

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
	lockmatev1.UnimplementedClassicAuthServiceServer
	classicAuth domainusecasesmanagement.ClassicAuth
}

func NewServer(classicAuth domainusecasesmanagement.ClassicAuth) lockmatev1.ClassicAuthServiceServer {
	return &server{classicAuth: classicAuth}
}

func (s *server) Create(ctx context.Context, req *lockmatev1.ClassicAuthServiceCreateRequest) (*lockmatev1.ClassicAuthServiceCreateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	id, err := s.classicAuth.Create(ctx, caller, domainusecasesmanagement.CreateClassicAuthRequest{
		UserId:          req.GetUserId(),
		Username:        req.GetUsername(),
		Email:           req.Email,
		Phone:           req.Phone,
		Password:        req.GetPassword(),
		IsEmailVerified: req.IsEmailVerified,
		IsPhoneVerified: req.IsPhoneVerified,
		Preferences:     presentationgrpcmapper.OptionalPreferences(req.Preferences),
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.ClassicAuthServiceCreateResponse{Id: id}, nil
}

func (s *server) ReadById(ctx context.Context, req *lockmatev1.ClassicAuthServiceReadByIdRequest) (*lockmatev1.ClassicAuthServiceReadByIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.classicAuth.ReadById(ctx, caller, req.GetId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.ClassicAuthServiceReadByIdResponse{ClassicAuth: presentationgrpcmapper.ClassicAuth(*result)}, nil
}

func (s *server) ReadByUserId(ctx context.Context, req *lockmatev1.ClassicAuthServiceReadByUserIdRequest) (*lockmatev1.ClassicAuthServiceReadByUserIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.classicAuth.ReadByUserId(ctx, caller, req.GetUserId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.ClassicAuthServiceReadByUserIdResponse{ClassicAuth: presentationgrpcmapper.ClassicAuth(*result)}, nil
}

func (s *server) ReadByUsername(ctx context.Context, req *lockmatev1.ClassicAuthServiceReadByUsernameRequest) (*lockmatev1.ClassicAuthServiceReadByUsernameResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.classicAuth.ReadByUsername(ctx, caller, req.GetUsername())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.ClassicAuthServiceReadByUsernameResponse{ClassicAuth: presentationgrpcmapper.ClassicAuth(*result)}, nil
}

func (s *server) ReadByEmail(ctx context.Context, req *lockmatev1.ClassicAuthServiceReadByEmailRequest) (*lockmatev1.ClassicAuthServiceReadByEmailResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.classicAuth.ReadByEmail(ctx, caller, req.GetEmail())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.ClassicAuthServiceReadByEmailResponse{ClassicAuth: presentationgrpcmapper.ClassicAuth(*result)}, nil
}

func (s *server) ReadByPhone(ctx context.Context, req *lockmatev1.ClassicAuthServiceReadByPhoneRequest) (*lockmatev1.ClassicAuthServiceReadByPhoneResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.classicAuth.ReadByPhone(ctx, caller, req.GetPhone())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.ClassicAuthServiceReadByPhoneResponse{ClassicAuth: presentationgrpcmapper.ClassicAuth(*result)}, nil
}

func (s *server) List(ctx context.Context, req *lockmatev1.ClassicAuthServiceListRequest) (*lockmatev1.ClassicAuthServiceListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	page, err := s.classicAuth.ReadByFilter(ctx, caller, domainusecasesmanagement.FilterClassicAuthRequest{
		Page:            req.GetPage(),
		Limit:           req.GetLimit(),
		Search:          req.Search,
		UserId:          req.UserId,
		IsEmailVerified: req.IsEmailVerified,
		IsPhoneVerified: req.IsPhoneVerified,
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.ClassicAuthServiceListResponse{
		Data: presentationgrpcmapper.ClassicAuths(page.Items),
		Page: &lockmatev1.PageResponse{Page: req.GetPage(), Limit: req.GetLimit(), TotalItems: page.Total},
	}, nil
}

func (s *server) Update(ctx context.Context, req *lockmatev1.ClassicAuthServiceUpdateRequest) (*lockmatev1.ClassicAuthServiceUpdateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.classicAuth.UpdateById(ctx, caller, req.GetId(), domainusecasesmanagement.UpdateClassicAuthRequest{
		Username:        req.Username,
		Email:           req.Email,
		EmailSetNull:    req.GetEmailSetNull(),
		Phone:           req.Phone,
		PhoneSetNull:    req.GetPhoneSetNull(),
		IsEmailVerified: req.IsEmailVerified,
		IsPhoneVerified: req.IsPhoneVerified,
		Preferences:     presentationgrpcmapper.OptionalPreferences(req.Preferences),
	}); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.ClassicAuthServiceUpdateResponse{}, nil
}

func (s *server) ResetPassword(ctx context.Context, req *lockmatev1.ClassicAuthServiceResetPasswordRequest) (*lockmatev1.ClassicAuthServiceResetPasswordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.classicAuth.ResetPassword(ctx, caller, req.GetUserId(), domainusecasesmanagement.ResetUserPasswordRequest{Password: req.GetPassword()}); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.ClassicAuthServiceResetPasswordResponse{}, nil
}

func (s *server) Delete(ctx context.Context, req *lockmatev1.ClassicAuthServiceDeleteRequest) (*lockmatev1.ClassicAuthServiceDeleteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.classicAuth.DeleteById(ctx, caller, req.GetId()); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.ClassicAuthServiceDeleteResponse{}, nil
}
