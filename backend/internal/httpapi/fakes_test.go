package httpapi

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/chef-stol/backend/internal/domain"
	"github.com/chef-stol/backend/internal/service"
)

// fakeAuthRepo/fakeSlotsRepo are map-backed in-memory stand-ins for service.AuthRepo/SlotsRepo,
// used to drive the real *service.AuthService/*service.SlotsService under the real HTTP handlers
// without any external dependency. Mirrors the semantics of internal/storage's Postgres
// implementation (see auth_repo.go/slots_repo.go for the error conventions being matched).
type fakeAuthRepo struct {
	byEmail  map[string]domain.Client
	byID     map[string]domain.Client
	sessions map[string]fakeSession
	nextID   int
}

type fakeSession struct {
	clientID  string
	expiresAt time.Time
	revoked   bool
}

func newFakeAuthRepo() *fakeAuthRepo {
	return &fakeAuthRepo{
		byEmail:  make(map[string]domain.Client),
		byID:     make(map[string]domain.Client),
		sessions: make(map[string]fakeSession),
	}
}

func (f *fakeAuthRepo) CreateClient(ctx context.Context, email, passwordHash string) (domain.Client, error) {
	if _, ok := f.byEmail[email]; ok {
		return domain.Client{}, domain.ErrEmailTaken
	}
	f.nextID++
	c := domain.Client{
		ID:           fmt.Sprintf("client-%d", f.nextID),
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
	f.byEmail[email] = c
	f.byID[c.ID] = c
	return c, nil
}

func (f *fakeAuthRepo) FindClientByEmail(ctx context.Context, email string) (domain.Client, error) {
	c, ok := f.byEmail[email]
	if !ok {
		return domain.Client{}, domain.ErrInvalidCredentials
	}
	return c, nil
}

func (f *fakeAuthRepo) CreateSession(ctx context.Context, clientID, tokenHash string, expiresAt time.Time) error {
	f.sessions[tokenHash] = fakeSession{clientID: clientID, expiresAt: expiresAt}
	return nil
}

func (f *fakeAuthRepo) RevokeSession(ctx context.Context, tokenHash string) error {
	s, ok := f.sessions[tokenHash]
	if !ok {
		return nil
	}
	s.revoked = true
	f.sessions[tokenHash] = s
	return nil
}

func (f *fakeAuthRepo) FindClientBySessionToken(ctx context.Context, tokenHash string) (domain.Client, error) {
	s, ok := f.sessions[tokenHash]
	if !ok || s.revoked || time.Now().After(s.expiresAt) {
		return domain.Client{}, domain.ErrUnauthorized
	}
	c, ok := f.byID[s.clientID]
	if !ok {
		return domain.Client{}, domain.ErrUnauthorized
	}
	return c, nil
}

type fakeSlotsRepo struct {
	slots []domain.Slot
}

func (f *fakeSlotsRepo) List(ctx context.Context, from, to time.Time) ([]domain.Slot, error) {
	return f.slots, nil
}

func (f *fakeSlotsRepo) Get(ctx context.Context, id string) (domain.Slot, error) {
	for _, s := range f.slots {
		if s.ID == id {
			return s, nil
		}
	}
	return domain.Slot{}, domain.ErrSlotNotFound
}

// newTestRouter wires the real handlers/services on top of fake repos — no external dependency,
// but exercises the full request -> response path (JSON decoding, routing, status mapping).
func newTestRouter(authRepo *fakeAuthRepo, slotsRepo *fakeSlotsRepo) http.Handler {
	authSvc := service.NewAuthService(authRepo, time.Hour)
	slotsSvc := service.NewSlotsService(slotsRepo)
	h := Handlers{Auth: NewAuthHandler(authSvc), Slots: NewSlotsHandler(slotsSvc)}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewRouter(h, authSvc, "test-internal-token", logger)
}

// fakeBookingsRepo is a map-backed stand-in for service.BookingsRepo. Get/List/Cancel/
// SubmitRating enforce per-client ownership by checking the stored booking's ClientID, mirroring
// internal/storage/bookings_repo.go's row-level ownership checks (WHERE b.client_id = $1 /
// bookingClientID != clientID -> ErrForbidden). Create/Cancel/SubmitRating additionally return
// whatever canned error a test configures, standing in for the SQL-layer business rules (seat
// CAS, 24h cancel boundary, rating eligibility) that this fake does not reimplement.
type fakeBookingsRepo struct {
	bookings map[string]domain.Booking

	createResult domain.Booking
	createErr    error

	cancelResult domain.Booking
	cancelErr    error

	ratingResult domain.Booking
	ratingErr    error
}

func newFakeBookingsRepo() *fakeBookingsRepo {
	return &fakeBookingsRepo{bookings: make(map[string]domain.Booking)}
}

func (f *fakeBookingsRepo) Create(ctx context.Context, clientID, slotID string, equipment domain.EquipmentChoice) (domain.Booking, error) {
	if f.createErr != nil {
		return domain.Booking{}, f.createErr
	}
	return f.createResult, nil
}

func (f *fakeBookingsRepo) List(ctx context.Context, clientID string) ([]domain.Booking, error) {
	var out []domain.Booking
	for _, b := range f.bookings {
		if b.ClientID == clientID {
			out = append(out, b)
		}
	}
	return out, nil
}

func (f *fakeBookingsRepo) Get(ctx context.Context, id string) (domain.Booking, error) {
	b, ok := f.bookings[id]
	if !ok {
		return domain.Booking{}, domain.ErrBookingNotFound
	}
	return b, nil
}

func (f *fakeBookingsRepo) Cancel(ctx context.Context, bookingID, clientID string) (domain.Booking, error) {
	b, ok := f.bookings[bookingID]
	if !ok {
		return domain.Booking{}, domain.ErrBookingNotFound
	}
	if b.ClientID != clientID {
		return domain.Booking{}, domain.ErrForbidden
	}
	if f.cancelErr != nil {
		return domain.Booking{}, f.cancelErr
	}
	return f.cancelResult, nil
}

func (f *fakeBookingsRepo) SubmitRating(ctx context.Context, bookingID, clientID string, stars int, comment *string) (domain.Booking, error) {
	b, ok := f.bookings[bookingID]
	if !ok {
		return domain.Booking{}, domain.ErrBookingNotFound
	}
	if b.ClientID != clientID {
		return domain.Booking{}, domain.ErrForbidden
	}
	if f.ratingErr != nil {
		return domain.Booking{}, f.ratingErr
	}
	// Mirrors bookings_repo.go's SubmitRating, which validates the 1-5 stars range only in the
	// SQL layer (see report finding: nothing in the handler/DTO validates this beforehand).
	if stars < 1 || stars > 5 {
		return domain.Booking{}, domain.ErrInvalidRating
	}
	return f.ratingResult, nil
}

// newTestRouterWithBookings extends newTestRouter with a wired BookingsHandler. Added as a
// separate function (rather than changing newTestRouter's signature) so the existing auth/slots
// handler tests calling newTestRouter are untouched.
func newTestRouterWithBookings(authRepo *fakeAuthRepo, slotsRepo *fakeSlotsRepo, bookingsRepo *fakeBookingsRepo) http.Handler {
	authSvc := service.NewAuthService(authRepo, time.Hour)
	slotsSvc := service.NewSlotsService(slotsRepo)
	bookingsSvc := service.NewBookingsService(bookingsRepo)
	h := Handlers{Auth: NewAuthHandler(authSvc), Slots: NewSlotsHandler(slotsSvc), Bookings: NewBookingsHandler(bookingsSvc)}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewRouter(h, authSvc, "test-internal-token", logger)
}
