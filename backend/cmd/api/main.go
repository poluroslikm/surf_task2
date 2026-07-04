package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chef-stol/backend/internal/config"
	"github.com/chef-stol/backend/internal/httpapi"
	"github.com/chef-stol/backend/internal/service"
	"github.com/chef-stol/backend/internal/storage"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := storage.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	authRepo := storage.NewAuthRepo(pool)
	slotsRepo := storage.NewSlotsRepo(pool)
	bookingsRepo := storage.NewBookingsRepo(pool)

	authSvc := service.NewAuthService(authRepo, cfg.SessionTTL)
	slotsSvc := service.NewSlotsService(slotsRepo)
	bookingsSvc := service.NewBookingsService(bookingsRepo)

	router := httpapi.NewRouter(httpapi.Handlers{
		Auth:     httpapi.NewAuthHandler(authSvc),
		Slots:    httpapi.NewSlotsHandler(slotsSvc),
		Bookings: httpapi.NewBookingsHandler(bookingsSvc),
	}, authSvc, logger)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("http server starting", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
