package model

import "time"

type InviteCode struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Code      string     `gorm:"uniqueIndex;size:32;not null" json:"code"`
	CreatedBy uint       `gorm:"not null" json:"created_by"`
	UsedBy    *uint      `json:"used_by"`
	UsedAt    *time.Time `json:"used_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}
