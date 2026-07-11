package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        int
	DatabaseURL string
	SessionKey  string
	Environment string
	MaxWorkers  int
	LogLevel    string
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("PORT", "8080"))
	maxWorkers, _ := strconv.Atoi(getEnv("MAX_WORKERS", "10"))
	return &Config{
		Port:        port,
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost:5432/oj?sslmode=disable"),
		SessionKey:  getEnv("SESSION_KEY", "super-secret-key-change-me"),
		Environment: getEnv("ENVIRONMENT", "development"),
		MaxWorkers:  maxWorkers,
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
