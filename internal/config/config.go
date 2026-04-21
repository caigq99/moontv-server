package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          int
	DBPath        string
	JWTSecret     string
	APIKeySecret  string // 32 bytes for AES-256-GCM
	AdminUsername string
	AdminPassword string
}

func Load() *Config {
	return &Config{
		Port:          getEnvInt("PORT", 8080),
		DBPath:        getEnv("DB_PATH", "./data/moontv.db"),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		APIKeySecret:  getEnv("APIKEY_SECRET", ""),
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
