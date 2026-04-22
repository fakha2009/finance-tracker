package config

import (
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	Port                string
	DatabaseURL         string
	JWTSecret           string
	Env                 string
	ExchangeAPIEndpoint string
	AllowedOrigin       string
	CronSecret          string
}

func Load() (*Config, error) {
	port := getEnv("PORT", "8080")
	databaseURL := firstNonEmpty(
		os.Getenv("DATABASE_URL"),
		os.Getenv("POSTGRES_URL_NON_POOLING"),
		os.Getenv("POSTGRES_URL"),
		os.Getenv("POSTGRES_PRISMA_URL"),
	)

	if databaseURL == "" {
		databaseURL = buildDatabaseURL(
			getEnv("POSTGRES_USER", "postgres"),
			getEnv("POSTGRES_PASSWORD", ""),
			getEnv("POSTGRES_HOST", "localhost"),
			getEnv("POSTGRES_PORT", "5432"),
			getEnv("POSTGRES_DATABASE", "postgres"),
			getEnv("PGSSLMODE", "require"),
		)
	}

	jwtSecret := firstNonEmpty(
		getEnv("JWT_SECRET", ""),
		getEnv("SUPABASE_JWT_SECRET", ""),
	)
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key-change-in-production"
	}

	env := getEnv("GIN_MODE", "debug")
	exchangeAPIEndpoint := getEnv("EXCHANGE_API_ENDPOINT", "https://api.exchangerate-api.com/v4/latest/USD")
	allowedOrigin := getEnv("ALLOWED_ORIGIN", "http://localhost:3000")
	cronSecret := getEnv("CRON_SECRET", "default-cron-secret-change-in-prod")

	return &Config{
		Port:                port,
		DatabaseURL:         databaseURL,
		JWTSecret:           jwtSecret,
		Env:                 env,
		ExchangeAPIEndpoint: exchangeAPIEndpoint,
		AllowedOrigin:       allowedOrigin,
		CronSecret:          cronSecret,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func buildDatabaseURL(user, password, host, port, database, sslmode string) string {
	if user == "" || host == "" || port == "" || database == "" {
		return ""
	}

	escapedPassword := url.QueryEscape(password)
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, escapedPassword, host, port, database, sslmode)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
