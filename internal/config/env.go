package config

import (
	"encoding/base64"
	"math"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var Env = domainmodelsapp.Env{
	LoggerFormat:   domainmodelslogger.FormatJson,
	LoggerLevel:    domainmodelslogger.LevelInfo,
	LoggerUseTrace: true,

	PortGrpcServer: 50051,
	PortHttpServer: 10701,

	HttpServerReadHeaderTimeout: 5 * time.Second,
	HttpServerReadTimeout:       15 * time.Second,
	HttpServerWriteTimeout:      30 * time.Second,
	HttpServerIdleTimeout:       60 * time.Second,
	HttpServerShutdownTimeout:   10 * time.Second,

	PostgresHost:           "127.0.0.1",
	PostgresPort:           5432,
	PostgresUsername:       "postgres",
	PostgresPassword:       "postgres",
	PostgresDatabase:       "lockmate",
	PostgresSslMode:        "disable",
	PostgresMaxConnections: 20,
	PostgresConnectTimeout: 10 * time.Second,

	TokenAccessDuration:  15 * time.Minute,
	TokenRefreshDuration: 24 * time.Hour,

	SigningKeyRotationInterval:  720 * time.Hour,
	SigningKeyRetentionInterval: time.Hour,
	SigningKeyReconcileInterval: time.Hour,

	SeederEnabled:      true,
	PasswordBcryptCost: bcrypt.DefaultCost,
	HttpPublicUrl:      "https://lockmate-dev.dhonan.my.id",

	GoogleOAuthEnabled:         false,
	GoogleOAuthClientId:        "",
	GoogleOAuthClientSecret:    "",
	GoogleOAuthLaunchDuration:  2 * time.Minute,
	GoogleOAuthAttemptDuration: 10 * time.Minute,
	GoogleOAuthHttpTimeout:     10 * time.Second,
	AuthHandoffDuration:        time.Minute,
	AuthFlowCleanupInterval:    5 * time.Minute,
}

func LoadEnv() error {
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		return domainmodelserror.NewError(".env could not be read or contains an invalid entry", domainmodelserror.EnvironmentInvalid, nil)
	}

	Env.LoggerFormat = envGetLoggerFormat("LOCKMATE_LOGGER_FORMAT", Env.LoggerFormat)
	Env.LoggerLevel = envGetLoggerLevel("LOCKMATE_LOGGER_LEVEL", Env.LoggerLevel)
	Env.LoggerUseTrace = envGetBool("LOCKMATE_LOGGER_USE_TRACE", Env.LoggerUseTrace)

	Env.PortGrpcServer = envGetUint16("LOCKMATE_PORT_GRPC_SERVER", Env.PortGrpcServer)
	Env.PortHttpServer = envGetUint16("LOCKMATE_PORT_HTTP_SERVER", Env.PortHttpServer)

	Env.HttpServerReadHeaderTimeout = envGetDuration("LOCKMATE_HTTP_SERVER_READ_HEADER_TIMEOUT", Env.HttpServerReadHeaderTimeout)
	Env.HttpServerReadTimeout = envGetDuration("LOCKMATE_HTTP_SERVER_READ_TIMEOUT", Env.HttpServerReadTimeout)
	Env.HttpServerWriteTimeout = envGetDuration("LOCKMATE_HTTP_SERVER_WRITE_TIMEOUT", Env.HttpServerWriteTimeout)
	Env.HttpServerIdleTimeout = envGetDuration("LOCKMATE_HTTP_SERVER_IDLE_TIMEOUT", Env.HttpServerIdleTimeout)
	Env.HttpServerShutdownTimeout = envGetDuration("LOCKMATE_HTTP_SERVER_SHUTDOWN_TIMEOUT", Env.HttpServerShutdownTimeout)

	Env.PostgresHost = envGetString("LOCKMATE_POSTGRES_HOST", Env.PostgresHost)
	Env.PostgresPort = envGetUint16("LOCKMATE_POSTGRES_PORT", Env.PostgresPort)
	Env.PostgresUsername = envGetString("LOCKMATE_POSTGRES_USERNAME", Env.PostgresUsername)
	Env.PostgresPassword = envGetString("LOCKMATE_POSTGRES_PASSWORD", Env.PostgresPassword)
	Env.PostgresDatabase = envGetString("LOCKMATE_POSTGRES_DATABASE", Env.PostgresDatabase)
	Env.PostgresSslMode = envGetString("LOCKMATE_POSTGRES_SSL_MODE", Env.PostgresSslMode)
	Env.PostgresMaxConnections = envGetInt32("LOCKMATE_POSTGRES_MAX_CONNECTIONS", Env.PostgresMaxConnections)
	Env.PostgresConnectTimeout = envGetDuration("LOCKMATE_POSTGRES_CONNECT_TIMEOUT", Env.PostgresConnectTimeout)

	Env.TokenRefreshSecret = envGetString("LOCKMATE_TOKEN_REFRESH_SECRET", Env.TokenRefreshSecret)
	Env.TokenAccessDuration = envGetDuration("LOCKMATE_TOKEN_ACCESS_DURATION", Env.TokenAccessDuration)
	Env.TokenRefreshDuration = envGetDuration("LOCKMATE_TOKEN_REFRESH_DURATION", Env.TokenRefreshDuration)

	Env.SigningKeyEncryptionKey = envGetBase64("LOCKMATE_SIGNING_KEY_ENCRYPTION_KEY", Env.SigningKeyEncryptionKey)
	Env.SigningKeyEncryptionKeyPrevious = envGetBase64("LOCKMATE_SIGNING_KEY_ENCRYPTION_KEY_PREVIOUS", Env.SigningKeyEncryptionKeyPrevious)
	Env.SigningKeyRotationInterval = envGetDuration("LOCKMATE_SIGNING_KEY_ROTATION_INTERVAL", Env.SigningKeyRotationInterval)
	Env.SigningKeyRetentionInterval = envGetDuration("LOCKMATE_SIGNING_KEY_RETENTION_INTERVAL", Env.SigningKeyRetentionInterval)
	Env.SigningKeyReconcileInterval = envGetDuration("LOCKMATE_SIGNING_KEY_RECONCILE_INTERVAL", Env.SigningKeyReconcileInterval)

	Env.PasswordBcryptCost = envGetInt("LOCKMATE_PASSWORD_BCRYPT_COST", Env.PasswordBcryptCost)
	Env.SeederEnabled = envGetBool("LOCKMATE_SEEDER_ENABLED", Env.SeederEnabled)
	Env.SeederSuperPassword = envGetString("LOCKMATE_SEEDER_SUPER_PASSWORD", Env.SeederSuperPassword)
	Env.SeederAdminPassword = envGetString("LOCKMATE_SEEDER_ADMIN_PASSWORD", Env.SeederAdminPassword)
	Env.SeederUserPassword = envGetString("LOCKMATE_SEEDER_USER_PASSWORD", Env.SeederUserPassword)
	Env.SeederClientSecretKey = envGetString("LOCKMATE_SEEDER_CLIENT_SECRET_KEY", Env.SeederClientSecretKey)
	Env.HttpPublicUrl = envGetString("LOCKMATE_HTTP_PUBLIC_URL", Env.HttpPublicUrl)

	Env.GoogleOAuthEnabled = envGetBool("LOCKMATE_GOOGLE_OAUTH_ENABLED", Env.GoogleOAuthEnabled)
	Env.GoogleOAuthClientId = envGetString("LOCKMATE_GOOGLE_OAUTH_CLIENT_ID", Env.GoogleOAuthClientId)
	Env.GoogleOAuthClientSecret = envGetString("LOCKMATE_GOOGLE_OAUTH_CLIENT_SECRET", Env.GoogleOAuthClientSecret)
	Env.GoogleOAuthLaunchDuration = envGetDuration("LOCKMATE_GOOGLE_OAUTH_LAUNCH_DURATION", Env.GoogleOAuthLaunchDuration)
	Env.GoogleOAuthAttemptDuration = envGetDuration("LOCKMATE_GOOGLE_OAUTH_ATTEMPT_DURATION", Env.GoogleOAuthAttemptDuration)
	Env.AuthRedirectAllowHttp = envGetBool("LOCKMATE_AUTH_REDIRECT_ALLOW_HTTP", Env.AuthRedirectAllowHttp)
	Env.AuthHandoffDuration = envGetDuration("LOCKMATE_AUTH_HANDOFF_DURATION", Env.AuthHandoffDuration)

	Env.GoogleOAuthHttpTimeout = envGetDuration("LOCKMATE_GOOGLE_OAUTH_HTTP_TIMEOUT", Env.GoogleOAuthHttpTimeout)
	Env.AuthFlowCleanupInterval = envGetDuration("LOCKMATE_AUTH_FLOW_CLEANUP_INTERVAL", Env.AuthFlowCleanupInterval)

	return validate()
}

