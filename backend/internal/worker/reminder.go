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
	FindDueForReminder(ctx context.Context, windowStart, windowEnd time.Time) ([]domain.Booking, error)
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
			runReminderTick(ctx, repo, push, interval, logger)
		}
	}
}

// runReminderTick computes the due window as [now+24h-interval, now+24h) — sized exactly to the
// tick interval so consecutive ticks tile the 24h-ahead timeline without a gap (a booking never
// gets skipped) or overlap (never double-sent, also guarded by reminder_sent_at).
func runReminderTick(ctx context.Context, repo BookingsRepo, push *service.PushService, interval time.Duration, logger *slog.Logger) {
	now := time.Now().UTC()
	windowStart := now.Add(24*time.Hour - interval)
	windowEnd := now.Add(24 * time.Hour)

	due, err := repo.FindDueForReminder(ctx, windowStart, windowEnd)
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
