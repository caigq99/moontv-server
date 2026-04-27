package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port          int
	DBPath        string
	JWTSecret     string
	APIKeySecret  string // 32 bytes for AES-256-GCM
	APIKeyPrefix  string
	AdminUsername string
	AdminPassword string
}

func Load() *Config {
	cfg := &Config{
		Port:          getEnvInt("PORT", 8080),
		DBPath:        getEnv("DB_PATH", "./data/moontv.db"),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		APIKeySecret:  getEnv("APIKEY_SECRET", ""),
		APIKeyPrefix:  getEnv("APIKEY_PREFIX", "mtv_"),
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
	}
	return cfg
}

func (c *Config) Validate() error {
	var errs []string
	if c.JWTSecret == "" {
		errs = append(errs, "JWT_SECRET is required")
	}
	if len(c.APIKeySecret) < 16 {
		errs = append(errs, "APIKEY_SECRET must be at least 16 characters")
	}
	if c.AdminPassword == "" {
		errs = append(errs, "ADMIN_PASSWORD is required")
	}
	if len(errs) > 0 {
		return fmt.Errorf("invalid config: %v", errs)
	}
	return nil
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
