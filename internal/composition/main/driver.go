package compositionmain

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/MateWorkspace/lockmate/internal/config"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type driver struct {
	slogLogger   *slog.Logger
	gin          *gin.Engine
	httpServer   *http.Server
	postgresPool *pgxpool.Pool
}

func (l *launcher) newDriver(ctx context.Context) error {
	const tag = path + "/driver"

	options := &slog.HandlerOptions{Level: slog.LevelDebug}
	var handler slog.Handler
	if config.Env.LoggerFormat == domainmodelslogger.FormatJson {
		handler = slog.NewJSONHandler(os.Stdout, options)
	} else {
		handler = slog.NewTextHandler(os.Stdout, options)
	}
	logger := slog.New(handler)

	pool, err := newPostgresPool(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "failed to initialize PostgreSQL", "tag", tag, "err", err)
		return err
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	if err := engine.SetTrustedProxies(nil); err != nil {
		logger.ErrorContext(ctx, "failed to configure HTTP trusted proxies", "tag", tag, "err", err)
		pool.Close()
		return err
	}
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", config.Env.PortHttpServer), Handler: engine,
		ReadHeaderTimeout: config.Env.HttpServerReadHeaderTimeout,
		ReadTimeout:       config.Env.HttpServerReadTimeout,
		WriteTimeout:      config.Env.HttpServerWriteTimeout,
		IdleTimeout:       config.Env.HttpServerIdleTimeout,
	}

	l.drv = &driver{
		slogLogger:   logger,
		gin:          engine,
		httpServer:   server,
		postgresPool: pool,
	}

	return nil
}

func newPostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	address := &url.URL{
		Scheme: "postgres", User: url.UserPassword(config.Env.PostgresUsername, config.Env.PostgresPassword),
		Host: net.JoinHostPort(config.Env.PostgresHost, strconv.Itoa(int(config.Env.PostgresPort))),
		Path: "/" + config.Env.PostgresDatabase,
	}

	query := address.Query()
	query.Set("sslmode", config.Env.PostgresSslMode)
	address.RawQuery = query.Encode()
	cfg, err := pgxpool.ParseConfig(address.String())
	if err != nil {
		// Parser errors can contain the credential-bearing URL.
		return nil, domainmodelserror.NewError("PostgreSQL connection settings are invalid", domainmodelserror.EnvironmentInvalid, nil)
	}

	cfg.MaxConns = config.Env.PostgresMaxConnections
	cfg.ConnConfig.ConnectTimeout = config.Env.PostgresConnectTimeout
	connectCtx, cancel := context.WithTimeout(ctx, config.Env.PostgresConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func (d *driver) cleanup() {
	if d.postgresPool != nil {
		d.postgresPool.Close()
	}
}
