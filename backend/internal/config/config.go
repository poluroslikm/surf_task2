package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	SessionTTL  time.Duration

	// Push (FEAT-03/04/05). All four are required — fail Load() the same way DATABASE_URL
	// does — because without them push sending (registration, reminders, force-cancel) can't
	// function at all; there is no meaningful zero-value fallback for a VAPID key pair or a
	// shared internal-tools secret.
	VAPIDPublicKey     string
	VAPIDPrivateKey    string
	VAPIDSubject       string
	InternalToolsToken string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:           getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		SessionTTL:         30 * 24 * time.Hour,
		VAPIDPublicKey:     os.Getenv("VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey:    os.Getenv("VAPID_PRIVATE_KEY"),
		VAPIDSubject:       os.Getenv("VAPID_SUBJECT"),
		InternalToolsToken: os.Getenv("INTERNAL_TOOLS_TOKEN"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.VAPIDPublicKey == "" {
		return Config{}, fmt.Errorf("VAPID_PUBLIC_KEY is required")
	}
	if cfg.VAPIDPrivateKey == "" {
		return Config{}, fmt.Errorf("VAPID_PRIVATE_KEY is required")
	}
	if cfg.VAPIDSubject == "" {
		return Config{}, fmt.Errorf("VAPID_SUBJECT is required")
	}
	if cfg.InternalToolsToken == "" {
		return Config{}, fmt.Errorf("INTERNAL_TOOLS_TOKEN is required")
	}
	if v := os.Getenv("SESSION_TTL_HOURS"); v != "" {
		hours, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid SESSION_TTL_HOURS: %w", err)
		}
		cfg.SessionTTL = time.Duration(hours) * time.Hour
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
