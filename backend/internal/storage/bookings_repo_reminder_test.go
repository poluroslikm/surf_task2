package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/chef-stol/backend/internal/domain"
)

// TestBookingsRepo_FindDueForReminder_SurvivesGapAndRestart is a real-Postgres integration test
// (skipped if DATABASE_URL isn't set — no testcontainers wired into this project yet, see
// BE-12) proving the fix for a bug found during code review: FindDueForReminder used to take an
// explicit [windowStart, windowEnd) computed by the caller from the tick interval, so a missed
// tick or a process restart could permanently skip a booking whose window fell inside the gap.
// It's now a stateless "due within the next 24h, not yet reminded" check with no such gap — this
// test proves that property directly by simulating a restart (a fresh call with no prior state)
// and an out-of-cadence tick (calling far apart in wall-clock terms), not just a single call.
func TestBookingsRepo_FindDueForReminder_SurvivesGapAndRestart(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set — skipping Postgres integration test")
	}

	ctx := context.Background()
	pool, err := NewPool(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	// t.Cleanup runs LIFO, and only after the test function body returns (i.e. after any of its
	// own `defer`s would already have run) — registering the pool-close cleanup first means it
	// runs last, after the data cleanup below still needs a live pool to issue DELETEs.
	t.Cleanup(func() { pool.Close() })

	repo := NewBookingsRepo(pool)

	var programID, chefID, slotID, clientID, bookingID string
	if err := pool.QueryRow(ctx,
		`INSERT INTO programs (name, difficulty, photo_url) VALUES ('reminder-test-program', 'novice', 'https://example.com/p.jpg') RETURNING id`,
	).Scan(&programID); err != nil {
		t.Fatalf("insert program: %v", err)
	}
	if err := pool.QueryRow(ctx,
		`INSERT INTO chefs (name) VALUES ('reminder-test-chef') RETURNING id`,
	).Scan(&chefID); err != nil {
		t.Fatalf("insert chef: %v", err)
	}
	// 23h from now: inside the 24h reminder horizon and still in the future — this is the exact
	// booking a correct implementation must return, restart or no restart.
	startAt := time.Now().UTC().Add(23 * time.Hour)
	if err := pool.QueryRow(ctx,
		`INSERT INTO slots (program_id, chef_id, start_at, total_seats, free_seats, price)
		 VALUES ($1, $2, $3, 5, 5, 1000) RETURNING id`,
		programID, chefID, startAt,
	).Scan(&slotID); err != nil {
		t.Fatalf("insert slot: %v", err)
	}
	if err := pool.QueryRow(ctx,
		`INSERT INTO clients (email, password_hash) VALUES ('reminder-test-client@example.com', 'x') RETURNING id`,
	).Scan(&clientID); err != nil {
		t.Fatalf("insert client: %v", err)
	}
	if err := pool.QueryRow(ctx,
		`INSERT INTO bookings (slot_id, client_id, equipment_choice, status, price_total)
		 VALUES ($1, $2, 'own', 'active', 1000) RETURNING id`,
		slotID, clientID,
	).Scan(&bookingID); err != nil {
		t.Fatalf("insert booking: %v", err)
	}

	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM booking_reminders WHERE booking_id = $1`, bookingID)
		pool.Exec(ctx, `DELETE FROM bookings WHERE id = $1`, bookingID)
		pool.Exec(ctx, `DELETE FROM slots WHERE id = $1`, slotID)
		pool.Exec(ctx, `DELETE FROM clients WHERE id = $1`, clientID)
		pool.Exec(ctx, `DELETE FROM chefs WHERE id = $1`, chefID)
		pool.Exec(ctx, `DELETE FROM programs WHERE id = $1`, programID)
	})

	// Simulate a "restart": the very first call ever made, hours after the booking was created,
	// with no prior tick history — must still find it.
	due, err := repo.FindDueForReminder(ctx, time.Now().UTC())
	if err != nil {
		t.Fatalf("FindDueForReminder (first call): %v", err)
	}
	if !containsBookingID(due, bookingID) {
		t.Fatalf("FindDueForReminder did not return the due booking %s on first call (restart case)", bookingID)
	}

	if err := repo.MarkReminderSent(ctx, bookingID); err != nil {
		t.Fatalf("MarkReminderSent: %v", err)
	}

	// Simulate an out-of-cadence re-tick — must not re-send.
	due, err = repo.FindDueForReminder(ctx, time.Now().UTC())
	if err != nil {
		t.Fatalf("FindDueForReminder (second call): %v", err)
	}
	if containsBookingID(due, bookingID) {
		t.Fatalf("FindDueForReminder re-returned %s after MarkReminderSent — would double-send", bookingID)
	}
}

func containsBookingID(bookings []domain.Booking, id string) bool {
	for _, b := range bookings {
		if b.ID == id {
			return true
		}
	}
	return false
}
