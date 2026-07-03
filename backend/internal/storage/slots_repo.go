package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chef-stol/backend/internal/domain"
)

type SlotsRepo struct {
	pool *pgxpool.Pool
}

func NewSlotsRepo(pool *pgxpool.Pool) *SlotsRepo {
	return &SlotsRepo{pool: pool}
}

const slotSelect = `
	SELECT
		s.id, s.start_at, s.total_seats, s.free_seats, s.price, s.status, s.cancellation_reason,
		p.id, p.name, p.difficulty, p.photo_url, p.ingredients, p.allergens,
		ch.id, ch.name
	FROM slots s
	JOIN programs p ON p.id = s.program_id
	JOIN chefs ch ON ch.id = s.chef_id`

func scanSlot(row pgx.Row) (domain.Slot, error) {
	var s domain.Slot
	err := row.Scan(
		&s.ID, &s.StartAt, &s.TotalSeats, &s.FreeSeats, &s.Price, &s.Status, &s.CancellationReason,
		&s.Program.ID, &s.Program.Name, &s.Program.Difficulty, &s.Program.PhotoURL, &s.Program.Ingredients, &s.Program.Allergens,
		&s.Chef.ID, &s.Chef.Name,
	)
	return s, err
}

// List returns slots with start_at in [from, to], soonest first (FR-3). Filled and
// studio-cancelled slots are included, not filtered (FR-6a, FR-18).
func (r *SlotsRepo) List(ctx context.Context, from, to time.Time) ([]domain.Slot, error) {
	q := slotSelect + ` WHERE s.start_at >= $1 AND s.start_at <= $2 ORDER BY s.start_at ASC`
	rows, err := r.pool.Query(ctx, q, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []domain.Slot
	for rows.Next() {
		s, err := scanSlot(rows)
		if err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, rows.Err()
}

func (r *SlotsRepo) Get(ctx context.Context, id string) (domain.Slot, error) {
	q := slotSelect + ` WHERE s.id = $1`
	s, err := scanSlot(r.pool.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Slot{}, domain.ErrSlotNotFound
		}
		return domain.Slot{}, err
	}
	return s, nil
}
