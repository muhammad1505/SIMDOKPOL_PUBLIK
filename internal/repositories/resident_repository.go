package repositories

import (
	"simdokpol/internal/models"

	"gorm.io/gorm"
)

// ResidentRepository mendefinisikan kontrak untuk operasi data penduduk.
type ResidentRepository interface {
	// FindByNIK mencari penduduk berdasarkan NIK. Menggunakan transaksi jika disediakan.
	FindByNIK(tx *gorm.DB, nik string) (*models.Resident, error)
	// Create menyimpan data penduduk baru. Menggunakan transaksi jika disediakan.
	Create(tx *gorm.DB, resident *models.Resident) (*models.Resident, error)
}

type residentRepository struct {
	db *gorm.DB
}

// NewResidentRepository adalah factory untuk ResidentRepository.
func NewResidentRepository(db *gorm.DB) ResidentRepository {
	return &residentRepository{db: db}
}

func (r *residentRepository) FindByNIK(tx *gorm.DB, nik string) (*models.Resident, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	var resident models.Resident
	if err := db.Where("nik = ?", nik).First(&resident).Error; err != nil {
		return nil, err
	}
	return &resident, nil
}

func (r *residentRepository) Create(tx *gorm.DB, resident *models.Resident) (*models.Resident, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	if err := db.Create(resident).Error; err != nil {
		return nil, err
	}
	return resident, nil
}