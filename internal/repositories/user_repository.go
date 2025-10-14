package repositories

import (
	"simdokpol/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindAll(statusFilter string) ([]models.User, error)
	FindByID(id uint) (*models.User, error)
	FindByNRP(nrp string) (*models.User, error)
	FindOperators() ([]models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	Restore(id uint) error
	CountAll() (int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindAll(statusFilter string) ([]models.User, error) {
	var users []models.User
	db := r.db.Order("nama_lengkap asc")
	if statusFilter == "inactive" {
		db = db.Unscoped().Where("deleted_at IS NOT NULL")
	}
	err := db.Find(&users).Error
	return users, err
}

func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.Unscoped().First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByNRP(nrp string) (*models.User, error) {
	var user models.User
	// Gunakan Unscoped() agar bisa menemukan pengguna yang sudah di-soft delete
	if err := r.db.Unscoped().Where("nrp = ?", nrp).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindOperators() ([]models.User, error) {
	var users []models.User
	err := r.db.Where("peran = ?", "OPERATOR").Order("nama_lengkap asc").Find(&users).Error
	return users, err
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *userRepository) Restore(id uint) error {
	return r.db.Unscoped().Model(&models.User{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *userRepository) CountAll() (int64, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}