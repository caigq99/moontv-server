package repository

import (
	"time"

	"github.com/moontv/server/internal/database"
	"github.com/moontv/server/internal/model"
)

func CountUsers() int64 {
	var count int64
	database.DB.Model(&model.User{}).Count(&count)
	return count
}

func CountActiveUsers() int64 {
	var count int64
	database.DB.Model(&model.User{}).Where("banned = ? AND api_key_cipher != ''", false).Count(&count)
	return count
}

func CountSources() int64 {
	var count int64
	database.DB.Model(&model.Source{}).Count(&count)
	return count
}

func CountAPICallsSince(since time.Time) int64 {
	var count int64
	database.DB.Model(&model.APICallLog{}).Where("created_at >= ?", since).Count(&count)
	return count
}
