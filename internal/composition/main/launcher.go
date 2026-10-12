package compositionmain

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MateWorkspace/lockmate/internal/config"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
)

const path = "main"

type launcher struct {
	drv   *driver
	infra *infrastructure
	app   *application
	pres  *presentation
}

func Launch() int {
	if err := config.LoadEnv(); err != nil {
		slog.Error("failed to load environment configuration", "tag", path+"/Launch", "err", err)
		return 1
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	l := &launcher{}
	if err := l.newDriver(ctx); err != nil {
		return 1
	}
	defer l.drv.cleanup()

	l.newInfrastructure()
	l.newApplication()
	l.newPresentation()

	maintenanceCtx, cancelMaintenance := context.WithCancel(ctx)
	maintenanceDone := make(chan struct{})
	go func() {
		defer close(maintenanceDone)
		ticker := time.NewTicker(config.Env.AuthFlowCleanupInterval)
		defer ticker.Stop()
		for {
			cleanupCtx, cancel := context.WithTimeout(maintenanceCtx, config.Env.PostgresConnectTimeout)
			_ = l.app.miscAuthFlow.PurgeExpired(cleanupCtx)
			cancel()
			select {
			case <-maintenanceCtx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	defer func() {
		cancelMaintenance()
		<-maintenanceDone
	}()

	l.infra.logger.Info(ctx, path+"/Launch", "HTTP service starting", domainmodelslogger.Meta{
		"port": config.Env.PortHttpServer,
	})

	if err := RunServer(ctx, l.drv.httpServer, config.Env.HttpServerShutdownTimeout); err != nil {
		l.infra.logger.Error(context.Background(), path+"/Launch", "HTTP service failed", domainmodelslogger.Meta{
			"err": err,
		})
		return 1
	}

	l.infra.logger.Info(context.Background(), path+"/Launch", "HTTP service stopped", nil)
	return 0
}

func RunServer(ctx context.Context, server *http.Server, shutdownTimeout time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	result := make(chan error, 1)
	go func() {
		result <- server.ListenAndServe()
	}()

	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err

	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			<-result
			return err
		}
		err := <-result
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
