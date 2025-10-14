package services

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
	"strconv"
	"strings"
	"time"
)

type LostDocumentService interface {
	CreateLostDocument(residentData models.Resident, items []models.LostItem, operatorID uint, lokasiHilang string, petugasPelaporID uint, pejabatPersetujuID uint) (*models.LostDocument, error)
	UpdateLostDocument(docID uint, residentData models.Resident, items []models.LostItem, lokasiHilang string, petugasPelaporID uint, pejabatPersetujuID uint, loggedInUserID uint) (*models.LostDocument, error)
	FindAll(query string, statusFilter string) ([]models.LostDocument, error)
	SearchGlobal(query string) ([]models.LostDocument, error)
	FindByID(id uint, actorID uint) (*models.LostDocument, error)
	DeleteLostDocument(id uint, loggedInUserID uint) error
}

type lostDocumentService struct {
	db            *gorm.DB
	docRepo       repositories.LostDocumentRepository
	residentRepo  repositories.ResidentRepository
	userRepo      repositories.UserRepository
	auditService  AuditLogService
	configService ConfigService
}

func NewLostDocumentService(db *gorm.DB, docRepo repositories.LostDocumentRepository, residentRepo repositories.ResidentRepository, userRepo repositories.UserRepository, auditService AuditLogService, configService ConfigService) LostDocumentService {
	return &lostDocumentService{
		db:            db,
		docRepo:       docRepo,
		residentRepo:  residentRepo,
		userRepo:      userRepo,
		auditService:  auditService,
		configService: configService,
	}
}

func (s *lostDocumentService) FindByID(id uint, actorID uint) (*models.LostDocument, error) {
	doc, err := s.docRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	actor, err := s.userRepo.FindByID(actorID)
	if err != nil {
		return nil, errors.New("pengguna tidak valid")
	}

	if actor.Peran != models.RoleSuperAdmin && doc.OperatorID != actorID {
		return nil, errors.New("akses ditolak")
	}

	appConfig, _ := s.configService.GetConfig()
	archiveDuration := time.Duration(appConfig.ArchiveDurationDays) * 24 * time.Hour

	if doc.Status == "DITERBITKAN" && time.Now().After(doc.TanggalLaporan.Add(archiveDuration)) {
		doc.Status = "DIARSIPKAN"
	}

	return doc, nil
}

func (s *lostDocumentService) generateDocumentNumber() (string, error) {
	loc, err := s.configService.GetLocation()
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	year := now.Year()
	month := int(now.Month())
	monthRoman := intToRoman(month)
	lastNumFromDB := 0
	lastDoc, err := s.docRepo.GetLastDocumentOfYear(year)
	if err != nil && err != gorm.ErrRecordNotFound {
		return "", err
	}
	if err == nil && lastDoc != nil {
		parts := strings.Split(lastDoc.NomorSurat, "/")
		if len(parts) > 1 {
			if num, err := strconv.Atoi(parts[1]); err == nil {
				lastNumFromDB = num
			}
		}
	}
	lastNumFromConfig := 0
	appConfig, err := s.configService.GetConfig()
	if err != nil {
		log.Printf("PERINGATAN: Tidak dapat memuat konfigurasi untuk penomoran surat: %v", err)
	} else {
		if num, err := strconv.Atoi(appConfig.NomorSuratTerakhir); err == nil {
			lastNumFromConfig = num
		}
	}
	trueLastNumber := 0
	if lastNumFromDB > lastNumFromConfig {
		trueLastNumber = lastNumFromDB
	} else {
		trueLastNumber = lastNumFromConfig
	}
	runningNumber := trueLastNumber + 1
	return fmt.Sprintf(appConfig.FormatNomorSurat, runningNumber, monthRoman, year), nil
}

func (s *lostDocumentService) CreateLostDocument(residentData models.Resident, items []models.LostItem, operatorID uint, lokasiHilang string, petugasPelaporID uint, pejabatPersetujuID uint) (*models.LostDocument, error) {
	var createdDocID uint
	var finalDocNumber string
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var existingResident models.Resident
		err := tx.Where("nama_lengkap = ? AND tanggal_lahir = ?", residentData.NamaLengkap, residentData.TanggalLahir).First(&existingResident).Error
		if err == gorm.ErrRecordNotFound {
			// Komentar: Ini adalah solusi untuk menangani pemohon tanpa NIK.
			// Untuk produksi, alur ini mungkin perlu disempurnakan,
			// misalnya dengan menyimpan sebagai draf atau integrasi dengan data kependudukan.
			residentData.NIK = fmt.Sprintf("TEMP%d", time.Now().UnixNano())
			newResident, createErr := s.residentRepo.Create(tx, &residentData)
			if createErr != nil {
				return createErr
			}
			existingResident = *newResident
		} else if err != nil {
			return err
		}
		docNumber, err := s.generateDocumentNumber()
		if err != nil {
			return err
		}
		finalDocNumber = docNumber
		loc, err := s.configService.GetLocation()
		if err != nil {
			loc = time.UTC
		}
		now := time.Now().In(loc)
		newDoc := &models.LostDocument{
			NomorSurat:         docNumber,
			TanggalLaporan:     now,
			Status:             "DITERBITKAN",
			LokasiHilang:       lokasiHilang,
			ResidentID:         existingResident.ID,
			PetugasPelaporID:   petugasPelaporID,
			PejabatPersetujuID: &pejabatPersetujuID,
			OperatorID:         operatorID,
			TanggalPersetujuan: &now,
			LostItems:          items,
		}
		created, err := s.docRepo.Create(tx, newDoc)
		if err != nil {
			return err
		}
		createdDocID = created.ID
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.auditService.LogActivity(operatorID, models.AuditCreateDocument, fmt.Sprintf("Membuat surat keterangan hilang baru dengan nomor: %s", finalDocNumber))
	finalDoc, err := s.docRepo.FindByID(createdDocID)
	if err != nil {
		return nil, err
	}
	return finalDoc, nil
}

