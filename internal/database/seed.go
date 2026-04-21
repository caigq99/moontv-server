package database

import (
	"log"

	"github.com/moontv/server/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func Seed(adminUsername, adminPassword string) {
	var count int64
	DB.Model(&model.User{}).Where("username = ?", adminUsername).Count(&count)
	if count > 0 {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash admin password: %v", err)
	}

	admin := model.User{
		Username:     adminUsername,
		PasswordHash: string(hash),
		Role:         "admin",
	}
	if err := DB.Create(&admin).Error; err != nil {
		log.Fatalf("failed to create admin user: %v", err)
	}
	log.Printf("admin user '%s' created", adminUsername)
}
