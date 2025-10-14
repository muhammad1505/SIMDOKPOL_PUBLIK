package mocks

import (
	"simdokpol/internal/models" // <-- BARIS INI YANG DITAMBAHKAN
	"github.com/stretchr/testify/mock"
)

type AuditLogService struct {
	mock.Mock
}

func (_m *AuditLogService) LogActivity(userID uint, action string, details string) {
	_m.Called(userID, action, details)
}

func (_m *AuditLogService) FindAll() ([]models.AuditLog, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]models.AuditLog), ret.Error(1)
}