package repositories

import (
	"simdokpol/internal/models"

	"gorm.io/gorm"
)

type AuditLogRepository interface {
	Create(log *models.AuditLog) error
	FindAll() ([]models.AuditLog, error)
}

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *auditLogRepository) FindAll() ([]models.AuditLog, error) {
	var logs []models.AuditLog
	// Preload User untuk mendapatkan data pengguna yang melakukan aksi
	err := r.db.Preload("User").Order("timestamp desc").Find(&logs).Error
	return logs, err
}