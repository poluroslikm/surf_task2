package service

import (
	"context"
	"errors"
	"testing"

	"github.com/chef-stol/backend/internal/domain"
)

// fakeBookingsRepo is a map/canned-response stand-in for BookingsRepo. Get looks a booking up
// by ID only (mirroring internal/storage/bookings_repo.go's Get, which takes no clientID —
// ownership enforcement is BookingsService.Get's job, and is what's under test here).
// Create/List/Cancel/SubmitRating record the arguments they were called with and return
// whatever canned result/error a test configures, since those methods are thin pass-throughs
// to the repo and the repo's real business rules (seat CAS, 24h cancel boundary, rating
// eligibility) live in SQL, not in a fake worth re-implementing.
type fakeBookingsRepo struct {
	bookings map[string]domain.Booking

	createCalledClientID string
	createCalledSlotID   string
	createCalledEquip    domain.EquipmentChoice
	createResult         domain.Booking
	createErr            error

	listCalledClientID string
	listResult         []domain.Booking
	listErr            error

	cancelCalledBookingID string
	cancelCalledClientID  string
	cancelResult          domain.Booking
	cancelErr             error

	ratingCalledBookingID string
	ratingCalledClientID  string
	ratingCalledStars     int
	ratingCalledComment   *string
	ratingResult          domain.Booking
	ratingErr             error
}

func newFakeBookingsRepo() *fakeBookingsRepo {
	return &fakeBookingsRepo{bookings: make(map[string]domain.Booking)}
}

func (f *fakeBookingsRepo) Create(ctx context.Context, clientID, slotID string, equipment domain.EquipmentChoice) (domain.Booking, error) {
	f.createCalledClientID = clientID
	f.createCalledSlotID = slotID
	f.createCalledEquip = equipment
	return f.createResult, f.createErr
}

func (f *fakeBookingsRepo) List(ctx context.Context, clientID string) ([]domain.Booking, error) {
	f.listCalledClientID = clientID
	return f.listResult, f.listErr
}

func (f *fakeBookingsRepo) Get(ctx context.Context, id string) (domain.Booking, error) {
	b, ok := f.bookings[id]
	if !ok {
		return domain.Booking{}, domain.ErrBookingNotFound
	}
	return b, nil
}

func (f *fakeBookingsRepo) Cancel(ctx context.Context, bookingID, clientID string) (domain.Booking, error) {
	f.cancelCalledBookingID = bookingID
	f.cancelCalledClientID = clientID
	return f.cancelResult, f.cancelErr
}

func (f *fakeBookingsRepo) SubmitRating(ctx context.Context, bookingID, clientID string, stars int, comment *string) (domain.Booking, error) {
	f.ratingCalledBookingID = bookingID
	f.ratingCalledClientID = clientID
	f.ratingCalledStars = stars
	f.ratingCalledComment = comment
	return f.ratingResult, f.ratingErr
}

// TestBookingsService_GetOwnBooking: the requesting client is the booking's owner, so the
// booking comes back unchanged.
func TestBookingsService_GetOwnBooking(t *testing.T) {
	want := domain.Booking{ID: "booking-1", ClientID: "client-1", Status: domain.BookingActive}
	repo := newFakeBookingsRepo()
	repo.bookings["booking-1"] = want
	svc := NewBookingsService(repo)

	got, err := svc.Get(context.Background(), "client-1", "booking-1")
	if err != nil {
		t.Fatalf("Get(own): %v", err)
	}
	if got.ID != want.ID || got.ClientID != want.ClientID || got.Status != want.Status {
		t.Fatalf("Get(own) = %+v, want %+v", got, want)
	}
}

// TestBookingsService_GetForbidsOtherClientsBooking: NFR-8 — getBooking returns 403 (via
// domain.ErrForbidden), not the raw booking, when the booking belongs to a different client.
// This is the one real piece of service-layer business logic in this file.
func TestBookingsService_GetForbidsOtherClientsBooking(t *testing.T) {
	repo := newFakeBookingsRepo()
	repo.bookings["booking-1"] = domain.Booking{ID: "booking-1", ClientID: "client-owner"}
	svc := NewBookingsService(repo)

	got, err := svc.Get(context.Background(), "client-other", "booking-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Get(other client) error = %v, want ErrForbidden", err)
	}
	if got.ID != "" {
		t.Fatalf("Get(other client) returned non-zero booking %+v, want zero value alongside the error", got)
	}
}

// TestBookingsService_GetPropagatesNotFound: an unknown booking ID's error passes straight
// through, unchanged, from the repo.
func TestBookingsService_GetPropagatesNotFound(t *testing.T) {
	repo := newFakeBookingsRepo()
	svc := NewBookingsService(repo)

	_, err := svc.Get(context.Background(), "client-1", "missing")
	if !errors.Is(err, domain.ErrBookingNotFound) {
		t.Fatalf("Get(missing) error = %v, want ErrBookingNotFound", err)
	}
}

// TestBookingsService_CreatePassesThroughArgsAndResult confirms Create is a pure pass-through:
// the exact clientID/slotID/equipment reach the repo, and the repo's result comes back
// unchanged.
func TestBookingsService_CreatePassesThroughArgsAndResult(t *testing.T) {
	repo := newFakeBookingsRepo()
	repo.createResult = domain.Booking{ID: "booking-new", ClientID: "client-1", SlotID: "slot-1"}
	svc := NewBookingsService(repo)

	got, err := svc.Create(context.Background(), "client-1", "slot-1", domain.EquipmentRental)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.ID != "booking-new" {
		t.Fatalf("Create result = %+v, want ID booking-new", got)
	}
	if repo.createCalledClientID != "client-1" || repo.createCalledSlotID != "slot-1" || repo.createCalledEquip != domain.EquipmentRental {
		t.Fatalf("repo.Create called with (%q, %q, %q), want (client-1, slot-1, rental)",
			repo.createCalledClientID, repo.createCalledSlotID, repo.createCalledEquip)
	}
}

