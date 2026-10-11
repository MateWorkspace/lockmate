package compositionmain_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	compositionmain "github.com/MateWorkspace/lockmate/internal/composition/main"
	"github.com/MateWorkspace/lockmate/internal/config"
)

func availableAddress(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	return address
}
func waitForServer(t *testing.T, address string) {
	t.Helper()
	client := &http.Client{Timeout: 100 * time.Millisecond}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		response, err := client.Get("http://" + address + "/ready")
		if err == nil {
			response.Body.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server did not start")
}
func TestRunServerDrainsBeforeReturning(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	shuttingDown := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	mux.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		_, _ = io.WriteString(w, "completed")
	})
	server := &http.Server{Addr: availableAddress(t), Handler: mux, ReadHeaderTimeout: time.Second}
	server.RegisterOnShutdown(func() { close(shuttingDown) })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- compositionmain.RunServer(ctx, server, 3*time.Second) }()
	defer server.Close()
	waitForServer(t, server.Addr)
	responseResult := make(chan error, 1)
	go func() {
		response, err := http.Get("http://" + server.Addr + "/work")
		if err != nil {
			responseResult <- err
			return
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err == nil && string(body) != "completed" {
			err = errors.New("request did not finish")
		}
		responseResult <- err
	}()
	<-started
	cancel()
	<-shuttingDown
	select {
	case err := <-result:
		t.Fatalf("returned before drain: %v", err)
	default:
	}
	close(release)
	if err := <-responseResult; err != nil {
		t.Fatal(err)
	}
	if err := <-result; err != nil {
		t.Fatal(err)
	}
	if response, err := http.Get("http://" + server.Addr + "/ready"); err == nil {
		response.Body.Close()
		t.Fatal("listener stayed open")
	}
}
func TestRunServerShutdownTimeoutClosesConnections(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	mux.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) { close(started); <-release })
	server := &http.Server{Addr: availableAddress(t), Handler: mux, ReadHeaderTimeout: time.Second}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- compositionmain.RunServer(ctx, server, 50*time.Millisecond) }()
	defer server.Close()
	waitForServer(t, server.Addr)
	go func() {
		response, err := http.Get("http://" + server.Addr + "/work")
		if err == nil {
			response.Body.Close()
		}
	}()
	<-started
	cancel()
	err := <-result
	close(release)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected drain timeout, got %v", err)
	}
}
func TestRunServerStartupFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	server := &http.Server{Addr: listener.Addr().String()}
	if err := compositionmain.RunServer(context.Background(), server, time.Second); err == nil {
		t.Fatal("occupied listener succeeded")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := compositionmain.RunServer(ctx, server, time.Second); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled startup continued")
	}
}
func TestLaunchRejectsInvalidConfiguration(t *testing.T) {
	before := config.Env
	t.Cleanup(func() { config.Env = before })
	t.Chdir(t.TempDir())
	t.Setenv("LOCKMATE_TOKEN_ACCESS_SECRET", "")
	config.Env.TokenAccessSecret = ""
	if code := compositionmain.Launch(); code != 1 {
		t.Fatalf("invalid configuration exit code %d", code)
	}
}
