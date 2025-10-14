package services

import (
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
	"time"
)

type AuditLogService interface {
	LogActivity(userID uint, action string, details string)
	FindAll() ([]models.AuditLog, error)
}

type auditLogService struct {
	repo repositories.AuditLogRepository
}

func NewAuditLogService(repo repositories.AuditLogRepository) AuditLogService {
	return &auditLogService{repo: repo}
}

// LogActivity berjalan sebagai goroutine agar tidak memblokir proses utama.
func (s *auditLogService) LogActivity(userID uint, action string, details string) {
	go func() {
		logEntry := &models.AuditLog{
			UserID:    userID,
			Aksi:      action,
			Detail:    details,
			Timestamp: time.Now(),
		}
		_ = s.repo.Create(logEntry)
	}()
}

func (s *auditLogService) FindAll() ([]models.AuditLog, error) {
	return s.repo.FindAll()
}