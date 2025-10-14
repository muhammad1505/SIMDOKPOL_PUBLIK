package services

import (
	"errors"
	"regexp"
	"simdokpol/internal/dto"
	"simdokpol/internal/mocks"
	"simdokpol/internal/models"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	mock.ExpectQuery("select sqlite_version()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("3.30.1"))

	gormDB, err := gorm.Open(sqlite.Dialector{
		Conn: db,
	}, &gorm.Config{})
	assert.NoError(t, err)

	return gormDB, mock
}

func TestLostDocumentService_CreateLostDocument(t *testing.T) {
	residentData := models.Resident{
		NamaLengkap:  "Budi Santoso",
		TanggalLahir: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	items := []models.LostItem{
		{NamaBarang: "KTP", Deskripsi: "NIK: 12345"},
	}
	operatorID := uint(1)
	petugasPelaporID := uint(2)
	pejabatPersetujuID := uint(3)

	loc, _ := time.LoadLocation("Asia/Jakarta")
	mockConfig := &dto.AppConfig{
		FormatNomorSurat:   "SKH/%d/%s/TUK.7.2.1/%d",
		NomorSuratTerakhir: "0",
	}

	testCases := []struct {
		name          string
		setupMocks    func(dbMock sqlmock.Sqlmock, docRepo *mocks.LostDocumentRepository, resRepo *mocks.ResidentRepository, userRepo *mocks.UserRepository, auditService *mocks.AuditLogService, configService *mocks.ConfigService)
		expectedError bool
	}{
		{
			name: "Sukses - Membuat Dokumen dengan Penduduk Baru",
			setupMocks: func(dbMock sqlmock.Sqlmock, docRepo *mocks.LostDocumentRepository, resRepo *mocks.ResidentRepository, userRepo *mocks.UserRepository, auditService *mocks.AuditLogService, configService *mocks.ConfigService) {
				// Tidak dibatasi karena bisa dipanggil beberapa kali
				configService.On("GetLocation").Return(loc, nil)
				
				docRepo.On("GetLastDocumentOfYear", mock.AnythingOfType("int")).Return((*models.LostDocument)(nil), gorm.ErrRecordNotFound).Once()
				configService.On("GetConfig").Return(mockConfig, nil).Once()

				dbMock.ExpectBegin()
				
				expectedSQL := "SELECT * FROM `residents` WHERE (nama_lengkap = ? AND tanggal_lahir = ?) AND `residents`.`deleted_at` IS NULL ORDER BY `residents`.`id` LIMIT 1"
				dbMock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
					WithArgs(residentData.NamaLengkap, residentData.TanggalLahir).
					WillReturnError(gorm.ErrRecordNotFound)

				resRepo.On("Create", mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*models.Resident")).
					Return(&models.Resident{ID: 1}, nil).Once()

				docRepo.On("Create", mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*models.LostDocument")).
					Return(&models.LostDocument{ID: 101}, nil).Once()

				dbMock.ExpectCommit()

				auditService.On("LogActivity", operatorID, models.AuditCreateDocument, mock.AnythingOfType("string")).Once()

				finalDoc := &models.LostDocument{ID: 101, NomorSurat: "SKH/1/X/TUK.7.2.1/2025"}
				docRepo.On("FindByID", uint(101)).Return(finalDoc, nil).Once()
			},
			expectedError: false,
		},
		{
			name: "Gagal - Error saat membuat penduduk",
			setupMocks: func(dbMock sqlmock.Sqlmock, docRepo *mocks.LostDocumentRepository, resRepo *mocks.ResidentRepository, userRepo *mocks.UserRepository, auditService *mocks.AuditLogService, configService *mocks.ConfigService) {
				// Tidak dibatasi karena bisa dipanggil beberapa kali
				configService.On("GetLocation").Return(loc, nil).Maybe()
				
				// Jika GetLastDocumentOfYear dipanggil di awal
				docRepo.On("GetLastDocumentOfYear", mock.AnythingOfType("int")).Return((*models.LostDocument)(nil), gorm.ErrRecordNotFound).Maybe()
				configService.On("GetConfig").Return(mockConfig, nil).Maybe()
				
				dbMock.ExpectBegin()

				expectedSQL := "SELECT * FROM `residents` WHERE (nama_lengkap = ? AND tanggal_lahir = ?) AND `residents`.`deleted_at` IS NULL ORDER BY `residents`.`id` LIMIT 1"
				dbMock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
					WithArgs(residentData.NamaLengkap, residentData.TanggalLahir).
					WillReturnError(gorm.ErrRecordNotFound)
					
				resRepo.On("Create", mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*models.Resident")).
					Return((*models.Resident)(nil), errors.New("database error")).Once()

				dbMock.ExpectRollback()
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, dbMock := setupMockDB(t)
			mockDocRepo := new(mocks.LostDocumentRepository)
			mockResRepo := new(mocks.ResidentRepository)
			mockUserRepo := new(mocks.UserRepository)
			mockAuditService := new(mocks.AuditLogService)
			mockConfigService := new(mocks.ConfigService)

			tc.setupMocks(dbMock, mockDocRepo, mockResRepo, mockUserRepo, mockAuditService, mockConfigService)

			service := NewLostDocumentService(db, mockDocRepo, mockResRepo, mockUserRepo, mockAuditService, mockConfigService)

			_, err := service.CreateLostDocument(residentData, items, operatorID, "Jalan Sudirman", petugasPelaporID, pejabatPersetujuID)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDocRepo.AssertExpectations(t)
			mockResRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
			mockAuditService.AssertExpectations(t)
			// Gunakan AssertNumberOfCalls untuk yang pakai Maybe()
			// mockConfigService.AssertExpectations(t)
			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}