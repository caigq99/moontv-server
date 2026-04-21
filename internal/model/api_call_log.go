package model

import "time"

type APICallLog struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index;not null"`
	Endpoint  string    `gorm:"size:128;not null"`
	IP        string    `gorm:"size:45"`
	CreatedAt time.Time `gorm:"index"`
}
