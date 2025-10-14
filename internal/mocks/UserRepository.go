package mocks

import (
	"simdokpol/internal/models"
	"github.com/stretchr/testify/mock"
)

// UserRepository adalah mock untuk repositories.UserRepository
type UserRepository struct {
	mock.Mock
}

func (_m *UserRepository) Create(user *models.User) error {
	ret := _m.Called(user)
	return ret.Error(0)
}

func (_m *UserRepository) FindAll(statusFilter string) ([]models.User, error) {
	ret := _m.Called(statusFilter)
	var r0 []models.User
	if rf, ok := ret.Get(0).(func(string) []models.User); ok {
		r0 = rf(statusFilter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.User)
		}
	}
	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(statusFilter)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

func (_m *UserRepository) FindByID(id uint) (*models.User, error) {
	ret := _m.Called(id)
	var r0 *models.User
	if rf, ok := ret.Get(0).(func(uint) *models.User); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.User)
		}
	}
	var r1 error
	if rf, ok := ret.Get(1).(func(uint) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

func (_m *UserRepository) FindByNRP(nrp string) (*models.User, error) {
	ret := _m.Called(nrp)
	var r0 *models.User
	if rf, ok := ret.Get(0).(func(string) *models.User); ok {
		r0 = rf(nrp)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.User)
		}
	}
	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(nrp)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

func (_m *UserRepository) FindOperators() ([]models.User, error) {
	ret := _m.Called()
	var r0 []models.User
	if rf, ok := ret.Get(0).(func() []models.User); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.User)
		}
	}
	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

func (_m *UserRepository) Update(user *models.User) error {
	ret := _m.Called(user)
	return ret.Error(0)
}

func (_m *UserRepository) Delete(id uint) error {
	ret := _m.Called(id)
	return ret.Error(0)
}

func (_m *UserRepository) Restore(id uint) error {
	ret := _m.Called(id)
	return ret.Error(0)
}

func (_m *UserRepository) CountAll() (int64, error) {
	ret := _m.Called()
	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}
	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}