package config_test

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MateWorkspace/lockmate/internal/config"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
)

var initialEnv = config.Env

func setupEnvironment(t *testing.T) {
	t.Helper()
	previous := config.Env
	config.Env = initialEnv
	t.Cleanup(func() { config.Env = previous })
	t.Chdir(t.TempDir())

	for _, key := range []string{
		"LOCKMATE_HTTP_PUBLIC_URL", "LOCKMATE_GOOGLE_OAUTH_ENABLED", "LOCKMATE_GOOGLE_OAUTH_CLIENT_ID", "LOCKMATE_GOOGLE_OAUTH_CLIENT_SECRET", "LOCKMATE_GOOGLE_OAUTH_LAUNCH_DURATION", "LOCKMATE_GOOGLE_OAUTH_ATTEMPT_DURATION", "LOCKMATE_AUTH_HANDOFF_DURATION", "LOCKMATE_GOOGLE_OAUTH_HTTP_TIMEOUT", "LOCKMATE_AUTH_FLOW_CLEANUP_INTERVAL",
		"LOCKMATE_LOGGER_FORMAT", "LOCKMATE_LOGGER_LEVEL", "LOCKMATE_LOGGER_USE_TRACE",
		"LOCKMATE_PORT_GRPC_SERVER", "LOCKMATE_PORT_HTTP_SERVER",
		"LOCKMATE_HTTP_SERVER_READ_HEADER_TIMEOUT", "LOCKMATE_HTTP_SERVER_READ_TIMEOUT",
		"LOCKMATE_HTTP_SERVER_WRITE_TIMEOUT", "LOCKMATE_HTTP_SERVER_IDLE_TIMEOUT", "LOCKMATE_HTTP_SERVER_SHUTDOWN_TIMEOUT",
		"LOCKMATE_POSTGRES_HOST", "LOCKMATE_POSTGRES_PORT", "LOCKMATE_POSTGRES_USERNAME",
		"LOCKMATE_POSTGRES_PASSWORD", "LOCKMATE_POSTGRES_DATABASE", "LOCKMATE_POSTGRES_SSL_MODE",
		"LOCKMATE_POSTGRES_MAX_CONNECTIONS", "LOCKMATE_POSTGRES_CONNECT_TIMEOUT",
		"LOCKMATE_TOKEN_ACCESS_SECRET", "LOCKMATE_TOKEN_REFRESH_SECRET",
		"LOCKMATE_TOKEN_ACCESS_DURATION", "LOCKMATE_TOKEN_REFRESH_DURATION",
		"LOCKMATE_PASSWORD_BCRYPT_COST", "LOCKMATE_CONTROL_SPACE_SLUG",
	} {
		t.Setenv(key, "")
	}

	t.Setenv("LOCKMATE_TOKEN_ACCESS_SECRET", "test-access-secret")
	t.Setenv("LOCKMATE_TOKEN_REFRESH_SECRET", "test-refresh-secret")
}

func TestLoadEnvDefaultsAndOverrides(t *testing.T) {
	setupEnvironment(t)
	if err := config.LoadEnv(); err != nil {
		t.Fatal(err)
	}

	if config.Env.LoggerFormat != domainmodelslogger.FormatJson || config.Env.LoggerLevel != domainmodelslogger.LevelInfo || config.Env.PortGrpcServer != 50051 || config.Env.PortHttpServer != 10701 || config.Env.PostgresMaxConnections != 20 {
		t.Fatal("incorrect initial defaults")
	}

	t.Setenv("LOCKMATE_LOGGER_FORMAT", " PLAIN ")
	t.Setenv("LOCKMATE_LOGGER_LEVEL", "debug")
	t.Setenv("LOCKMATE_LOGGER_USE_TRACE", "false")
	t.Setenv("LOCKMATE_PORT_HTTP_SERVER", "8000")
	t.Setenv("LOCKMATE_POSTGRES_MAX_CONNECTIONS", "30")
	t.Setenv("LOCKMATE_POSTGRES_CONNECT_TIMEOUT", "1500ms")
	t.Setenv("LOCKMATE_TOKEN_ACCESS_DURATION", "30")
	t.Setenv("LOCKMATE_PASSWORD_BCRYPT_COST", "12")
	if err := config.LoadEnv(); err != nil {
		t.Fatal(err)
	}

	if config.Env.LoggerFormat != domainmodelslogger.FormatPlain || config.Env.LoggerLevel != domainmodelslogger.LevelDebug || config.Env.LoggerUseTrace || config.Env.PortHttpServer != 8000 || config.Env.PostgresMaxConnections != 30 || config.Env.PostgresConnectTimeout != 1500*time.Millisecond || config.Env.TokenAccessDuration != 30*time.Second || config.Env.PasswordBcryptCost != 12 {
		t.Fatal("environment overrides ignored")
	}
}

