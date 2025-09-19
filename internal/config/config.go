package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	AppPort int

	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret string

	RatesProvider             string
	RatesRefreshIntervalHours int
}

func Load() Config {
	cfg := Config{}
	cfg.AppPort = intFromEnv("APP_PORT", 8080)
	cfg.DBHost = getenv("DB_HOST", "localhost")
	cfg.DBPort = intFromEnv("DB_PORT", 5432)
	cfg.DBUser = getenv("DB_USER", "postgres")
	cfg.DBPassword = getenv("DB_PASSWORD", "postgres")
	cfg.DBName = getenv("DB_NAME", "currency")
	cfg.DBSSLMode = getenv("DB_SSLMODE", "disable")
	cfg.JWTSecret = getenv("JWT_SECRET", "changeme")
	cfg.RatesProvider = getenv("RATES_PROVIDER", "frankfurter")
	cfg.RatesRefreshIntervalHours = intFromEnv("RATES_REFRESH_INTERVAL_HOURS", 6)
	return cfg
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
func intFromEnv(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("Invalid int env %s=%s", key, v)
		return def
	}
	return i
}
