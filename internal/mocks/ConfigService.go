package mocks

import (
	"simdokpol/internal/dto" // <-- IMPORT DIUBAH
	"time"

	"github.com/stretchr/testify/mock"
)

type ConfigService struct {
	mock.Mock
}

func (_m *ConfigService) IsSetupComplete() (bool, error) {
	ret := _m.Called()
	return ret.Get(0).(bool), ret.Error(1)
}

func (_m *ConfigService) GetConfig() (*dto.AppConfig, error) { // <-- RETURN VALUE DIUBAH
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*dto.AppConfig), ret.Error(1)
}

func (_m *ConfigService) SaveConfig(configData map[string]string) error {
	return _m.Called(configData).Error(0)
}

func (_m *ConfigService) GetLocation() (*time.Location, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*time.Location), ret.Error(1)
}