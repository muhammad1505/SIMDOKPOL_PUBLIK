package repositories

import (
	"simdokpol/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause" // <-- IMPORT YANG HILANG DITAMBAHKAN DI SINI
)

type ConfigRepository interface {
	Get(key string) (*models.Configuration, error)
	GetAll() (map[string]string, error)
	Set(key, value string) error
	SetMultiple(configs map[string]string) error
}

type configRepository struct {
	db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) ConfigRepository {
	return &configRepository{db: db}
}

func (r *configRepository) Get(key string) (*models.Configuration, error) {
	var config models.Configuration
	if err := r.db.Where("key = ?", key).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *configRepository) GetAll() (map[string]string, error) {
	var configs []models.Configuration
	if err := r.db.Find(&configs).Error; err != nil {
		return nil, err
	}
	configMap := make(map[string]string)
	for _, c := range configs {
		configMap[c.Key] = c.Value
	}
	return configMap, nil
}

func (r *configRepository) Set(key, value string) error {
	// Menggunakan `clause.OnConflict` yang benar
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&models.Configuration{Key: key, Value: value}).Error
}

func (r *configRepository) SetMultiple(configs map[string]string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for key, value := range configs {
			config := models.Configuration{Key: key, Value: value}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value"}),
			}).Create(&config).Error; err != nil {
				return err
			}
		}
		return nil
	})
}