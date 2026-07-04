package domain

import "time"

// BookingStatus — bookings/models.yaml#/components/schemas/BookingStatus. cancelled_by_studio
// is a server-initiated transition (studio-side slot cancellation) that nothing in this
// package's write paths sets, but the type still has to represent it on read.
type BookingStatus string

const (
	BookingActive            BookingStatus = "active"
	BookingCancelled         BookingStatus = "cancelled"
	BookingLateCancel        BookingStatus = "late_cancel"
	BookingCancelledByStudio BookingStatus = "cancelled_by_studio"
)

// EquipmentChoice — bookings/models.yaml#/components/schemas/EquipmentChoice. rental is never
// rejected for capacity reasons — the rental fund is unconstrained (FR-9).
type EquipmentChoice string

const (
	EquipmentOwn    EquipmentChoice = "own"
	EquipmentRental EquipmentChoice = "rental"
)

// Rating — data-model.md → Booking.rating. Embedded 1:1 in Booking, not a standalone resource
// (FR-19/FR-20); settable at most once, never edited afterwards.
type Rating struct {
	Stars     int
	Comment   *string
	CreatedAt time.Time
}

// Booking — data-model.md → Booking. FreeCancellationUntil is never stored: it's computed
// (slot.start_at − 24h) at read time from the joined Slot and populated by the repo, never by
// the client (FR-15/FR-16).
type Booking struct {
	ID                    string
	SlotID                string
	ClientID              string
	EquipmentChoice       EquipmentChoice
	Status                BookingStatus
	PriceTotal            int
	FreeCancellationUntil time.Time
	Rating                *Rating
	CreatedAt             time.Time
	CancelledAt           *time.Time
	Slot                  Slot
}