func validate() error {
	if Env.HttpServerReadHeaderTimeout < time.Second {
		return domainmodelserror.NewError("LOCKMATE_HTTP_SERVER_READ_HEADER_TIMEOUT must be at least one second", domainmodelserror.EnvironmentInvalid, nil)
	}
	if Env.HttpServerReadTimeout < time.Second {
		return domainmodelserror.NewError("LOCKMATE_HTTP_SERVER_READ_TIMEOUT must be at least one second", domainmodelserror.EnvironmentInvalid, nil)
	}
	if Env.HttpServerWriteTimeout < time.Second {
		return domainmodelserror.NewError("LOCKMATE_HTTP_SERVER_WRITE_TIMEOUT must be at least one second", domainmodelserror.EnvironmentInvalid, nil)
	}
	if Env.HttpServerIdleTimeout < time.Second {
		return domainmodelserror.NewError("LOCKMATE_HTTP_SERVER_IDLE_TIMEOUT must be at least one second", domainmodelserror.EnvironmentInvalid, nil)
	}
	if Env.HttpServerShutdownTimeout < time.Second {
		return domainmodelserror.NewError("LOCKMATE_HTTP_SERVER_SHUTDOWN_TIMEOUT must be at least one second", domainmodelserror.EnvironmentInvalid, nil)
	}

	if len(Env.SigningKeyEncryptionKey) != 32 {
		return domainmodelserror.NewError("LOCKMATE_SIGNING_KEY_ENCRYPTION_KEY must be 32 base64-encoded bytes", domainmodelserror.EnvironmentRequired, nil)
	}

	if len(Env.SigningKeyEncryptionKeyPrevious) != 0 && len(Env.SigningKeyEncryptionKeyPrevious) != 32 {
		return domainmodelserror.NewError("LOCKMATE_SIGNING_KEY_ENCRYPTION_KEY_PREVIOUS must be 32 base64-encoded bytes", domainmodelserror.EnvironmentInvalid, nil)
	}

	if Env.SigningKeyRotationInterval < time.Second {
		return domainmodelserror.NewError("LOCKMATE_SIGNING_KEY_ROTATION_INTERVAL must be at least one second", domainmodelserror.EnvironmentInvalid, nil)
	}

	if Env.SigningKeyRetentionInterval < time.Second {
		return domainmodelserror.NewError("LOCKMATE_SIGNING_KEY_RETENTION_INTERVAL must be at least one second", domainmodelserror.EnvironmentInvalid, nil)
	}

	if Env.SigningKeyRetentionInterval < Env.TokenAccessDuration {
		return domainmodelserror.NewError("LOCKMATE_SIGNING_KEY_RETENTION_INTERVAL must be at least the access-token duration", domainmodelserror.EnvironmentInvalid, nil)
	}

	if Env.SigningKeyReconcileInterval < time.Second {
		return domainmodelserror.NewError("LOCKMATE_SIGNING_KEY_RECONCILE_INTERVAL must be at least one second", domainmodelserror.EnvironmentInvalid, nil)
	}

	if Env.SigningKeyReconcileInterval > Env.SigningKeyRotationInterval {
		return domainmodelserror.NewError("LOCKMATE_SIGNING_KEY_RECONCILE_INTERVAL must not exceed LOCKMATE_SIGNING_KEY_ROTATION_INTERVAL", domainmodelserror.EnvironmentInvalid, nil)
	}

	if strings.TrimSpace(Env.TokenRefreshSecret) == "" {
		return domainmodelserror.NewError("LOCKMATE_TOKEN_REFRESH_SECRET is required", domainmodelserror.EnvironmentRequired, nil)
	}

	if Env.TokenAccessDuration < time.Second {
		return domainmodelserror.NewError("LOCKMATE_TOKEN_ACCESS_DURATION must be at least one second", domainmodelserror.EnvironmentInvalid, nil)
	}

	if Env.TokenRefreshDuration < time.Second {
		return domainmodelserror.NewError("LOCKMATE_TOKEN_REFRESH_DURATION must be at least one second", domainmodelserror.EnvironmentInvalid, nil)
	}

	if Env.PortGrpcServer == 0 {
		return domainmodelserror.NewError("LOCKMATE_PORT_GRPC_SERVER must be between 1 and 65535", domainmodelserror.EnvironmentInvalid, nil)
	}

	if Env.PortHttpServer == 0 {
		return domainmodelserror.NewError("LOCKMATE_PORT_HTTP_SERVER must be between 1 and 65535", domainmodelserror.EnvironmentInvalid, nil)
	}

	if Env.PostgresPort == 0 {
		return domainmodelserror.NewError("LOCKMATE_POSTGRES_PORT must be between 1 and 65535", domainmodelserror.EnvironmentInvalid, nil)
	}

	if Env.PostgresMaxConnections <= 0 {
		return domainmodelserror.NewError("LOCKMATE_POSTGRES_MAX_CONNECTIONS must be positive", domainmodelserror.EnvironmentInvalid, nil)
	}

	if Env.PostgresConnectTimeout <= 0 {
		return domainmodelserror.NewError("LOCKMATE_POSTGRES_CONNECT_TIMEOUT must be positive", domainmodelserror.EnvironmentInvalid, nil)
	}

	if Env.PasswordBcryptCost < bcrypt.MinCost || Env.PasswordBcryptCost > bcrypt.MaxCost {
		return domainmodelserror.NewError("LOCKMATE_PASSWORD_BCRYPT_COST must be between 4 and 31", domainmodelserror.EnvironmentInvalid, nil)
	}

	if Env.SeederEnabled {
		if strings.TrimSpace(Env.SeederSuperPassword) == "" {
			return domainmodelserror.NewError("LOCKMATE_SEEDER_SUPER_PASSWORD is required when seeding is enabled", domainmodelserror.EnvironmentRequired, nil)
		}

		if strings.TrimSpace(Env.SeederAdminPassword) == "" {
			return domainmodelserror.NewError("LOCKMATE_SEEDER_ADMIN_PASSWORD is required when seeding is enabled", domainmodelserror.EnvironmentRequired, nil)
		}

		if strings.TrimSpace(Env.SeederUserPassword) == "" {
			return domainmodelserror.NewError("LOCKMATE_SEEDER_USER_PASSWORD is required when seeding is enabled", domainmodelserror.EnvironmentRequired, nil)
		}

		if strings.TrimSpace(Env.SeederClientSecretKey) == "" {
			return domainmodelserror.NewError("LOCKMATE_SEEDER_CLIENT_SECRET_KEY is required when seeding is enabled", domainmodelserror.EnvironmentRequired, nil)
		}

	}

	if Env.GoogleOAuthEnabled {
		if strings.TrimSpace(Env.GoogleOAuthClientId) == "" {
			return domainmodelserror.NewError("LOCKMATE_GOOGLE_OAUTH_CLIENT_ID is required", domainmodelserror.EnvironmentRequired, nil)
		}
		if strings.TrimSpace(Env.GoogleOAuthClientSecret) == "" {
			return domainmodelserror.NewError("LOCKMATE_GOOGLE_OAUTH_CLIENT_SECRET is required", domainmodelserror.EnvironmentRequired, nil)
		}
		u, err := url.Parse(Env.HttpPublicUrl)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.String() != Env.HttpPublicUrl || !validPublicHost(u) || strings.ContainsAny(Env.HttpPublicUrl, " \t\r\n?#") {
			return domainmodelserror.NewError("LOCKMATE_HTTP_PUBLIC_URL must be an HTTPS origin", domainmodelserror.EnvironmentInvalid, nil)
		}
	}

	if Env.GoogleOAuthLaunchDuration < time.Second {
		return domainmodelserror.NewError("LOCKMATE_GOOGLE_OAUTH_LAUNCH_DURATION must be at least one second", domainmodelserror.EnvironmentInvalid, nil)
	}

	if Env.GoogleOAuthAttemptDuration < time.Second {
		return domainmodelserror.NewError("LOCKMATE_GOOGLE_OAUTH_ATTEMPT_DURATION must be at least one second", domainmodelserror.EnvironmentInvalid, nil)
	}

	if Env.GoogleOAuthHttpTimeout < time.Second {
		return domainmodelserror.NewError("LOCKMATE_GOOGLE_OAUTH_HTTP_TIMEOUT must be at least one second", domainmodelserror.EnvironmentInvalid, nil)
	}

	if Env.AuthHandoffDuration < time.Second {
		return domainmodelserror.NewError("LOCKMATE_AUTH_HANDOFF_DURATION must be at least one second", domainmodelserror.EnvironmentInvalid, nil)
	}

	if Env.AuthFlowCleanupInterval < time.Second {
		return domainmodelserror.NewError("LOCKMATE_AUTH_FLOW_CLEANUP_INTERVAL must be at least one second", domainmodelserror.EnvironmentInvalid, nil)
	}

	return nil
}

