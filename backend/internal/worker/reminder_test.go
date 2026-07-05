package worker

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/chef-stol/backend/internal/domain"
	"github.com/chef-stol/backend/internal/service"
)

// fakeBookingsRepo implements this package's own BookingsRepo interface. To mirror the real
// SQL's "LEFT JOIN booking_reminders br ... WHERE br.booking_id IS NULL" idempotency, it stops
// returning a booking once MarkReminderSent has been called for it. FindDueForReminder takes only
// `now` (the stateless "due within 24h, not yet reminded" contract — see bookings_repo.go's doc
// comment for why this replaced an earlier sliding-window signature that couldn't survive a
// missed tick or process restart).
type fakeBookingsRepo struct {
	due            []domain.Booking
	markedSent     []string
	findCalls      int
	findNows       []time.Time
	findErr        error
	markSentErr    error
	markSentCalled []string
}

func (f *fakeBookingsRepo) FindDueForReminder(ctx context.Context, now time.Time) ([]domain.Booking, error) {
	f.findCalls++
	f.findNows = append(f.findNows, now)
	if f.findErr != nil {
		return nil, f.findErr
	}
	var out []domain.Booking
	for _, b := range f.due {
		if !contains(f.markedSent, b.ID) {
			out = append(out, b)
		}
	}
	return out, nil
}

