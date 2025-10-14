package mocks

import (
	"simdokpol/internal/models"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type ResidentRepository struct {
	mock.Mock
}

func (_m *ResidentRepository) FindByNIK(tx *gorm.DB, nik string) (*models.Resident, error) {
	ret := _m.Called(tx, nik)
	return ret.Get(0).(*models.Resident), ret.Error(1)
}

func (_m *ResidentRepository) Create(tx *gorm.DB, resident *models.Resident) (*models.Resident, error) {
	ret := _m.Called(tx, resident)
	return ret.Get(0).(*models.Resident), ret.Error(1)
}