func envGetString(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envGetBase64(key string, fallback []byte) []byte {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return fallback
	}

	return decoded
}

func envGetStrings(key string, fallback []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	values := strings.Split(value, ",")
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		result = append(result, value)
	}
	if len(result) == 0 {
		return fallback
	}

	return result
}

func envGetInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	valInt, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return valInt
}

func envGetUint16(key string, fallback uint16) uint16 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	valUint16, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return fallback
	}

	return uint16(valUint16)
}

func envGetInt32(key string, fallback int32) int32 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	valInt32, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return fallback
	}

	return int32(valInt32)
}

func envGetBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	valBool, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return valBool
}

func envGetDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err == nil {
		return duration
	}

	seconds, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	if int64(seconds) > math.MaxInt64/int64(time.Second) || int64(seconds) < math.MinInt64/int64(time.Second) {
		return fallback
	}

	return time.Duration(seconds) * time.Second
}

func envGetLoggerFormat(key string, fallback domainmodelslogger.Format) domainmodelslogger.Format {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "plain":
		return domainmodelslogger.FormatPlain
	case "json":
		return domainmodelslogger.FormatJson
	default:
		return fallback
	}
}

func envGetLoggerLevel(key string, fallback domainmodelslogger.Level) domainmodelslogger.Level {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "none":
		return domainmodelslogger.LevelNone
	case "error":
		return domainmodelslogger.LevelError
	case "warn":
		return domainmodelslogger.LevelWarn
	case "info":
		return domainmodelslogger.LevelInfo
	case "debug":
		return domainmodelslogger.LevelDebug
	default:
		return fallback
	}
}

func validPublicHost(value *url.URL) bool {
	if value.Host != strings.ToLower(value.Host) || value.Hostname() == "" {
		return false
	}
	host := value.Hostname()
	if address, err := netip.ParseAddr(host); err == nil {
		if address.Zone() != "" || address.String() != host {
			return false
		}
	} else {
		if len(host) > 253 {
			return false
		}
		for _, label := range strings.Split(host, ".") {
			if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
				return false
			}
			for _, char := range label {
				if !(char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || char == '-') {
					return false
				}
			}
		}
	}
	if value.Port() != "" {
		port, err := strconv.ParseUint(value.Port(), 10, 16)
		if err != nil || port == 0 || port == 443 || strconv.FormatUint(port, 10) != value.Port() {
			return false
		}
	} else if strings.HasSuffix(value.Host, ":") {
		return false
	}
	return true
}
