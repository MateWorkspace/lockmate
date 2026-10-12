package domainmodelsapp

import (
	"net/netip"
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

	TokenRefreshSecret   string
	TokenAccessDuration  time.Duration
	TokenRefreshDuration time.Duration

	SigningKeyEncryptionKey         []byte
	SigningKeyEncryptionKeyPrevious []byte
	SigningKeyRotationInterval      time.Duration
	SigningKeyRetentionInterval     time.Duration
	SigningKeyReconcileInterval     time.Duration

	SeederEnabled         bool
	SeederSuperPassword   string
	SeederAdminPassword   string
	SeederUserPassword    string
	SeederClientSecretKey string

	PasswordBcryptCost int
	HttpPublicUrl      string

	ClientIpHeader string
	TrustedProxies []netip.Prefix

	GoogleOAuthEnabled         bool
	GoogleOAuthClientId        string
	GoogleOAuthClientSecret    string
	GoogleOAuthLaunchDuration  time.Duration
	GoogleOAuthAttemptDuration time.Duration
	GoogleOAuthHttpTimeout     time.Duration

	AuthRedirectAllowHttp bool

	AuthHandoffDuration     time.Duration
	AuthFlowCleanupInterval time.Duration
}
