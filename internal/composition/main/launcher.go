package compositionmain

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	seederdata "github.com/MateWorkspace/lockmate/database/seeder"
	"github.com/MateWorkspace/lockmate/internal/config"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	"google.golang.org/grpc"
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
	if err := seederdata.Load(); err != nil {
		slog.Error("failed to load embedded seed definitions", "tag", path+"/Launch", "err", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	l := &launcher{}
	if err := l.newDriver(ctx); err != nil {
		return 1
	}
	defer l.drv.cleanup()

	if err := l.newInfrastructure(); err != nil {
		slog.Error("failed to initialize infrastructure", "tag", path+"/Launch", "err", err)
		return 1
	}
	if err := l.infra.validator.SpaceSlug(seederdata.SeedData.Space.Slug); err != nil {
		l.infra.logger.Error(ctx, path+"/Launch", "invalid control-space slug in seed definitions", domainmodelslogger.Meta{"err": err})
		return 1
	}

	l.newApplication()
	if config.Env.SeederEnabled {
		if err := l.app.seeder.Run(ctx); err != nil {
			l.infra.logger.Error(ctx, path+"/Launch", "failed to seed control-space bootstrap data", domainmodelslogger.Meta{"err": err})
			return 1
		}
		l.infra.logger.Info(ctx, path+"/Launch", "seeding completed successfully", nil)
	}
	if err := l.newPresentation(ctx); err != nil {
		l.infra.logger.Error(ctx, path+"/Launch", "failed to start presentation", domainmodelslogger.Meta{"err": err})
		return 1
	}
	defer l.pres.stop()

	l.infra.logger.Info(ctx, path+"/Launch", "service starting", domainmodelslogger.Meta{
		"http_port": config.Env.PortHttpServer,
		"grpc_port": config.Env.PortGrpcServer,
	})

	if err := RunServers(ctx, l.drv.httpServer, l.drv.grpcServer, l.drv.grpcListener, config.Env.HttpServerShutdownTimeout, config.Env.GrpcServerShutdownTimeout); err != nil {
		l.infra.logger.Error(context.Background(), path+"/Launch", "service failed", domainmodelslogger.Meta{
			"err": err,
		})
		return 1
	}

	l.infra.logger.Info(context.Background(), path+"/Launch", "service stopped", nil)
	return 0
}

// RunServers serves HTTP and gRPC concurrently and drains both on cancellation.
func RunServers(ctx context.Context, httpServer *http.Server, grpcServer *grpc.Server, grpcListener net.Listener, httpShutdownTimeout, grpcShutdownTimeout time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	httpResult := make(chan error, 1)
	grpcResult := make(chan error, 1)
	go func() {
		httpResult <- httpServer.ListenAndServe()
	}()
	go func() {
		grpcResult <- grpcServer.Serve(grpcListener)
	}()

	select {
	case err := <-httpResult:
		stopGrpc(grpcServer, grpcShutdownTimeout)
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err

	case err := <-grpcResult:
		_ = httpServer.Close()
		<-httpResult
		if err == nil || errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}
		return err

	case <-ctx.Done():
		shutdownHTTP(httpServer, httpResult, httpShutdownTimeout)
		stopGrpc(grpcServer, grpcShutdownTimeout)
		<-grpcResult
		return nil
	}
}

func shutdownHTTP(server *http.Server, result chan error, timeout time.Duration) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
	}
	<-result
}

func stopGrpc(server *grpc.Server, timeout time.Duration) {
	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
		server.Stop()
	}
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
