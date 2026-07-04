package domain

import "errors"

// Sentinel errors mapped to contractual HTTP responses in internal/httpapi/errors.go.
var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrSlotNotFound       = errors.New("slot not found")

	// Bookings domain (BE-06..BE-09).

	// ErrInvalidSlotID is createBooking's "slot_id doesn't exist" case. Unlike getSlot,
	// createBooking's contract has no documented 404 response, so this maps to 400
	// bad_request rather than reusing ErrSlotNotFound (which is wired to 404 elsewhere).
	ErrInvalidSlotID    = errors.New("slot does not exist")
	ErrSlotCancelled    = errors.New("slot cancelled")
	ErrBookingNotFound  = errors.New("booking not found")
	ErrForbidden        = errors.New("forbidden")
	ErrAlreadyCancelled = errors.New("booking already cancelled")
	ErrSlotStarted      = errors.New("slot already started")
	ErrAlreadyRated     = errors.New("booking already rated")
	ErrNotRatable       = errors.New("booking not ratable")
	ErrInvalidRating    = errors.New("rating stars out of range")
	// ErrDuplicateBooking: client already has an active booking on this slot (unique index
	// bookings_active_client_slot_uidx). No dedicated contract code exists for this — see
	// httpapi/errors.go for the code chosen.
	ErrDuplicateBooking = errors.New("client already has an active booking on this slot")
)

// ErrSlotFull carries the actual free-seats count for createBooking's 409 response
// (details.available_seats, common/models.yaml) — the one domain error needing dynamic data,
// so it's a type rather than a plain sentinel.
type ErrSlotFull struct {
	AvailableSeats int
}

func (e *ErrSlotFull) Error() string { return "slot full" }
