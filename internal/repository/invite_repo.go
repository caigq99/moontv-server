package repository

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/moontv/server/internal/database"
	"github.com/moontv/server/internal/model"
)

func GenerateInviteCodes(createdBy uint, count int, expiresDays int) ([]model.InviteCode, error) {
	var codes []model.InviteCode
	var expiresAt *time.Time
	if expiresDays > 0 {
		t := time.Now().AddDate(0, 0, expiresDays)
		expiresAt = &t
	}

	for i := 0; i < count; i++ {
		b := make([]byte, 8)
		if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("generate random code: %w", err)
	}
		code := model.InviteCode{
			Code:      hex.EncodeToString(b),
			CreatedBy: createdBy,
			ExpiresAt: expiresAt,
		}
		if err := database.DB.Create(&code).Error; err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, nil
}

func GetInviteByCode(code string) (*model.InviteCode, error) {
	var invite model.InviteCode
	if err := database.DB.Where("code = ?", code).First(&invite).Error; err != nil {
		return nil, err
	}
	return &invite, nil
}

func MarkInviteUsed(code string, usedBy uint) error {
	now := time.Now()
	return database.DB.Model(&model.InviteCode{}).Where("code = ?", code).Updates(map[string]any{
		"used_by": usedBy,
		"used_at": now,
	}).Error
}

func ListInvites(page, pageSize int) ([]model.InviteCode, int64, error) {
	var codes []model.InviteCode
	var total int64
	if err := database.DB.Model(&model.InviteCode{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := database.DB.Offset((page-1)*pageSize).Limit(pageSize).Order("created_at desc").Find(&codes).Error
	return codes, total, err
}

func DeleteInvite(code string) error {
	return database.DB.Where("code = ?", code).Delete(&model.InviteCode{}).Error
}
