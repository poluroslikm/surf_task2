package domain

import "time"

// Client — data-model.md → Client. PasswordHash is populated on read but must never be
// serialized back to the API (see httpapi/dto, which has no such field).
type Client struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
