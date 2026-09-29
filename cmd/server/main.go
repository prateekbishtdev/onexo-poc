package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prateekbishtdev/onexo-poc/internal/cache"
	"github.com/prateekbishtdev/onexo-poc/internal/config"
	"github.com/prateekbishtdev/onexo-poc/internal/database"
	"github.com/prateekbishtdev/onexo-poc/internal/handler"
	"github.com/prateekbishtdev/onexo-poc/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	rdb, err := cache.NewRedis(ctx, cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		return err
	}
	defer rdb.Close()

	health := handler.NewHealthHandler(map[string]handler.Checker{
		"postgres": handler.CheckerFunc(db.Ping),
		"redis": handler.CheckerFunc(func(ctx context.Context) error {
			return rdb.Ping(ctx).Err()
		}),
	})

	srv := server.New(":"+cfg.HTTPPort, health, logger)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", srv.Addr, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
