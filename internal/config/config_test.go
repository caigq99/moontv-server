package config

import (
	"os"
	"testing"
)

func TestValidateRejectsEmptyJWTSecret(t *testing.T) {
	cfg := &Config{JWTSecret: "", APIKeySecret: "0123456789abcdef", AdminPassword: "pw"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty JWT_SECRET")
	}
}

func TestValidateRejectsShortAPIKeySecret(t *testing.T) {
	cfg := &Config{JWTSecret: "secret", APIKeySecret: "short", AdminPassword: "pw"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for short APIKEY_SECRET")
	}
}

func TestValidateRejectsEmptyAdminPassword(t *testing.T) {
	cfg := &Config{JWTSecret: "secret", APIKeySecret: "0123456789abcdef", AdminPassword: ""}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty ADMIN_PASSWORD")
	}
}

func TestValidateAcceptsValidConfig(t *testing.T) {
	cfg := &Config{JWTSecret: "secret", APIKeySecret: "0123456789abcdef", AdminPassword: "pw"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestLoadReadsEnvVars(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("JWT_SECRET", "envsecret")
	os.Setenv("APIKEY_SECRET", "envapikey1234567")
	os.Setenv("ADMIN_PASSWORD", "envpw")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("APIKEY_SECRET")
		os.Unsetenv("ADMIN_PASSWORD")
	}()

	cfg := Load()
	if cfg.Port != 9090 {
		t.Fatalf("expected port 9090, got %d", cfg.Port)
	}
	if cfg.JWTSecret != "envsecret" {
		t.Fatalf("expected jwt secret from env, got %q", cfg.JWTSecret)
	}
}
