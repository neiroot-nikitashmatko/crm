package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	CORSOrigins []string
	JWTSecret   string
	JWTTTL      time.Duration

	BeelineAPIToken      string
	BeelineWebhookSecret string
	BeelineCreatedByUser string
}

func Load() (Config, error) {
	loadDotEnvIfExists()

	cfg := Config{
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		CORSOrigins: splitCSV(envOrDefault("CORS_ORIGINS", "http://localhost:5173")),
		JWTSecret:   strings.TrimSpace(os.Getenv("JWT_SECRET")),
		JWTTTL:      parseJWTTTL(envOrDefault("JWT_TTL_HOURS", "24")),

		BeelineAPIToken:      strings.TrimSpace(os.Getenv("BEELINE_API_TOKEN")),
		BeelineWebhookSecret: strings.TrimSpace(os.Getenv("BEELINE_WEBHOOK_SECRET")),
		BeelineCreatedByUser: strings.TrimSpace(os.Getenv("BEELINE_CREATED_BY_USER_ID")),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}

	return cfg, nil
}

func parseJWTTTL(raw string) time.Duration {
	hours, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || hours <= 0 {
		return 24 * time.Hour
	}
	return time.Duration(hours) * time.Hour
}

func loadDotEnvIfExists() {
	paths := []string{
		".env",
		filepath.Join("backend", ".env"),
	}

	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}

			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key == "" {
				continue
			}

			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, value)
			}
		}
		return
	}
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
