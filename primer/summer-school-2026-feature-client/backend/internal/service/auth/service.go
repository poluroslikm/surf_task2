package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"regexp"
	"time"
)

var (
	ErrInvalidPhone    = errors.New("invalid phone")
	ErrInvalidCode     = errors.New("invalid code")
	ErrTooManyRequests = errors.New("too many requests")
	ErrInvalidSession  = errors.New("invalid session")
	phonePattern       = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	codePattern        = regexp.MustCompile(`^\d{4,6}$`)
)

const (
	loginPurpose  = "login"
	otpCodeLength = 4
)

type Client struct {
	ID        string
	Name      *string
	Phone     string
	CreatedAt time.Time
}

type RequestCodeResult struct {
	TTLSeconds         int
	ResendAfterSeconds int
	Code               string
}

type VerifyCodeResult struct {
	Token  string
	Client Client
	IsNew  bool
}

type Repository interface {
	LatestOTP(ctx context.Context, phone, purpose string) (OTP, bool, error)
	CreateOTP(ctx context.Context, phone, purpose, codeHash string, expiresAt time.Time) error
	ConsumeOTP(ctx context.Context, id string, now time.Time) error
	IncrementOTPAttempts(ctx context.Context, id string) error
	FindClientByPhone(ctx context.Context, phone string) (Client, bool, error)
	CreateClient(ctx context.Context, phone string, now time.Time) (Client, error)
	CreateSession(ctx context.Context, clientID, tokenHash string, expiresAt time.Time) error
	RevokeSession(ctx context.Context, tokenHash string, now time.Time) error
}

type OTP struct {
	ID           string
	CodeHash     string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	ConsumedAt   *time.Time
	AttemptCount int
}

type Service struct {
	repo        Repository
	logger      *slog.Logger
	now         func() time.Time
	codeTTL     time.Duration
	resendAfter time.Duration
	sessionTTL  time.Duration
	maxAttempts int
}

func NewService(repo Repository, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		repo:        repo,
		logger:      logger,
		now:         time.Now,
		codeTTL:     5 * time.Minute,
		resendAfter: time.Minute,
		sessionTTL:  24 * time.Hour,
		maxAttempts: 5,
	}
}

func (s *Service) RequestCode(ctx context.Context, phone string) (RequestCodeResult, error) {
	if !phonePattern.MatchString(phone) {
		return RequestCodeResult{}, ErrInvalidPhone
	}

	now := s.now().UTC()
	latest, ok, err := s.repo.LatestOTP(ctx, phone, loginPurpose)
	if err != nil {
		return RequestCodeResult{}, err
	}
	if ok && latest.ConsumedAt == nil && now.Sub(latest.CreatedAt) < s.resendAfter {
		return RequestCodeResult{}, ErrTooManyRequests
	}

	code, err := randomDigits(otpCodeLength)
	if err != nil {
		return RequestCodeResult{}, err
	}
	if err := s.repo.CreateOTP(ctx, phone, loginPurpose, HashOTP(phone, loginPurpose, code), now.Add(s.codeTTL)); err != nil {
		return RequestCodeResult{}, err
	}

	s.logger.Info("dev otp generated", "phone", phone, "purpose", loginPurpose, "code", code)
	return RequestCodeResult{TTLSeconds: int(s.codeTTL.Seconds()), ResendAfterSeconds: int(s.resendAfter.Seconds()), Code: code}, nil
}

func (s *Service) VerifyCode(ctx context.Context, phone, code string) (VerifyCodeResult, error) {
	if !phonePattern.MatchString(phone) || !codePattern.MatchString(code) {
		return VerifyCodeResult{}, ErrInvalidCode
	}

	now := s.now().UTC()
	otp, ok, err := s.repo.LatestOTP(ctx, phone, loginPurpose)
	if err != nil {
		return VerifyCodeResult{}, err
	}
	if !ok || otp.ConsumedAt != nil || !now.Before(otp.ExpiresAt) || otp.AttemptCount >= s.maxAttempts {
		return VerifyCodeResult{}, ErrInvalidCode
	}
	if otp.CodeHash != HashOTP(phone, loginPurpose, code) {
		_ = s.repo.IncrementOTPAttempts(ctx, otp.ID)
		return VerifyCodeResult{}, ErrInvalidCode
	}

	client, found, err := s.repo.FindClientByPhone(ctx, phone)
	if err != nil {
		return VerifyCodeResult{}, err
	}
	isNew := false
	if !found {
		client, err = s.repo.CreateClient(ctx, phone, now)
		if err != nil {
			return VerifyCodeResult{}, err
		}
		isNew = true
	}

	token, err := randomToken()
	if err != nil {
		return VerifyCodeResult{}, err
	}
	if err := s.repo.CreateSession(ctx, client.ID, HashToken(token), now.Add(s.sessionTTL)); err != nil {
		return VerifyCodeResult{}, err
	}
	if err := s.repo.ConsumeOTP(ctx, otp.ID, now); err != nil {
		return VerifyCodeResult{}, err
	}

	return VerifyCodeResult{Token: token, Client: client, IsNew: isNew}, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return ErrInvalidSession
	}
	return s.repo.RevokeSession(ctx, HashToken(token), s.now().UTC())
}

func HashOTP(phone, purpose, code string) string {
	sum := sha256.Sum256([]byte(phone + "|" + purpose + "|" + code))
	return base64.RawStdEncoding.EncodeToString(sum[:])
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawStdEncoding.EncodeToString(sum[:])
}

func randomDigits(length int) (string, error) {
	buf := make([]byte, length)
	for i := range buf {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("generate otp digit: %w", err)
		}
		buf[i] = byte('0' + n.Int64())
	}
	return string(buf), nil
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
