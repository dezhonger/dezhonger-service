package app

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL  string
	ListenAddr   string
	PublicOrigin string
	ServerRegion string
	CookieName   string
	CookieSecure bool
	SessionTTL   time.Duration
}

func LoadConfig() (Config, error) {
	cfg := Config{
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		ListenAddr:   envOr("LISTEN_ADDR", ":8080"),
		PublicOrigin: os.Getenv("PUBLIC_ORIGIN"),
		ServerRegion: envOr("SERVER_REGION", "Tencent Cloud · Hong Kong"),
		CookieName:   envOr("SESSION_COOKIE_NAME", "dz_session"),
		CookieSecure: envBool("COOKIE_SECURE", true),
		SessionTTL:   envDuration("SESSION_TTL", 30*24*time.Hour),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.SessionTTL < time.Hour {
		return Config{}, errors.New("SESSION_TTL must be at least one hour")
	}
	return cfg, nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