// TestBookingsService_CreatePassesThroughRepoError confirms a repo error (here, the dynamic
// *domain.ErrSlotFull) is forwarded unchanged, not swallowed or wrapped.
func TestBookingsService_CreatePassesThroughRepoError(t *testing.T) {
	repo := newFakeBookingsRepo()
	repo.createErr = &domain.ErrSlotFull{AvailableSeats: 0}
	svc := NewBookingsService(repo)

	_, err := svc.Create(context.Background(), "client-1", "slot-1", domain.EquipmentOwn)
	var slotFull *domain.ErrSlotFull
	if !errors.As(err, &slotFull) {
		t.Fatalf("Create error = %v, want *domain.ErrSlotFull", err)
	}
	if slotFull.AvailableSeats != 0 {
		t.Fatalf("ErrSlotFull.AvailableSeats = %d, want 0", slotFull.AvailableSeats)
	}
}

// TestBookingsService_ListPassesThroughClientIDAndResult confirms List forwards the caller's
// clientID to the repo and returns exactly what the repo returned.
func TestBookingsService_ListPassesThroughClientIDAndResult(t *testing.T) {
	repo := newFakeBookingsRepo()
	repo.listResult = []domain.Booking{{ID: "b1"}, {ID: "b2"}}
	svc := NewBookingsService(repo)

	got, err := svc.List(context.Background(), "client-1")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("List returned %d bookings, want 2", len(got))
	}
	if repo.listCalledClientID != "client-1" {
		t.Fatalf("repo.List called with clientID=%q, want client-1", repo.listCalledClientID)
	}
}

// TestBookingsService_CancelPassesThroughArgsAndResult confirms Cancel forwards
// bookingID/clientID and returns the repo's resulting booking (e.g. status=late_cancel)
// unchanged.
func TestBookingsService_CancelPassesThroughArgsAndResult(t *testing.T) {
	repo := newFakeBookingsRepo()
	repo.cancelResult = domain.Booking{ID: "booking-1", Status: domain.BookingLateCancel}
	svc := NewBookingsService(repo)

	got, err := svc.Cancel(context.Background(), "client-1", "booking-1")
	if err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if got.Status != domain.BookingLateCancel {
		t.Fatalf("Cancel result status = %q, want late_cancel", got.Status)
	}
	if repo.cancelCalledBookingID != "booking-1" || repo.cancelCalledClientID != "client-1" {
		t.Fatalf("repo.Cancel called with (%q, %q), want (booking-1, client-1)",
			repo.cancelCalledBookingID, repo.cancelCalledClientID)
	}
}

// TestBookingsService_CancelPassesThroughRepoError confirms a repo error (e.g. the slot has
// already started) is forwarded unchanged.
func TestBookingsService_CancelPassesThroughRepoError(t *testing.T) {
	repo := newFakeBookingsRepo()
	repo.cancelErr = domain.ErrSlotStarted
	svc := NewBookingsService(repo)

	_, err := svc.Cancel(context.Background(), "client-1", "booking-1")
	if !errors.Is(err, domain.ErrSlotStarted) {
		t.Fatalf("Cancel error = %v, want ErrSlotStarted", err)
	}
}

// TestBookingsService_SubmitRatingPassesThroughArgsAndResult confirms SubmitRating forwards
// every argument (bookingID, clientID, stars, comment) and returns the repo's result unchanged.
func TestBookingsService_SubmitRatingPassesThroughArgsAndResult(t *testing.T) {
	repo := newFakeBookingsRepo()
	comment := "Было очень вкусно!"
	repo.ratingResult = domain.Booking{ID: "booking-1", Rating: &domain.Rating{Stars: 5, Comment: &comment}}
	svc := NewBookingsService(repo)

	got, err := svc.SubmitRating(context.Background(), "client-1", "booking-1", 5, &comment)
	if err != nil {
		t.Fatalf("SubmitRating: %v", err)
	}
	if got.Rating == nil || got.Rating.Stars != 5 {
		t.Fatalf("SubmitRating result = %+v, want rating.stars=5", got)
	}
	if repo.ratingCalledBookingID != "booking-1" || repo.ratingCalledClientID != "client-1" ||
		repo.ratingCalledStars != 5 || repo.ratingCalledComment != &comment {
		t.Fatalf("repo.SubmitRating called with (%q,%q,%d,%p), want (booking-1,client-1,5,%p)",
			repo.ratingCalledBookingID, repo.ratingCalledClientID, repo.ratingCalledStars, repo.ratingCalledComment, &comment)
	}
}

// TestBookingsService_SubmitRatingPassesThroughRepoError confirms a repo error (e.g. already
// rated) is forwarded unchanged.
func TestBookingsService_SubmitRatingPassesThroughRepoError(t *testing.T) {
	repo := newFakeBookingsRepo()
	repo.ratingErr = domain.ErrAlreadyRated
	svc := NewBookingsService(repo)

	_, err := svc.SubmitRating(context.Background(), "client-1", "booking-1", 5, nil)
	if !errors.Is(err, domain.ErrAlreadyRated) {
		t.Fatalf("SubmitRating error = %v, want ErrAlreadyRated", err)
	}
}
