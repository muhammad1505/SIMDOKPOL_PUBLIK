package mocks

import (
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
	"time"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type LostDocumentRepository struct {
	mock.Mock
}

func (_m *LostDocumentRepository) Create(tx *gorm.DB, doc *models.LostDocument) (*models.LostDocument, error) {
	ret := _m.Called(tx, doc)
	return ret.Get(0).(*models.LostDocument), ret.Error(1)
}

func (_m *LostDocumentRepository) FindByID(id uint) (*models.LostDocument, error) {
	ret := _m.Called(id)
	return ret.Get(0).(*models.LostDocument), ret.Error(1)
}

func (_m *LostDocumentRepository) FindAll(query string, statusFilter string, archiveDurationDays int) ([]models.LostDocument, error) {
	ret := _m.Called(query, statusFilter, archiveDurationDays)
	return ret.Get(0).([]models.LostDocument), ret.Error(1)
}

func (_m *LostDocumentRepository) SearchGlobal(query string) ([]models.LostDocument, error) {
	ret := _m.Called(query)
	return ret.Get(0).([]models.LostDocument), ret.Error(1)
}

func (_m *LostDocumentRepository) Update(tx *gorm.DB, doc *models.LostDocument) (*models.LostDocument, error) {
	ret := _m.Called(tx, doc)
	return ret.Get(0).(*models.LostDocument), ret.Error(1)
}

func (_m *LostDocumentRepository) Delete(tx *gorm.DB, id uint) error {
	return _m.Called(tx, id).Error(0)
}

func (_m *LostDocumentRepository) GetLastDocumentOfYear(year int) (*models.LostDocument, error) {
	ret := _m.Called(year)
	return ret.Get(0).(*models.LostDocument), ret.Error(1)
}

func (_m *LostDocumentRepository) CountByDateRange(start time.Time, end time.Time) (int64, error) {
	ret := _m.Called(start, end)
	return ret.Get(0).(int64), ret.Error(1)
}

func (_m *LostDocumentRepository) GetMonthlyIssuanceForYear(year int) ([]repositories.MonthlyCount, error) {
	ret := _m.Called(year)
	return ret.Get(0).([]repositories.MonthlyCount), ret.Error(1)
}

func (_m *LostDocumentRepository) GetItemCompositionStats() ([]repositories.ItemCompositionStat, error) {
	ret := _m.Called()
	return ret.Get(0).([]repositories.ItemCompositionStat), ret.Error(1)
}

func (_m *LostDocumentRepository) FindExpiringDocumentsForUser(userID uint, expiryDateStart time.Time, expiryDateEnd time.Time) ([]models.LostDocument, error) {
	ret := _m.Called(userID, expiryDateStart, expiryDateEnd)
	return ret.Get(0).([]models.LostDocument), ret.Error(1)
}