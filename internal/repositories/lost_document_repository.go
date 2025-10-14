package repositories

import (
	"fmt"
	"simdokpol/internal/models"
	"time"

	"gorm.io/gorm"
)

type MonthlyCount struct {
	Year  int `gorm:"column:year"`
	Month int `gorm:"column:month"`
	Count int `gorm:"column:count"`
}

type ItemCompositionStat struct {
	NamaBarang string `gorm:"column:nama_barang"`
	Count      int    `gorm:"column:count"`
}

type LostDocumentRepository interface {
	Create(tx *gorm.DB, doc *models.LostDocument) (*models.LostDocument, error)
	FindByID(id uint) (*models.LostDocument, error)
	FindAll(query string, statusFilter string, archiveDurationDays int) ([]models.LostDocument, error)
	SearchGlobal(query string) ([]models.LostDocument, error)
	Update(tx *gorm.DB, doc *models.LostDocument) (*models.LostDocument, error)
	Delete(tx *gorm.DB, id uint) error
	GetLastDocumentOfYear(year int) (*models.LostDocument, error)
	CountByDateRange(start time.Time, end time.Time) (int64, error)
	GetMonthlyIssuanceForYear(year int) ([]MonthlyCount, error)
	GetItemCompositionStats() ([]ItemCompositionStat, error)
	FindExpiringDocumentsForUser(userID uint, expiryDateStart time.Time, expiryDateEnd time.Time) ([]models.LostDocument, error) // <-- METHOD BARU
}

type lostDocumentRepository struct {
	db *gorm.DB
}

func NewLostDocumentRepository(db *gorm.DB) LostDocumentRepository {
	return &lostDocumentRepository{db: db}
}

// === FUNGSI BARU UNTUK NOTIFIKASI ===
func (r *lostDocumentRepository) FindExpiringDocumentsForUser(userID uint, expiryDateStart time.Time, expiryDateEnd time.Time) ([]models.LostDocument, error) {
	var docs []models.LostDocument
	err := r.db.
		Where("operator_id = ?", userID).
		Where("tanggal_laporan BETWEEN ? AND ?", expiryDateStart, expiryDateEnd).
		Order("tanggal_laporan asc").
		Find(&docs).Error
	return docs, err
}
// === AKHIR FUNGSI BARU ===

func (r *lostDocumentRepository) CountByDateRange(start time.Time, end time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.LostDocument{}).Where("tanggal_laporan BETWEEN ? AND ?", start, end).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *lostDocumentRepository) FindAll(query string, statusFilter string, archiveDurationDays int) ([]models.LostDocument, error) {
	var docs []models.LostDocument
	db := r.db.
		Preload("Resident").
		Preload("LostItems").
		Preload("PetugasPelapor").
		Preload("PejabatPersetuju").
		Preload("Operator").
		Order("tanggal_laporan desc")

	archiveDate := time.Now().Add(-time.Duration(archiveDurationDays) * 24 * time.Hour)
	if statusFilter == "archived" {
		db = db.Where("tanggal_laporan <= ?", archiveDate)
	} else {
		db = db.Where("tanggal_laporan > ?", archiveDate)
	}

	if query != "" {
		searchQuery := fmt.Sprintf("%%%s%%", query)
		db = db.Joins("JOIN residents ON lost_documents.resident_id = residents.id").
			Where("lost_documents.nomor_surat LIKE ? OR residents.nama_lengkap LIKE ?", searchQuery, searchQuery)
	}

	err := db.Find(&docs).Error
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *lostDocumentRepository) SearchGlobal(query string) ([]models.LostDocument, error) {
	var docs []models.LostDocument
	db := r.db.
		Preload("Resident").
		Preload("LostItems").
		Preload("PetugasPelapor").
		Preload("PejabatPersetuju").
		Preload("Operator").
		Order("tanggal_laporan desc")

	if query != "" {
		searchQuery := fmt.Sprintf("%%%s%%", query)
		db = db.Joins("JOIN residents ON lost_documents.resident_id = residents.id").
			Where("lost_documents.nomor_surat LIKE ? OR residents.nama_lengkap LIKE ?", searchQuery, searchQuery)
	} else {
		return docs, nil
	}

	err := db.Find(&docs).Error
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *lostDocumentRepository) Create(tx *gorm.DB, doc *models.LostDocument) (*models.LostDocument, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	if err := db.Create(doc).Error; err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *lostDocumentRepository) FindByID(id uint) (*models.LostDocument, error) {
	var doc models.LostDocument
	err := r.db.Preload("Resident").Preload("LostItems").Preload("PetugasPelapor").Preload("PejabatPersetuju").Preload("Operator").Preload("LastUpdatedBy").First(&doc, id).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *lostDocumentRepository) Update(tx *gorm.DB, doc *models.LostDocument) (*models.LostDocument, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(doc).Error; err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *lostDocumentRepository) Delete(tx *gorm.DB, id uint) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Delete(&models.LostDocument{}, id).Error
}

func (r *lostDocumentRepository) GetLastDocumentOfYear(year int) (*models.LostDocument, error) {
	var doc models.LostDocument
	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := startOfYear.AddDate(1, 0, 0).Add(-time.Nanosecond)
	err := r.db.Unscoped().Where("created_at BETWEEN ? AND ?", startOfYear, endOfYear).Where("nomor_surat NOT LIKE ?", "DELETED_%").Order("id desc").First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *lostDocumentRepository) GetMonthlyIssuanceForYear(year int) ([]MonthlyCount, error) {
	var results []MonthlyCount
	err := r.db.Model(&models.LostDocument{}).Select("CAST(strftime('%Y', tanggal_laporan) AS INTEGER) as year, CAST(strftime('%m', tanggal_laporan) AS INTEGER) as month, COUNT(id) as count").Where("CAST(strftime('%Y', tanggal_laporan) AS INTEGER) = ?", year).Group("year, month").Order("month asc").Scan(&results).Error
	return results, err
}

func (r *lostDocumentRepository) GetItemCompositionStats() ([]ItemCompositionStat, error) {
	var results []ItemCompositionStat
	err := r.db.Model(&models.LostItem{}).Select("nama_barang, COUNT(id) as count").Group("nama_barang").Order("count desc").Scan(&results).Error
	return results, err
}