// Package config loads application configuration from environment variables.
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration for the server.
type Config struct {
	Port string

	PostgresDSN string
	MongoURI    string
	MongoDBName string

	JWTSecret     string
	HMACSecret    string
	AdminEmail    string
	AdminPassword string

	FrontendOrigin string
}

// Load reads a .env file if present (dev convenience) and builds a Config
// from environment variables, returning an error if a required var is missing.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:          getEnv("PORT", "8080"),
		PostgresDSN:   os.Getenv("POSTGRES_DSN"),
		MongoURI:      os.Getenv("MONGO_URI"),
		MongoDBName:   getEnv("MONGO_DB_NAME", "user_records"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		HMACSecret:    os.Getenv("HMAC_SECRET"),
		AdminEmail:    os.Getenv("ADMIN_EMAIL"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),

		FrontendOrigin: getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),
	}

	required := map[string]string{
		"POSTGRES_DSN":   cfg.PostgresDSN,
		"MONGO_URI":      cfg.MongoURI,
		"JWT_SECRET":     cfg.JWTSecret,
		"HMAC_SECRET":    cfg.HMACSecret,
		"ADMIN_EMAIL":    cfg.AdminEmail,
		"ADMIN_PASSWORD": cfg.AdminPassword,
	}
	for name, val := range required {
		if val == "" {
			return nil, fmt.Errorf("config: missing required env var %s", name)
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
