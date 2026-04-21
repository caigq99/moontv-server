package database

import (
	"log"
	"os"
	"path/filepath"

	"github.com/moontv/server/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(dbPath string) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("failed to create db directory: %v", err)
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL&_busy_timeout=5000"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	if err := DB.AutoMigrate(
		&model.User{},
		&model.Source{},
		&model.InviteCode{},
		&model.APICallLog{},
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
}
