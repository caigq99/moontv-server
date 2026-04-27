package main

import (
	"fmt"
	"log"

	"github.com/moontv/server/internal/config"
	"github.com/moontv/server/internal/database"
	"github.com/moontv/server/internal/router"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("config error: %v", err)
	}

	database.Init(cfg.DBPath)
	database.Seed(cfg.AdminUsername, cfg.AdminPassword)

	r := router.Setup(cfg.JWTSecret, []byte(cfg.APIKeySecret), cfg.APIKeyPrefix)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