func TestTypedParsingFallsBackWithoutOverflow(t *testing.T) {
	setupEnvironment(t)
	t.Setenv("LOCKMATE_PORT_GRPC_SERVER", "65536")
	t.Setenv("LOCKMATE_PORT_HTTP_SERVER", "-1")
	t.Setenv("LOCKMATE_POSTGRES_MAX_CONNECTIONS", "2147483648")
	t.Setenv("LOCKMATE_TOKEN_ACCESS_DURATION", "9223372036854775807")
	t.Setenv("LOCKMATE_LOGGER_FORMAT", "unsupported")
	t.Setenv("LOCKMATE_LOGGER_LEVEL", "unsupported")
	t.Setenv("LOCKMATE_LOGGER_USE_TRACE", "malformed")
	t.Setenv("LOCKMATE_PASSWORD_BCRYPT_COST", "malformed")
	if err := config.LoadEnv(); err != nil {
		t.Fatal(err)
	}

	if config.Env.PortGrpcServer != 50051 || config.Env.PortHttpServer != 10701 || config.Env.PostgresMaxConnections != 20 || config.Env.TokenAccessDuration != 15*time.Minute || config.Env.LoggerFormat != domainmodelslogger.FormatJson || config.Env.LoggerLevel != domainmodelslogger.LevelInfo || !config.Env.LoggerUseTrace || config.Env.PasswordBcryptCost != 10 {
		t.Fatal("malformed or overflowing values did not use fallback")
	}
}

func TestLoadEnvCallsValidate(t *testing.T) {
	for _, test := range []struct {
		key, value string
		kind       error
	}{
		{"LOCKMATE_TOKEN_ACCESS_SECRET", "", domainmodelserror.EnvironmentRequired},
		{"LOCKMATE_TOKEN_REFRESH_SECRET", " ", domainmodelserror.EnvironmentRequired},
		{"LOCKMATE_TOKEN_ACCESS_DURATION", "999ms", domainmodelserror.EnvironmentInvalid},
		{"LOCKMATE_TOKEN_REFRESH_DURATION", "0", domainmodelserror.EnvironmentInvalid},
		{"LOCKMATE_PORT_GRPC_SERVER", "0", domainmodelserror.EnvironmentInvalid},
		{"LOCKMATE_PORT_HTTP_SERVER", "0", domainmodelserror.EnvironmentInvalid},
		{"LOCKMATE_POSTGRES_PORT", "0", domainmodelserror.EnvironmentInvalid},
		{"LOCKMATE_POSTGRES_MAX_CONNECTIONS", "-1", domainmodelserror.EnvironmentInvalid},
		{"LOCKMATE_POSTGRES_CONNECT_TIMEOUT", "0", domainmodelserror.EnvironmentInvalid},
		{"LOCKMATE_PASSWORD_BCRYPT_COST", "32", domainmodelserror.EnvironmentInvalid},
	} {
		t.Run(test.key, func(t *testing.T) {
			setupEnvironment(t)
			t.Setenv(test.key, test.value)
			err := config.LoadEnv()
			if !errors.Is(err, test.kind) || !strings.Contains(err.Error(), test.key) {
				t.Fatalf("incorrect validation error: %v", err)
			}
		})
	}
}

