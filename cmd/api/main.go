package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/raoamogh/hireos/internal/platform/config"
	"github.com/raoamogh/hireos/internal/platform/logger"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	log := logger.New("api", os.Stdout)

	cfg, err := config.Load()
	if err != nil {
		log.Error("configuration loading failed", "error", err)
		return fmt.Errorf("load configuration: %w", err)
	}

	log.Info(
		"application configuration loaded",
		"environment", cfg.App.Environment,
		"port", cfg.App.Port,
	)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	if err := serve(ctx, cfg, log); err != nil {
		log.Error("API server failed", "error", err)
		return err
	}

	return nil
}

func newHTTPServer(cfg config.Config) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.App.Port),
		Handler:           http.NewServeMux(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func serve(
	ctx context.Context,
	cfg config.Config,
	log *slog.Logger,
) error {
	server := newHTTPServer(cfg)

	listner, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", server.Addr, err)
	}

	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- server.Serve(listner)
	}()

	log.Info(
		"HTTP server started",
		"address", server.Addr,
		"environment", cfg.App.Environment,
	)

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("HTTP server stopped unexpectedly: %w", err)

	case <-ctx.Done():
		log.Info("Shutdown signal recieved.")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error(
				"graceful shutdown failed; forcing server close",
				"error", err,
			)

			closeErr := server.Close()
			serveErr := <-serverErrors

			if errors.Is(serveErr, http.ErrServerClosed) {
				serveErr = nil
			}

			return errors.Join(
				fmt.Errorf("graceful shutdown: %w", err),
				closeErr,
				serveErr,
			)
		}

		if err := <-serverErrors; err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server shutdown: %w", err)
		}

		log.Info("HTTP server stopped gracefully")

		return nil
	}
}
