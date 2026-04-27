package repository

import (
	"time"

	"github.com/moontv/server/internal/database"
	"github.com/moontv/server/internal/model"
	"gorm.io/gorm"
)

func RegisterUser(username, passwordHash, inviteCode string) (*model.User, error) {
	var user *model.User

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var invite model.InviteCode
		if err := tx.Where("code = ?", inviteCode).First(&invite).Error; err != nil {
			return err
		}
		if invite.UsedBy != nil {
			return ErrInviteUsed
		}
		if invite.ExpiresAt != nil && time.Now().After(*invite.ExpiresAt) {
			return ErrInviteExpired
		}

		var existing model.User
		if err := tx.Where("username = ?", username).First(&existing).Error; err == nil {
			return ErrUsernameExists
		}

		user = &model.User{
			Username:     username,
			PasswordHash: passwordHash,
			Role:         "user",
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		now := time.Now()
		if err := tx.Model(&model.InviteCode{}).Where("code = ?", inviteCode).Updates(map[string]any{
			"used_by": user.ID,
			"used_at": now,
		}).Error; err != nil {
			return err
		}

		var globals []model.Source
		if err := tx.Where("user_id IS NULL").Order("sort_order asc, id asc").Find(&globals).Error; err != nil {
			return err
		}
		for _, g := range globals {
			s := model.Source{
				UserID:    &user.ID,
				Key:       g.Key,
				Name:      g.Name,
				APIUrl:    g.APIUrl,
				DetailUrl: g.DetailUrl,
				SortOrder: g.SortOrder,
			}
			if err := tx.Create(&s).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return user, err
}

var ErrInviteUsed = &registerError{"invite code already used"}
var ErrInviteExpired = &registerError{"invite code expired"}
var ErrUsernameExists = &registerError{"username already exists"}

type registerError struct{ msg string }

func (e *registerError) Error() string { return e.msg }
