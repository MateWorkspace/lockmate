package domainmodelsapp

import (
	"time"

	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
)

type Env struct {
	LoggerFormat   domainmodelslogger.Format
	LoggerLevel    domainmodelslogger.Level
	LoggerUseTrace bool

	PortGrpcServer uint16
	PortHttpServer uint16

	PostgresHost           string
	PostgresPort           uint16
	PostgresUsername       string
	PostgresPassword       string
	PostgresDatabase       string
	PostgresSslMode        string
	PostgresMaxConnections int32
	PostgresConnectTimeout time.Duration

	TokenAccessSecret    string
	TokenRefreshSecret   string
	TokenAccessDuration  time.Duration
	TokenRefreshDuration time.Duration

	PasswordBcryptCost int
	ControlSpaceSlug   string
}
