package services

import (
	"fmt"
	"io"
	"os"
	"simdokpol/internal/config"
	"simdokpol/internal/models"
	"strings"
	"time"
)

type BackupService interface {
	CreateBackup(actorID uint) (backupPath string, err error)
	RestoreBackup(uploadedFile io.Reader, actorID uint) error
}

type backupService struct {
	cfg           *config.Config
	configService ConfigService
	auditService  AuditLogService
}

func NewBackupService(cfg *config.Config, configService ConfigService, auditService AuditLogService) BackupService {
	return &backupService{
		cfg:           cfg,
		configService: configService,
		auditService:  auditService,
	}
}

func (s *backupService) getCleanDBPath() string {
	dsnParts := strings.Split(s.cfg.DBDSN, "?")
	return dsnParts[0]
}

func (s *backupService) CreateBackup(actorID uint) (string, error) {
	sourcePath := s.getCleanDBPath()

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file database sumber tidak ditemukan di: %s (path asli dari DSN: %s)", sourcePath, s.cfg.DBDSN)
	}

	appConfig, err := s.configService.GetConfig()
	if err != nil {
		return "", fmt.Errorf("gagal mendapatkan konfigurasi aplikasi: %w", err)
	}
	
	backupDir := appConfig.BackupPath
	if backupDir == "" {
		backupDir = "./backups"
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat direktori backup di '%s': %w", backupDir, err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	destinationPath := fmt.Sprintf("%s/backup-simdokpol-%s.db", backupDir, timestamp)

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("gagal membuka file database sumber: %w", err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file database backup: %w", err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return "", fmt.Errorf("gagal menyalin data ke file backup: %w", err)
	}

	s.auditService.LogActivity(actorID, models.AuditBackupCreated, fmt.Sprintf("Membuat file backup baru: %s", destinationPath))

	return destinationPath, nil
}

func (s *backupService) RestoreBackup(uploadedFile io.Reader, actorID uint) error {
	targetPath := s.getCleanDBPath()

	if _, err := os.Stat(targetPath); err == nil {
		preRestoreBackupPath := fmt.Sprintf("%s.before-restore-%s", targetPath, time.Now().Format("20060102150405"))
		
		sourceFile, err := os.Open(targetPath)
		if err != nil {
			return fmt.Errorf("gagal membuka database saat ini untuk backup darurat: %w", err)
		}
		defer sourceFile.Close()

		emergencyBackupFile, err := os.Create(preRestoreBackupPath)
		if err != nil {
			return fmt.Errorf("gagal membuat file backup darurat: %w", err)
		}
		defer emergencyBackupFile.Close()

		if _, err := io.Copy(emergencyBackupFile, sourceFile); err != nil {
			return fmt.Errorf("gagal menyalin untuk backup darurat: %w", err)
		}
	}

	destinationFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("gagal membuat/menimpa file database tujuan: %w", err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, uploadedFile)
	if err != nil {
		return fmt.Errorf("gagal menyalin data dari file yang diunggah: %w", err)
	}

	s.auditService.LogActivity(actorID, models.AuditRestoreFromFile, "Database dipulihkan dari file backup.")

	return nil
}