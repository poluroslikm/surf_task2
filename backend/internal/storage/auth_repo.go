package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chef-stol/backend/internal/domain"
)

type AuthRepo struct {
	pool *pgxpool.Pool
}

func NewAuthRepo(pool *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{pool: pool}
}

func (r *AuthRepo) CreateClient(ctx context.Context, email, passwordHash string) (domain.Client, error) {
	const q = `
		INSERT INTO clients (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, created_at`
	var c domain.Client
	err := r.pool.QueryRow(ctx, q, email, passwordHash).Scan(&c.ID, &c.Email, &c.PasswordHash, &c.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Client{}, domain.ErrEmailTaken
		}
		return domain.Client{}, err
	}
	return c, nil
}

func (r *AuthRepo) FindClientByEmail(ctx context.Context, email string) (domain.Client, error) {
	const q = `
		SELECT id, email, password_hash, created_at
		FROM clients
		WHERE email = $1`
	var c domain.Client
	err := r.pool.QueryRow(ctx, q, email).Scan(&c.ID, &c.Email, &c.PasswordHash, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Client{}, domain.ErrInvalidCredentials
		}
		return domain.Client{}, err
	}
	return c, nil
}

func (r *AuthRepo) CreateSession(ctx context.Context, clientID, tokenHash string, expiresAt time.Time) error {
	const q = `
		INSERT INTO auth_sessions (client_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`
	_, err := r.pool.Exec(ctx, q, clientID, tokenHash, expiresAt)
	return err
}

func (r *AuthRepo) RevokeSession(ctx context.Context, tokenHash string) error {
	const q = `
		UPDATE auth_sessions
		SET revoked_at = now()
		WHERE token_hash = $1 AND revoked_at IS NULL`
	_, err := r.pool.Exec(ctx, q, tokenHash)
	return err
}

func (r *AuthRepo) FindClientBySessionToken(ctx context.Context, tokenHash string) (domain.Client, error) {
	const q = `
		SELECT c.id, c.email, c.password_hash, c.created_at
		FROM auth_sessions s
		JOIN clients c ON c.id = s.client_id
		WHERE s.token_hash = $1
		  AND s.revoked_at IS NULL
		  AND s.expires_at > now()`
	var c domain.Client
	err := r.pool.QueryRow(ctx, q, tokenHash).Scan(&c.ID, &c.Email, &c.PasswordHash, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Client{}, domain.ErrUnauthorized
		}
		return domain.Client{}, err
	}
	return c, nil
}
