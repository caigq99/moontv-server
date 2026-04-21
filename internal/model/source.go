package model

import "time"

type Source struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    *uint     `gorm:"index" json:"user_id"` // nil = global source
	Key       string    `gorm:"size:64;not null" json:"key"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	APIUrl    string    `gorm:"size:512;not null" json:"api_url"`
	DetailUrl string    `gorm:"size:512" json:"detail_url,omitempty"`
	Disabled  bool      `gorm:"default:false" json:"disabled"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

func (Source) TableName() string {
	return "sources"
}
