// Package worker holds background goroutines started from cmd/api/main.go.
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/chef-stol/backend/internal/domain"
	"github.com/chef-stol/backend/internal/service"
)

// BookingsRepo is the persistence port used by the reminder worker (FEAT-04).
type BookingsRepo interface {
	FindDueForReminder(ctx context.Context, now time.Time) ([]domain.Booking, error)
	MarkReminderSent(ctx context.Context, bookingID string) error
}

// RunReminderWorker ticks every interval and sends a 24h-ahead reminder push for each booking
// entering the window (FR-21, FEAT-04, Should priority). Blocks until ctx is cancelled — run it
// in its own goroutine from main. reminder_sent_at is set regardless of push send success: a
// dead/unreachable subscription must not cause infinite retries (cleanup of dead subscriptions
// is PushSender's job on a 404/410 response, not this worker's retry problem).
func RunReminderWorker(ctx context.Context, repo BookingsRepo, push *service.PushService, interval time.Duration, logger *slog.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runReminderTick(ctx, repo, push, logger)
		}
	}
}

// runReminderTick asks the repo for whatever is due right now (fixed 24h horizon, not a per-tick
// sliding window — see FindDueForReminder's doc comment for why: this makes each tick
// self-contained and restart-safe, with no window math to keep in sync with the tick interval).
func runReminderTick(ctx context.Context, repo BookingsRepo, push *service.PushService, logger *slog.Logger) {
	now := time.Now().UTC()

	due, err := repo.FindDueForReminder(ctx, now)
	if err != nil {
		logger.Error("reminder worker: find due", "error", err)
		return
	}

	for _, b := range due {
		payload, err := json.Marshal(map[string]string{
			"title": "Напоминание",
			"body":  fmt.Sprintf("Класс «%s» начнётся через 24 часа", b.Slot.Program.Name),
		})
		if err == nil {
			push.SendToClient(ctx, b.ClientID, payload)
		}
		if err := repo.MarkReminderSent(ctx, b.ID); err != nil {
			logger.Error("reminder worker: mark sent", "booking_id", b.ID, "error", err)
		}
	}
}
