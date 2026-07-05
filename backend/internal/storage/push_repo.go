package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chef-stol/backend/internal/domain"
)

type PushRepo struct {
	pool *pgxpool.Pool
}

func NewPushRepo(pool *pgxpool.Pool) *PushRepo {
	return &PushRepo{pool: pool}
}

// Upsert stores or updates a subscription by endpoint (FEAT-03). A subscription's owner can
// legitimately change if the same browser endpoint gets reused by a different logged-in client
// on a shared device — this upserts the new owner rather than rejecting the conflict, so
// registerPushSubscription never returns 409.
func (r *PushRepo) Upsert(ctx context.Context, clientID, endpoint, p256dh, auth string) (domain.PushSubscription, error) {
	const q = `
		INSERT INTO push_subscriptions (client_id, endpoint, p256dh, auth)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (endpoint) DO UPDATE SET
			client_id = EXCLUDED.client_id,
			p256dh = EXCLUDED.p256dh,
			auth = EXCLUDED.auth
		RETURNING id, client_id, endpoint, p256dh, auth, created_at`
	var s domain.PushSubscription
	err := r.pool.QueryRow(ctx, q, clientID, endpoint, p256dh, auth).Scan(
		&s.ID, &s.ClientID, &s.Endpoint, &s.P256dh, &s.Auth, &s.CreatedAt,
	)
	if err != nil {
		return domain.PushSubscription{}, err
	}
	return s, nil
}

// ListByClient returns every subscription belonging to clientID — used to fan a single push
// message out to all of a client's registered browsers/devices (FEAT-04/FEAT-05).
func (r *PushRepo) ListByClient(ctx context.Context, clientID string) ([]domain.PushSubscription, error) {
	const q = `SELECT id, client_id, endpoint, p256dh, auth, created_at FROM push_subscriptions WHERE client_id = $1`
	rows, err := r.pool.Query(ctx, q, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.PushSubscription
	for rows.Next() {
		var s domain.PushSubscription
		if err := rows.Scan(&s.ID, &s.ClientID, &s.Endpoint, &s.P256dh, &s.Auth, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// Delete removes a subscription by endpoint — called when a push send comes back 404/410 Gone,
// meaning the browser has invalidated it (FEAT-03). Not finding the row is not an error: the
// caller doesn't need to distinguish "already gone" from "just deleted".
func (r *PushRepo) Delete(ctx context.Context, endpoint string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM push_subscriptions WHERE endpoint = $1`, endpoint)
	return err
}
