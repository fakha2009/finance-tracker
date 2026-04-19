package config

import (
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
	databaseURL := getEnv("DATABASE_URL", "")
	jwtSecret := getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production")
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

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
