package presentationgrpcservermiscjwks

import (
	"context"

	lockmatev1 "github.com/MateWorkspace/lockmate-proto/gen/go/lockmate/v1"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
	presentationgrpcmapper "github.com/MateWorkspace/lockmate/internal/presentation/grpc/mapper"
)

type server struct {
	lockmatev1.UnimplementedJwksServiceServer
	jwks domainusecasesmisc.Jwks
}

func NewServer(jwks domainusecasesmisc.Jwks) lockmatev1.JwksServiceServer {
	return &server{jwks: jwks}
}

func (s *server) Get(ctx context.Context, _ *lockmatev1.JwksServiceGetRequest) (*lockmatev1.JwksServiceGetResponse, error) {
	keys, err := s.jwks.PublicKeys(ctx)
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.JwksServiceGetResponse{Keys: presentationgrpcmapper.Jwks(keys)}, nil
}
