package presentationgrpcservermiscversion

import (
	"context"

	lockmatev1 "github.com/MateWorkspace/lockmate-proto/gen/go/lockmate/v1"
	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
)

type server struct {
	lockmatev1.UnimplementedVersionServiceServer
	info domainmodelsapp.Info
}

func NewServer(info domainmodelsapp.Info) lockmatev1.VersionServiceServer {
	return &server{info: info}
}

func (s *server) Get(_ context.Context, _ *lockmatev1.VersionServiceGetRequest) (*lockmatev1.VersionServiceGetResponse, error) {
	return &lockmatev1.VersionServiceGetResponse{Name: s.info.Name, Version: s.info.Version}, nil
}
