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

type BookingsRepo struct {
	pool *pgxpool.Pool
}

func NewBookingsRepo(pool *pgxpool.Pool) *BookingsRepo {
	return &BookingsRepo{pool: pool}
}

const bookingSelect = `
	SELECT
		b.id, b.slot_id, b.client_id, b.equipment_choice, b.status, b.price_total,
		b.rating_stars, b.rating_comment, b.rating_created_at, b.created_at, b.cancelled_at,
		s.id, s.start_at, s.total_seats, s.free_seats, s.price, s.status, s.cancellation_reason,
		p.id, p.name, p.difficulty, p.photo_url, p.ingredients, p.allergens,
		ch.id, ch.name
	FROM bookings b
	JOIN slots s ON s.id = b.slot_id
	JOIN programs p ON p.id = s.program_id
	JOIN chefs ch ON ch.id = s.chef_id`

// scanBooking populates FreeCancellationUntil from the joined slot's start_at (slot.start_at
// − 24h) — this column is never stored (data-model.md → Booking).
func scanBooking(row pgx.Row) (domain.Booking, error) {
	var b domain.Booking
	var ratingStars *int
	var ratingComment *string
	var ratingCreatedAt *time.Time
	err := row.Scan(
		&b.ID, &b.SlotID, &b.ClientID, &b.EquipmentChoice, &b.Status, &b.PriceTotal,
		&ratingStars, &ratingComment, &ratingCreatedAt, &b.CreatedAt, &b.CancelledAt,
		&b.Slot.ID, &b.Slot.StartAt, &b.Slot.TotalSeats, &b.Slot.FreeSeats, &b.Slot.Price, &b.Slot.Status, &b.Slot.CancellationReason,
		&b.Slot.Program.ID, &b.Slot.Program.Name, &b.Slot.Program.Difficulty, &b.Slot.Program.PhotoURL, &b.Slot.Program.Ingredients, &b.Slot.Program.Allergens,
		&b.Slot.Chef.ID, &b.Slot.Chef.Name,
	)
	if err != nil {
		return domain.Booking{}, err
	}
	b.FreeCancellationUntil = b.Slot.StartAt.Add(-24 * time.Hour)
	if ratingStars != nil {
		b.Rating = &domain.Rating{Stars: *ratingStars, Comment: ratingComment, CreatedAt: *ratingCreatedAt}
	}
	return b, nil
}

func (r *BookingsRepo) Get(ctx context.Context, id string) (domain.Booking, error) {
	q := bookingSelect + ` WHERE b.id = $1`
	b, err := scanBooking(r.pool.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Booking{}, domain.ErrBookingNotFound
		}
		return domain.Booking{}, err
	}
	return b, nil
}

