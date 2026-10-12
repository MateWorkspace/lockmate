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

	HttpServerReadHeaderTimeout time.Duration
	HttpServerReadTimeout       time.Duration
	HttpServerWriteTimeout      time.Duration
	HttpServerIdleTimeout       time.Duration
	HttpServerShutdownTimeout   time.Duration

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
	HttpPublicUrl      string

	GoogleOAuthEnabled         bool
	GoogleOAuthClientId        string
	GoogleOAuthClientSecret    string
	GoogleOAuthLaunchDuration  time.Duration
	GoogleOAuthAttemptDuration time.Duration
	GoogleOAuthHttpTimeout     time.Duration

	AuthHandoffDuration     time.Duration
	AuthFlowCleanupInterval time.Duration
}
