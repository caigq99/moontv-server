package repository

import (
	"github.com/moontv/server/internal/database"
	"github.com/moontv/server/internal/model"
)

func GetSourcesByUserID(userID uint) ([]model.Source, error) {
	var sources []model.Source
	err := database.DB.Where("user_id = ?", userID).Order("sort_order asc, id asc").Find(&sources).Error
	if err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		err = database.DB.Where("user_id IS NULL").Order("sort_order asc, id asc").Find(&sources).Error
	}
	return sources, err
}

func GetEnabledSourcesByUserID(userID uint) ([]model.Source, error) {
	var sources []model.Source
	err := database.DB.Where("user_id = ? AND disabled = ?", userID, false).Order("sort_order asc, id asc").Find(&sources).Error
	if err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		err = database.DB.Where("user_id IS NULL AND disabled = ?", false).Order("sort_order asc, id asc").Find(&sources).Error
	}
	return sources, err
}

func GetGlobalSources() ([]model.Source, error) {
	var sources []model.Source
	err := database.DB.Where("user_id IS NULL").Order("sort_order asc, id asc").Find(&sources).Error
	return sources, err
}

func GetSourceByKey(userID uint, key string) (*model.Source, error) {
	var source model.Source
	if err := database.DB.Where("user_id = ? AND key = ?", userID, key).First(&source).Error; err != nil {
		return nil, err
	}
	return &source, nil
}

func GetGlobalSourceByKey(key string) (*model.Source, error) {
	var source model.Source
	if err := database.DB.Where("user_id IS NULL AND key = ?", key).First(&source).Error; err != nil {
		return nil, err
	}
	return &source, nil
}

func CreateSource(source *model.Source) error {
	return database.DB.Create(source).Error
}

func UpdateSource(source *model.Source) error {
	return database.DB.Save(source).Error
}

func DeleteSource(userID uint, key string) error {
	return database.DB.Where("user_id = ? AND key = ?", userID, key).Delete(&model.Source{}).Error
}

func DeleteGlobalSource(key string) error {
	return database.DB.Where("user_id IS NULL AND key = ?", key).Delete(&model.Source{}).Error
}

func CopyGlobalSourcesToUser(userID uint) error {
	globals, err := GetGlobalSources()
	if err != nil {
		return err
	}
	for _, g := range globals {
		s := model.Source{
			UserID:    &userID,
			Key:       g.Key,
			Name:      g.Name,
			APIUrl:    g.APIUrl,
			DetailUrl: g.DetailUrl,
			SortOrder: g.SortOrder,
		}
		database.DB.Create(&s)
	}
	return nil
}

func UpdateSourceSortOrder(userID uint, keys []string) error {
	for i, key := range keys {
		database.DB.Model(&model.Source{}).Where("user_id = ? AND key = ?", userID, key).Update("sort_order", i)
	}
	return nil
}

func UpdateGlobalSourceSortOrder(keys []string) error {
	for i, key := range keys {
		database.DB.Model(&model.Source{}).Where("user_id IS NULL AND key = ?", key).Update("sort_order", i)
	}
	return nil
}
