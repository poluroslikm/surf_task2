package domain

import "time"

// PushSubscription — push/models.yaml#/components/schemas/PushSubscription, plus the owning
// client server-side. Identified by Endpoint (globally unique, DB-enforced) — ClientID is the
// mutable "owner" field: the same endpoint upserts to whichever client last registered it (the
// shared-device re-registration case), it is never rejected as a duplicate (FEAT-03).
type PushSubscription struct {
	ID        string
	ClientID  string
	Endpoint  string
	P256dh    string
	Auth      string
	CreatedAt time.Time
}
