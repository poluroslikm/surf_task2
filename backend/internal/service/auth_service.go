package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/chef-stol/backend/internal/domain"
)

// AuthRepo is the persistence port used by AuthService.
type AuthRepo interface {
	CreateClient(ctx context.Context, email, passwordHash string) (domain.Client, error)
	FindClientByEmail(ctx context.Context, email string) (domain.Client, error)
	CreateSession(ctx context.Context, clientID, tokenHash string, expiresAt time.Time) error
	RevokeSession(ctx context.Context, tokenHash string) error
	FindClientBySessionToken(ctx context.Context, tokenHash string) (domain.Client, error)
}

type AuthService struct {
	repo       AuthRepo
	sessionTTL time.Duration
}

func NewAuthService(repo AuthRepo, sessionTTL time.Duration) *AuthService {
	return &AuthService{repo: repo, sessionTTL: sessionTTL}
}

// Register creates a new client (FR-1) and immediately issues a session — auth/models.yaml
// AuthResponse returns a token from register, not just 201 with no credentials.
func (s *AuthService) Register(ctx context.Context, email, password string) (domain.Client, string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.Client{}, "", err
	}
	client, err := s.repo.CreateClient(ctx, email, string(hash))
	if err != nil {
		return domain.Client{}, "", err
	}
	token, err := s.issueSession(ctx, client.ID)
	if err != nil {
		return domain.Client{}, "", err
	}
	return client, token, nil
}

// Login verifies credentials (FR-2). Deliberately returns the same error for "no such
// email" and "wrong password" so the 401 response never leaks which one was wrong.
func (s *AuthService) Login(ctx context.Context, email, password string) (domain.Client, string, error) {
	client, err := s.repo.FindClientByEmail(ctx, email)
	if err != nil {
		return domain.Client{}, "", domain.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(client.PasswordHash), []byte(password)); err != nil {
		return domain.Client{}, "", domain.ErrInvalidCredentials
	}
	token, err := s.issueSession(ctx, client.ID)
	if err != nil {
		return domain.Client{}, "", err
	}
	return client, token, nil
}

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	return s.repo.RevokeSession(ctx, hashToken(rawToken))
}

// Authenticate resolves a bearer token to a client, used by the RequireAuth middleware
// (NFR-7/NFR-8).
func (s *AuthService) Authenticate(ctx context.Context, rawToken string) (domain.Client, error) {
	client, err := s.repo.FindClientBySessionToken(ctx, hashToken(rawToken))
	if err != nil {
		return domain.Client{}, domain.ErrUnauthorized
	}
	return client, nil
}

func (s *AuthService) issueSession(ctx context.Context, clientID string) (string, error) {
	raw, err := generateToken()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().Add(s.sessionTTL)
	if err := s.repo.CreateSession(ctx, clientID, hashToken(raw), expiresAt); err != nil {
		return "", err
	}
	return raw, nil
}

// generateToken produces the raw opaque bearer token. auth/models.yaml fixes
// bearerFormat: opaque-token — deliberately not a JWT. Only its SHA-256 hash is persisted.
func generateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
