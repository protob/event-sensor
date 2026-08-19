package config

import (
	"os"
	"strings"
)

const ticketmasterKeyPath = "/run/agenix/ticketmaster-api-key"

const defaultJWTSecret = "dev-secret-change-in-production"

// Config holds application configuration loaded from environment variables.
type Config struct {
	Bind                string
	Port                string
	DBPath              string
	TicketmasterAPIKey  string
	TicketmasterBaseURL string
	JWTSecret           string
}

// Load reads configuration from environment variables with sensible defaults.
// Bind defaults to loopback: on a laptop nothing changes, and on a server the app
// sits behind Caddy rather than on a public interface.
func Load() Config {
	return Config{
		Bind:               getEnv("ES_BIND", "127.0.0.1"),
		Port:               getEnv("PORT", "8080"),
		DBPath:             getEnv("DB_PATH", "data/event-sensor.db"),
		TicketmasterAPIKey: getTicketmasterKey(),
		// Empty means the real API; the client substitutes its default.
		TicketmasterBaseURL: getEnv("TM_BASE_URL", ""),
		JWTSecret:           getEnv("JWT_SECRET", defaultJWTSecret),
	}
}

// IsLoopbackBind reports whether the bind address is loopback-only, i.e. the kernel
// refuses off-machine connections.
func (c Config) IsLoopbackBind() bool {
	switch c.Bind {
	case "127.0.0.1", "::1", "localhost":
		return true
	default:
		return false
	}
}

// UsingDefaultSecret reports whether the JWT secret is still the public fallback
// from this repository. Fine on loopback; fatal elsewhere (see the startup gate
// in main.go).
func (c Config) UsingDefaultSecret() bool {
	return c.JWTSecret == defaultJWTSecret
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

// getTicketmasterKey reads from agenix file first, falls back to env var.
func getTicketmasterKey() string {
	if data, err := os.ReadFile(ticketmasterKeyPath); err == nil {
		return strings.TrimSpace(string(data))
	}
	return getEnv("TICKETMASTER_API_KEY", "")
}