// List returns every booking belonging to clientID, most-recent-slot-first (listBookings has
// no pagination and no status filter — api/bookings/api.yaml confirms neither exists for this
// endpoint, so all statuses including cancelled ones come back).
func (r *BookingsRepo) List(ctx context.Context, clientID string) ([]domain.Booking, error) {
	q := bookingSelect + ` WHERE b.client_id = $1 ORDER BY s.start_at DESC`
	rows, err := r.pool.Query(ctx, q, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Booking
	for rows.Next() {
		b, err := scanBooking(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// Create books a seat on slotID for clientID atomically (NFR-6, "0 double bookings"): the seat
// CAS-decrement and the booking insert happen in one transaction, so any failure after the
// decrement rolls the seat back. The UPDATE...RETURNING is the single source of truth for
// "was there a free seat" — there is no separate SELECT-then-check step that could race.
func (r *BookingsRepo) Create(ctx context.Context, clientID, slotID string, equipment domain.EquipmentChoice) (domain.Booking, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Booking{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op once Commit has succeeded

	const decrementQ = `
		UPDATE slots SET free_seats = free_seats - 1
		WHERE id = $1 AND status = 'scheduled' AND free_seats > 0
		RETURNING price`
	var price int
	err = tx.QueryRow(ctx, decrementQ, slotID).Scan(&price)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return domain.Booking{}, err
		}
		// 0 rows: the CAS condition failed. Re-check the slot's actual state to pick the
		// right error — unknown slot_id, studio-cancelled slot, or genuinely full.
		var status string
		var freeSeats int
		checkErr := tx.QueryRow(ctx, `SELECT status, free_seats FROM slots WHERE id = $1`, slotID).Scan(&status, &freeSeats)
		if checkErr != nil {
			if errors.Is(checkErr, pgx.ErrNoRows) {
				return domain.Booking{}, domain.ErrInvalidSlotID
			}
			return domain.Booking{}, checkErr
		}
		if status == string(domain.SlotCancelled) {
			return domain.Booking{}, domain.ErrSlotCancelled
		}
		return domain.Booking{}, &domain.ErrSlotFull{AvailableSeats: freeSeats}
	}

	const insertQ = `
		INSERT INTO bookings (slot_id, client_id, equipment_choice, price_total)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	var bookingID string
	err = tx.QueryRow(ctx, insertQ, slotID, clientID, string(equipment), price).Scan(&bookingID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "bookings_active_client_slot_uidx" {
			return domain.Booking{}, domain.ErrDuplicateBooking
		}
		return domain.Booking{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Booking{}, err
	}
	return r.Get(ctx, bookingID)
}

// Cancel applies the 24h rule (FR-15/FR-16). Ownership check and status transition happen
// under a row lock on the booking (FOR UPDATE) in one transaction, so two concurrent cancel
// attempts on the same booking can't both return the seat or disagree on the resulting status.
func (r *BookingsRepo) Cancel(ctx context.Context, bookingID, clientID string) (domain.Booking, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Booking{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var status, bookingClientID, slotID string
	var startAt time.Time
	const lockQ = `
		SELECT b.status, b.client_id, b.slot_id, s.start_at
		FROM bookings b
		JOIN slots s ON s.id = b.slot_id
		WHERE b.id = $1
		FOR UPDATE OF b`
	if err := tx.QueryRow(ctx, lockQ, bookingID).Scan(&status, &bookingClientID, &slotID, &startAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Booking{}, domain.ErrBookingNotFound
		}
		return domain.Booking{}, err
	}
	if bookingClientID != clientID {
		return domain.Booking{}, domain.ErrForbidden
	}
	if status != string(domain.BookingActive) {
		return domain.Booking{}, domain.ErrAlreadyCancelled
	}

	now := time.Now().UTC()
	if !now.Before(startAt) {
		return domain.Booking{}, domain.ErrSlotStarted
	}

	// Boundary inclusive: exactly 24h before start still counts as an early (free) cancel.
	early := !now.After(startAt.Add(-24 * time.Hour))
	newStatus := domain.BookingLateCancel
	if early {
		newStatus = domain.BookingCancelled
	}

	if _, err := tx.Exec(ctx, `UPDATE bookings SET status = $1, cancelled_at = now() WHERE id = $2`, string(newStatus), bookingID); err != nil {
		return domain.Booking{}, err
	}
	if early {
		// Late cancellations never return the seat (FR-16) — no penalty, it's just not
		// released back to the pool.
		if _, err := tx.Exec(ctx, `UPDATE slots SET free_seats = free_seats + 1 WHERE id = $1`, slotID); err != nil {
			return domain.Booking{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Booking{}, err
	}
	return r.Get(ctx, bookingID)
}

// SubmitRating validates and stores the chef rating (FR-19/FR-20). Checks run in the order the
// task specifies: not found, wrong owner, already rated, not ratable, then stars range — a
// booking that fails an earlier check reports that error even if stars is also out of range.
func (r *BookingsRepo) SubmitRating(ctx context.Context, bookingID, clientID string, stars int, comment *string) (domain.Booking, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Booking{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var status, bookingClientID string
	var startAt time.Time
	var ratingStars *int
	const lockQ = `
		SELECT b.status, b.client_id, b.rating_stars, s.start_at
		FROM bookings b
		JOIN slots s ON s.id = b.slot_id
		WHERE b.id = $1
		FOR UPDATE OF b`
	if err := tx.QueryRow(ctx, lockQ, bookingID).Scan(&status, &bookingClientID, &ratingStars, &startAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Booking{}, domain.ErrBookingNotFound
		}
		return domain.Booking{}, err
	}
	if bookingClientID != clientID {
		return domain.Booking{}, domain.ErrForbidden
	}
	if ratingStars != nil {
		return domain.Booking{}, domain.ErrAlreadyRated
	}
	if !(time.Now().UTC().After(startAt) && status == string(domain.BookingActive)) {
		return domain.Booking{}, domain.ErrNotRatable
	}
	if stars < 1 || stars > 5 {
		return domain.Booking{}, domain.ErrInvalidRating
	}

	const updateQ = `UPDATE bookings SET rating_stars = $1, rating_comment = $2, rating_created_at = now() WHERE id = $3`
	if _, err := tx.Exec(ctx, updateQ, stars, comment, bookingID); err != nil {
		return domain.Booking{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Booking{}, err
	}
	return r.Get(ctx, bookingID)
}
