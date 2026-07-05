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
	"github.com/chef-stol/backend/internal/worker"
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
	pushRepo := storage.NewPushRepo(pool)

	authSvc := service.NewAuthService(authRepo, cfg.SessionTTL)
	slotsSvc := service.NewSlotsService(slotsRepo)
	bookingsSvc := service.NewBookingsService(bookingsRepo)
	pushSender := service.NewPushSender(cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDSubject)
	pushSvc := service.NewPushService(pushRepo, pushSender, logger)
	internalSvc := service.NewInternalService(bookingsRepo, pushSvc)

	router := httpapi.NewRouter(httpapi.Handlers{
		Auth:     httpapi.NewAuthHandler(authSvc),
		Slots:    httpapi.NewSlotsHandler(slotsSvc),
		Bookings: httpapi.NewBookingsHandler(bookingsSvc),
		Push:     httpapi.NewPushHandler(pushSvc),
		Internal: httpapi.NewInternalHandler(internalSvc),
	}, authSvc, cfg.InternalToolsToken, logger)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// FEAT-04 (FR-21, Should priority): 24h-ahead booking reminders. Runs until ctx is
	// cancelled (same shutdown signal as the HTTP server).
	go worker.RunReminderWorker(ctx, bookingsRepo, pushSvc, 5*time.Minute, logger)

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
