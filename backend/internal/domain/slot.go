package domain

import "time"

// SlotStatus — slots/models.yaml#/components/schemas/SlotStatus.
type SlotStatus string

const (
	SlotScheduled SlotStatus = "scheduled"
	SlotCancelled SlotStatus = "cancelled"
)

// Slot — data-model.md → Slot. TotalSeats/FreeSeats are the source of truth for capacity;
// the 8/12-seat business rule is enforced when a slot is created, not re-derived here or on
// the client (FR-12).
type Slot struct {
	ID                 string
	Program            Program
	Chef               Chef
	StartAt            time.Time
	TotalSeats         int
	FreeSeats          int
	Price              int
	Status             SlotStatus
	CancellationReason *string
}
