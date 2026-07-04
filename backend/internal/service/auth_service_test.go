package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/chef-stol/backend/internal/domain"
)

// fakeAuthRepo is a map-backed in-memory stand-in for AuthRepo, mirroring the semantics of
// internal/storage/auth_repo.go's Postgres implementation (duplicate email -> ErrEmailTaken,
// unknown email -> ErrInvalidCredentials, unknown/expired/revoked session -> ErrUnauthorized).
type fakeAuthRepo struct {
	mu       sync.Mutex
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
	f.mu.Lock()
	defer f.mu.Unlock()
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
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byEmail[email]
	if !ok {
		return domain.Client{}, domain.ErrInvalidCredentials
	}
	return c, nil
}

func (f *fakeAuthRepo) CreateSession(ctx context.Context, clientID, tokenHash string, expiresAt time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sessions[tokenHash] = fakeSession{clientID: clientID, expiresAt: expiresAt}
	return nil
}

func (f *fakeAuthRepo) RevokeSession(ctx context.Context, tokenHash string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.sessions[tokenHash]
	if !ok {
		return nil // matches the real repo: revoking an unknown token affects 0 rows, no error
	}
	s.revoked = true
	f.sessions[tokenHash] = s
	return nil
}

func (f *fakeAuthRepo) FindClientBySessionToken(ctx context.Context, tokenHash string) (domain.Client, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
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

func TestAuthService_RegisterThenAuthenticate(t *testing.T) {
	svc := NewAuthService(newFakeAuthRepo(), time.Hour)

	client, token, err := svc.Register(context.Background(), "new@example.com", "password1")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if client.Email != "new@example.com" {
		t.Fatalf("client.Email = %q, want new@example.com", client.Email)
	}
	if token == "" {
		t.Fatal("Register returned empty token")
	}

	got, err := svc.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("Authenticate(issued token): %v", err)
	}
	if got.ID != client.ID {
		t.Fatalf("Authenticate returned client %q, want %q", got.ID, client.ID)
	}
}

func TestAuthService_RegisterDuplicateEmail(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewAuthService(repo, time.Hour)

	if _, _, err := svc.Register(context.Background(), "dup@example.com", "password1"); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	_, _, err := svc.Register(context.Background(), "dup@example.com", "password2")
	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Fatalf("second Register error = %v, want ErrEmailTaken", err)
	}
}

func TestAuthService_LoginSuccess(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewAuthService(repo, time.Hour)

	if _, _, err := svc.Register(context.Background(), "user@example.com", "correcthorse"); err != nil {
		t.Fatalf("Register: %v", err)
	}

	client, token, err := svc.Login(context.Background(), "user@example.com", "correcthorse")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if client.Email != "user@example.com" {
		t.Fatalf("client.Email = %q, want user@example.com", client.Email)
	}
	if _, err := svc.Authenticate(context.Background(), token); err != nil {
		t.Fatalf("Authenticate(login token): %v", err)
	}
}

// LOGIC-001: wrong password and nonexistent email must be indistinguishable to the caller.
func TestAuthService_LoginWrongPasswordAndUnknownEmailAreIndistinguishable(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewAuthService(repo, time.Hour)

	if _, _, err := svc.Register(context.Background(), "user@example.com", "correcthorse"); err != nil {
		t.Fatalf("Register: %v", err)
	}

	_, _, err := svc.Login(context.Background(), "user@example.com", "wrongpassword")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("Login with wrong password error = %v, want ErrInvalidCredentials", err)
	}

	_, _, err = svc.Login(context.Background(), "nosuchuser@example.com", "whatever1")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("Login with unknown email error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthService_LogoutRevokesSession(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewAuthService(repo, time.Hour)

	_, token, err := svc.Register(context.Background(), "user@example.com", "correcthorse")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	if err := svc.Logout(context.Background(), token); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	_, err = svc.Authenticate(context.Background(), token)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Authenticate after Logout error = %v, want ErrUnauthorized", err)
	}
}

func TestAuthService_AuthenticateNeverIssuedToken(t *testing.T) {
	svc := NewAuthService(newFakeAuthRepo(), time.Hour)

	_, err := svc.Authenticate(context.Background(), "never-issued-token")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Authenticate(bogus token) error = %v, want ErrUnauthorized", err)
	}
}
