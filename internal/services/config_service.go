package services

import (
	"simdokpol/internal/dto" // <-- IMPORT BARU
	"simdokpol/internal/repositories"
	"strconv"
	"time"

	"gorm.io/gorm"
)

const IsSetupCompleteKey = "is_setup_complete"

// DEFINISI AppConfig DIPINDAHKAN KE internal/dto/config_dto.go

type ConfigService interface {
	IsSetupComplete() (bool, error)
	GetConfig() (*dto.AppConfig, error) // <-- DIUBAH
	SaveConfig(configData map[string]string) error
	GetLocation() (*time.Location, error)
}

type configService struct {
	configRepo     repositories.ConfigRepository
	cachedLocation *time.Location
	cachedConfig   *dto.AppConfig // <-- DIUBAH
}

func NewConfigService(configRepo repositories.ConfigRepository) ConfigService {
	return &configService{configRepo: configRepo}
}

func (s *configService) SaveConfig(configData map[string]string) error {
	s.cachedLocation = nil
	s.cachedConfig = nil 
	return s.configRepo.SetMultiple(configData)
}

func (s *configService) GetLocation() (*time.Location, error) {
	if s.cachedLocation != nil {
		return s.cachedLocation, nil
	}

	config, err := s.GetConfig()
	if err != nil {
		return time.UTC, err
	}

	if config.ZonaWaktu == "" {
		return time.UTC, nil
	}

	loc, err := time.LoadLocation(config.ZonaWaktu)
	if err != nil {
		return time.UTC, err
	}

	s.cachedLocation = loc
	return s.cachedLocation, nil
}

func (s *configService) IsSetupComplete() (bool, error) {
	config, err := s.configRepo.Get(IsSetupCompleteKey)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return config.Value == "true", nil
}

func (s *configService) GetConfig() (*dto.AppConfig, error) { // <-- DIUBAH
	if s.cachedConfig != nil {
		return s.cachedConfig, nil
	}

	allConfigs, err := s.configRepo.GetAll()
	if err != nil {
		return nil, err
	}

	archiveDays, _ := strconv.Atoi(allConfigs["archive_duration_days"])

	// Gunakan dto.AppConfig
	appConfig := &dto.AppConfig{
		IsSetupComplete:     allConfigs[IsSetupCompleteKey] == "true",
		KopBaris1:           allConfigs["kop_baris_1"],
		KopBaris2:           allConfigs["kop_baris_2"],
		KopBaris3:           allConfigs["kop_baris_3"],
		NamaKantor:          allConfigs["nama_kantor"],
		TempatSurat:         allConfigs["tempat_surat"],
		FormatNomorSurat:    allConfigs["format_nomor_surat"],
		NomorSuratTerakhir:  allConfigs["nomor_surat_terakhir"],
		ZonaWaktu:           allConfigs["zona_waktu"],
		BackupPath:          allConfigs["backup_path"],
		ArchiveDurationDays: archiveDays,
	}

	s.cachedConfig = appConfig
	return appConfig, nil
}