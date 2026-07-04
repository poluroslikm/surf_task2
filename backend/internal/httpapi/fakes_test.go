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
	return NewRouter(h, authSvc, logger)
}
