package repository

import (
	"github.com/moontv/server/internal/database"
	"github.com/moontv/server/internal/model"
	"gorm.io/gorm"
)

func GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func CreateUser(user *model.User) error {
	return database.DB.Create(user).Error
}

func UpdateUser(user *model.User) error {
	return database.DB.Save(user).Error
}

func DeleteUser(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&model.Source{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.User{}, id).Error
	})
}

func ListUsers(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	database.DB.Model(&model.User{}).Count(&total)
	err := database.DB.Offset((page - 1) * pageSize).Limit(pageSize).Order("id asc").Find(&users).Error
	return users, total, err
}
