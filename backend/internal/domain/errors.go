package domain

import "errors"

// Sentinel errors mapped to contractual HTTP responses in internal/httpapi/errors.go.
var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrSlotNotFound       = errors.New("slot not found")
)