func (s *lostDocumentService) UpdateLostDocument(docID uint, residentData models.Resident, items []models.LostItem, lokasiHilang string, petugasPelaporID uint, pejabatPersetujuID uint, loggedInUserID uint) (*models.LostDocument, error) {
	var updatedDoc *models.LostDocument
	err := s.db.Transaction(func(tx *gorm.DB) error {
		existingDoc, err := s.docRepo.FindByID(docID)
		if err != nil {
			return err
		}
		loggedInUser, err := s.userRepo.FindByID(loggedInUserID)
		if err != nil {
			return errors.New("pengguna tidak valid")
		}
		if loggedInUser.Peran != models.RoleSuperAdmin && existingDoc.OperatorID != loggedInUserID {
			return errors.New("akses ditolak: Anda bukan pemilik dokumen ini")
		}
		existingDoc.Resident.NamaLengkap = residentData.NamaLengkap
		existingDoc.Resident.TempatLahir = residentData.TempatLahir
		existingDoc.Resident.TanggalLahir = residentData.TanggalLahir
		existingDoc.Resident.JenisKelamin = residentData.JenisKelamin
		existingDoc.Resident.Agama = residentData.Agama
		existingDoc.Resident.Pekerjaan = residentData.Pekerjaan
		existingDoc.Resident.Alamat = residentData.Alamat
		existingDoc.LokasiHilang = lokasiHilang
		existingDoc.PetugasPelaporID = petugasPelaporID
		existingDoc.PejabatPersetujuID = &pejabatPersetujuID
		existingDoc.LastUpdatedByID = &loggedInUserID
		if err := tx.Where("lost_document_id = ?", docID).Delete(&models.LostItem{}).Error; err != nil {
			return err
		}
		existingDoc.LostItems = items
		updatedDoc, err = s.docRepo.Update(tx, existingDoc)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.auditService.LogActivity(loggedInUserID, models.AuditUpdateDocument, fmt.Sprintf("Memperbarui dokumen dengan Nomor Surat: %s", updatedDoc.NomorSurat))
	return updatedDoc, nil
}

func (s *lostDocumentService) DeleteLostDocument(id uint, loggedInUserID uint) error {
	var docToDelete models.LostDocument
	if err := s.db.First(&docToDelete, id).Error; err != nil {
		return errors.New("dokumen tidak ditemukan")
	}
	originalNomorSurat := docToDelete.NomorSurat
	err := s.db.Transaction(func(tx *gorm.DB) error {
		loggedInUser, err := s.userRepo.FindByID(loggedInUserID)
		if err != nil {
			return errors.New("pengguna tidak valid")
		}
		if loggedInUser.Peran != models.RoleSuperAdmin && docToDelete.OperatorID != loggedInUserID {
			return errors.New("akses ditolak: Anda bukan pemilik dokumen ini")
		}
		modifiedNomorSurat := fmt.Sprintf("DELETED_%d_%s", time.Now().Unix(), docToDelete.NomorSurat)
		if err := tx.Model(&models.LostDocument{}).Where("id = ?", id).Update("nomor_surat", modifiedNomorSurat).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.LostDocument{}, id).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.auditService.LogActivity(loggedInUserID, models.AuditDeleteDocument, fmt.Sprintf("Menghapus dokumen dengan Nomor Surat: %s", originalNomorSurat))
	return nil
}

func (s *lostDocumentService) processDocsStatus(docs []models.LostDocument) ([]models.LostDocument, error) {
	appConfig, err := s.configService.GetConfig()
	if err != nil {
		return nil, err
	}
	archiveDuration := time.Duration(appConfig.ArchiveDurationDays) * 24 * time.Hour

	for i := range docs {
		if docs[i].Status == "DITERBITKAN" && time.Now().After(docs[i].TanggalLaporan.Add(archiveDuration)) {
			docs[i].Status = "DIARSIPKAN"
		}
	}
	return docs, nil
}

func (s *lostDocumentService) SearchGlobal(query string) ([]models.LostDocument, error) {
	docs, err := s.docRepo.SearchGlobal(query)
	if err != nil {
		return nil, err
	}
	return s.processDocsStatus(docs)
}

func (s *lostDocumentService) FindAll(query string, statusFilter string) ([]models.LostDocument, error) {
	appConfig, err := s.configService.GetConfig()
	if err != nil {
		return nil, err
	}

	docs, err := s.docRepo.FindAll(query, statusFilter, appConfig.ArchiveDurationDays)
	if err != nil {
		return nil, err
	}
	return s.processDocsStatus(docs)
}

func intToRoman(num int) string {
	romanNumeralMap := map[int]string{1: "I", 2: "II", 3: "III", 4: "IV", 5: "V", 6: "VI", 7: "VII", 8: "VIII", 9: "IX", 10: "X", 11: "XI", 12: "XII"}
	if val, ok := romanNumeralMap[num]; ok {
		return val
	}
	return ""
}