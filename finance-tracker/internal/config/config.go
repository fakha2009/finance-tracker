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
}

func Load() (*Config, error) {
	port := getEnv("PORT", "8080")
	databaseURL := getEnv("DATABASE_URL", "host=localhost port=5432 user=postgres password=fakha dbname=transactions sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production")
	env := getEnv("GIN_MODE", "debug")
	exchangeAPIEndpoint := getEnv("EXCHANGE_API_ENDPOINT", "https://api.exchangerate-api.com/v4/latest/USD")

	return &Config{
		Port:                port,
		DatabaseURL:         databaseURL,
		JWTSecret:           jwtSecret,
		Env:                 env,
		ExchangeAPIEndpoint: exchangeAPIEndpoint,
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