func TestLoadEnvUpdatesGlobalBeforeValidation(t *testing.T) {
	setupEnvironment(t)
	t.Setenv("LOCKMATE_PORT_HTTP_SERVER", "8000")
	t.Setenv("LOCKMATE_PASSWORD_BCRYPT_COST", "3")
	if err := config.LoadEnv(); !errors.Is(err, domainmodelserror.EnvironmentInvalid) {
		t.Fatalf("missing validation: %v", err)
	}

	if config.Env.PortHttpServer != 8000 || config.Env.PasswordBcryptCost != 3 {
		t.Fatal("LoadEnv did not populate global fields before validation")
	}
}

func TestExistingGlobalValuesAreFallbacks(t *testing.T) {
	setupEnvironment(t)
	config.Env.PostgresHost = "configured-db"
	config.Env.PortHttpServer = 8001
	t.Setenv("LOCKMATE_PORT_HTTP_SERVER", "malformed")
	if err := config.LoadEnv(); err != nil {
		t.Fatal(err)
	}

	if config.Env.PostgresHost != "configured-db" || config.Env.PortHttpServer != 8001 {
		t.Fatal("fallbacks did not use existing global values")
	}
}

func TestOptionalDotenvAndProcessPrecedence(t *testing.T) {
	setupEnvironment(t)
	if err := os.Unsetenv("LOCKMATE_POSTGRES_HOST"); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(".env", []byte("LOCKMATE_POSTGRES_HOST=file-db\nLOCKMATE_PORT_HTTP_SERVER=9000\n"), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("LOCKMATE_PORT_HTTP_SERVER", "8002")
	if err := config.LoadEnv(); err != nil {
		t.Fatal(err)
	}

	if config.Env.PostgresHost != "file-db" || os.Getenv("LOCKMATE_POSTGRES_HOST") != "file-db" || config.Env.PortHttpServer != 8002 {
		t.Fatal("dotenv loading or process precedence incorrect")
	}
}

func TestMalformedDotenvReturnsDomainError(t *testing.T) {
	setupEnvironment(t)
	if err := os.WriteFile(".env", []byte("LOCKMATE_TOKEN_ACCESS_SECRET=\"private-unclosed-secret\n"), 0600); err != nil {
		t.Fatal(err)
	}

	err := config.LoadEnv()
	if !errors.Is(err, domainmodelserror.EnvironmentInvalid) || strings.Contains(err.Error(), "private-unclosed-secret") {
		t.Fatalf("incorrect dotenv error: %v", err)
	}
}

func TestHTTPTimeoutEnvironment(t *testing.T) {
	setupEnvironment(t)
	if err := config.LoadEnv(); err != nil {
		t.Fatal(err)
	}
	if config.Env.HttpServerReadHeaderTimeout != 5*time.Second || config.Env.HttpServerReadTimeout != 15*time.Second || config.Env.HttpServerWriteTimeout != 30*time.Second || config.Env.HttpServerIdleTimeout != 60*time.Second || config.Env.HttpServerShutdownTimeout != 10*time.Second {
		t.Fatal("incorrect HTTP defaults")
	}
	t.Setenv("LOCKMATE_HTTP_SERVER_READ_HEADER_TIMEOUT", "2s")
	t.Setenv("LOCKMATE_HTTP_SERVER_READ_TIMEOUT", "20")
	t.Setenv("LOCKMATE_HTTP_SERVER_WRITE_TIMEOUT", "40s")
	t.Setenv("LOCKMATE_HTTP_SERVER_IDLE_TIMEOUT", "120s")
	t.Setenv("LOCKMATE_HTTP_SERVER_SHUTDOWN_TIMEOUT", "30s")
	if err := config.LoadEnv(); err != nil {
		t.Fatal(err)
	}
	if config.Env.HttpServerReadHeaderTimeout != 2*time.Second || config.Env.HttpServerReadTimeout != 20*time.Second || config.Env.HttpServerWriteTimeout != 40*time.Second || config.Env.HttpServerIdleTimeout != 120*time.Second || config.Env.HttpServerShutdownTimeout != 30*time.Second {
		t.Fatal("HTTP overrides ignored")
	}
	t.Setenv("LOCKMATE_HTTP_SERVER_SHUTDOWN_TIMEOUT", "0s")
	if err := config.LoadEnv(); !errors.Is(err, domainmodelserror.EnvironmentInvalid) {
		t.Fatalf("invalid shutdown timeout accepted: %v", err)
	}
}

func TestGoogleEnvironment(t *testing.T) {
	for _, test := range []struct {
		name, key, value string
		kind             error
	}{
		{"missing client", "LOCKMATE_GOOGLE_OAUTH_CLIENT_ID", "", domainmodelserror.EnvironmentRequired},
		{"missing secret", "LOCKMATE_GOOGLE_OAUTH_CLIENT_SECRET", "", domainmodelserror.EnvironmentRequired},
		{"insecure public origin", "LOCKMATE_HTTP_PUBLIC_URL", "http://lockmate-dev.dhonan.my.id", domainmodelserror.EnvironmentInvalid},
		{"uppercase host", "LOCKMATE_HTTP_PUBLIC_URL", "https://LOCKMATE-dev.dhonan.my.id", domainmodelserror.EnvironmentInvalid},
		{"invalid hostname", "LOCKMATE_HTTP_PUBLIC_URL", "https://bad..example.com", domainmodelserror.EnvironmentInvalid},
		{"invalid port", "LOCKMATE_HTTP_PUBLIC_URL", "https://example.com:99999", domainmodelserror.EnvironmentInvalid},
		{"public URL path", "LOCKMATE_HTTP_PUBLIC_URL", "https://lockmate-dev.dhonan.my.id/", domainmodelserror.EnvironmentInvalid},
		{"public URL query", "LOCKMATE_HTTP_PUBLIC_URL", "https://lockmate-dev.dhonan.my.id?", domainmodelserror.EnvironmentInvalid},
		{"public URL credentials", "LOCKMATE_HTTP_PUBLIC_URL", "https://private-secret@lockmate-dev.dhonan.my.id", domainmodelserror.EnvironmentInvalid},
		{"duration", "LOCKMATE_AUTH_HANDOFF_DURATION", "0s", domainmodelserror.EnvironmentInvalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			setupEnvironment(t)
			t.Setenv("LOCKMATE_GOOGLE_OAUTH_ENABLED", "true")
			t.Setenv("LOCKMATE_GOOGLE_OAUTH_CLIENT_ID", "client")
			t.Setenv("LOCKMATE_GOOGLE_OAUTH_CLIENT_SECRET", "private-secret")
			t.Setenv(test.key, test.value)
			err := config.LoadEnv()
			if !errors.Is(err, test.kind) || strings.Contains(err.Error(), "private-secret") {
				t.Fatal(err)
			}
		})
	}
	setupEnvironment(t)
	t.Setenv("LOCKMATE_GOOGLE_OAUTH_ENABLED", "true")
	t.Setenv("LOCKMATE_GOOGLE_OAUTH_CLIENT_ID", "client")
	t.Setenv("LOCKMATE_GOOGLE_OAUTH_CLIENT_SECRET", "secret")
	if err := config.LoadEnv(); err != nil {
		t.Fatal(err)
	}
	if config.Env.PortHttpServer != 10701 || config.Env.HttpPublicUrl != "https://lockmate-dev.dhonan.my.id" || config.Env.GoogleOAuthAttemptDuration != 10*time.Minute || config.Env.AuthHandoffDuration != time.Minute {
		t.Fatal("incorrect Google defaults")
	}
}