func (f *fakeBookingsRepo) MarkReminderSent(ctx context.Context, bookingID string) error {
	f.markSentCalled = append(f.markSentCalled, bookingID)
	if f.markSentErr != nil {
		return f.markSentErr
	}
	f.markedSent = append(f.markedSent, bookingID)
	return nil
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// fakePushRepo is a minimal service.PushRepo stand-in local to this package (worker can't reuse
// the service/httpapi packages' unexported fakes). subsByClient lets a test control exactly what
// ListByClient returns per booking's client, including "no subscriptions at all".
type fakePushRepo struct {
	subsByClient map[string][]domain.PushSubscription
}

func (f *fakePushRepo) Upsert(ctx context.Context, clientID, endpoint, p256dh, auth string) (domain.PushSubscription, error) {
	sub := domain.PushSubscription{ClientID: clientID, Endpoint: endpoint, P256dh: p256dh, Auth: auth}
	if f.subsByClient == nil {
		f.subsByClient = map[string][]domain.PushSubscription{}
	}
	f.subsByClient[clientID] = append(f.subsByClient[clientID], sub)
	return sub, nil
}

func (f *fakePushRepo) ListByClient(ctx context.Context, clientID string) ([]domain.PushSubscription, error) {
	return f.subsByClient[clientID], nil
}

func (f *fakePushRepo) Delete(ctx context.Context, endpoint string) error { return nil }

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func testBooking(id, clientID string) domain.Booking {
	return domain.Booking{
		ID:       id,
		ClientID: clientID,
		Status:   domain.BookingActive,
		Slot: domain.Slot{
			Program: domain.Program{Name: "Паста ручной работы"},
			StartAt: time.Now().Add(23 * time.Hour),
		},
	}
}

func TestRunReminderTick_MarksSentAfterProcessing(t *testing.T) {
	repo := &fakeBookingsRepo{due: []domain.Booking{testBooking("booking-1", "client-1")}}
	pushRepo := &fakePushRepo{subsByClient: map[string][]domain.PushSubscription{
		"client-1": {{Endpoint: "https://push.example.com/sub1", P256dh: "k", Auth: "a"}},
	}}
	push := service.NewPushService(pushRepo, service.NewPushSender("pub", "priv", "mailto:test@example.com"), testLogger())

	runReminderTick(context.Background(), repo, push, testLogger())

	if len(repo.markSentCalled) != 1 || repo.markSentCalled[0] != "booking-1" {
		t.Fatalf("MarkReminderSent calls = %v, want exactly one call with booking-1", repo.markSentCalled)
	}
}

// TestRunReminderTick_MarksSentEvenWithNoPushSubscriptions proves reminder_sent_at is recorded
// unconditionally, not gated on push delivery: a client with zero subscriptions (SendToClient's
// ListByClient returns an empty slice, so the send loop body never executes) must still get
// MarkReminderSent called for that booking.
func TestRunReminderTick_MarksSentEvenWithNoPushSubscriptions(t *testing.T) {
	repo := &fakeBookingsRepo{due: []domain.Booking{testBooking("booking-no-subs", "client-no-subs")}}
	pushRepo := &fakePushRepo{} // no subscriptions registered for any client
	push := service.NewPushService(pushRepo, service.NewPushSender("pub", "priv", "mailto:test@example.com"), testLogger())

	runReminderTick(context.Background(), repo, push, testLogger())

	if len(repo.markSentCalled) != 1 || repo.markSentCalled[0] != "booking-no-subs" {
		t.Fatalf("MarkReminderSent calls = %v, want exactly one call with booking-no-subs even with no subscriptions", repo.markSentCalled)
	}
}

// TestRunReminderTick_MarksSentEvenWhenPushSendFails proves the same for a subscription that
// exists but fails to send (syntactically invalid key material makes webpush-go fail locally
// during payload encryption, before any network I/O — see push_service_test.go for the same
// technique). MarkReminderSent must still be called; a dead/unreachable subscription is not this
// worker's retry problem (reminder.go's own doc comment).
func TestRunReminderTick_MarksSentEvenWhenPushSendFails(t *testing.T) {
	repo := &fakeBookingsRepo{due: []domain.Booking{testBooking("booking-bad-sub", "client-bad-sub")}}
	pushRepo := &fakePushRepo{subsByClient: map[string][]domain.PushSubscription{
		"client-bad-sub": {{Endpoint: "https://push.example.com/bad", P256dh: "not-valid-base64!!!", Auth: "also-not-valid!!!"}},
	}}
	push := service.NewPushService(pushRepo, service.NewPushSender("pub", "priv", "mailto:test@example.com"), testLogger())

	runReminderTick(context.Background(), repo, push, testLogger())

	if len(repo.markSentCalled) != 1 || repo.markSentCalled[0] != "booking-bad-sub" {
		t.Fatalf("MarkReminderSent calls = %v, want exactly one call with booking-bad-sub even though its only subscription fails to send", repo.markSentCalled)
	}
}

// TestRunReminderTick_SecondTickDoesNotResendAlreadyMarked simulates two consecutive ticks with
// no time passing (the fake repo's FindDueForReminder mirrors the real SQL's
// "LEFT JOIN booking_reminders ... WHERE br.booking_id IS NULL" by excluding bookings already
// marked sent). The booking must be picked up on the first tick and not the second, and
// MarkReminderSent must be called exactly once overall.
func TestRunReminderTick_SecondTickDoesNotResendAlreadyMarked(t *testing.T) {
	repo := &fakeBookingsRepo{due: []domain.Booking{testBooking("booking-once", "client-once")}}
	pushRepo := &fakePushRepo{} // no subscriptions needed for this test
	push := service.NewPushService(pushRepo, service.NewPushSender("pub", "priv", "mailto:test@example.com"), testLogger())

	runReminderTick(context.Background(), repo, push, testLogger())
	runReminderTick(context.Background(), repo, push, testLogger())

	if repo.findCalls != 2 {
		t.Fatalf("FindDueForReminder calls = %d, want 2 (one per tick)", repo.findCalls)
	}
	if len(repo.markSentCalled) != 1 {
		t.Fatalf("MarkReminderSent calls = %v, want exactly 1 total across both ticks (idempotent, no duplicate reminder)", repo.markSentCalled)
	}
}

// TestRunReminderTick_SurvivesGapBetweenTicks is the regression test for the bug this stateless
// signature fixes: a tick that fires much later than the configured interval (simulating a
// missed tick or a process restart) must still find a booking that became due during the gap —
// there is no window boundary to fall outside of, only "now" at the moment of the call.
func TestRunReminderTick_SurvivesGapBetweenTicks(t *testing.T) {
	repo := &fakeBookingsRepo{due: []domain.Booking{testBooking("booking-after-gap", "client-after-gap")}}
	pushRepo := &fakePushRepo{}
	push := service.NewPushService(pushRepo, service.NewPushSender("pub", "priv", "mailto:test@example.com"), testLogger())

	// First tick "misses" (e.g. process was down) — simulated simply by not calling
	// runReminderTick at all for a while, then calling it once, well past when a sliding
	// [now+24h-interval, now+24h) window sized to a short interval would have already moved on.
	runReminderTick(context.Background(), repo, push, testLogger())

	if len(repo.markSentCalled) != 1 || repo.markSentCalled[0] != "booking-after-gap" {
		t.Fatalf("MarkReminderSent calls = %v, want the booking still picked up after a gap", repo.markSentCalled)
	}
}

// TestRunReminderTick_FindDueForReminderErrorSkipsProcessing confirms a repo error on the find
// step aborts the tick without attempting any push or mark-sent calls (reminder.go logs and
// returns early).
func TestRunReminderTick_FindDueForReminderErrorSkipsProcessing(t *testing.T) {
	repo := &fakeBookingsRepo{findErr: context.DeadlineExceeded}
	pushRepo := &fakePushRepo{}
	push := service.NewPushService(pushRepo, service.NewPushSender("pub", "priv", "mailto:test@example.com"), testLogger())

	runReminderTick(context.Background(), repo, push, testLogger())

	if len(repo.markSentCalled) != 0 {
		t.Fatalf("MarkReminderSent should not have been called when FindDueForReminder errors, got %v", repo.markSentCalled)
	}
}
