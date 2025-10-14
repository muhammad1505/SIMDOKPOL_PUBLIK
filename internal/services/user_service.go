package services

import (
	"errors"
	"fmt"
	"simdokpol/internal/config"
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Create(user *models.User, actorID uint) error
	FindAll(statusFilter string) ([]models.User, error)
	FindByID(id uint) (*models.User, error)
	FindOperators() ([]models.User, error)
	Update(user *models.User, newPassword string, actorID uint) error
	Deactivate(id uint, actorID uint) error
	Activate(id uint, actorID uint) error
	ChangePassword(userID uint, oldPassword, newPassword string) error
	UpdateProfile(userID uint, dataToUpdate *models.User) (*models.User, error) // <-- METHOD BARU
}

type userService struct {
	userRepo     repositories.UserRepository
	auditService AuditLogService
	cfg          *config.Config
}

func NewUserService(userRepo repositories.UserRepository, auditService AuditLogService, cfg *config.Config) UserService {
	return &userService{
		userRepo:     userRepo,
		auditService: auditService,
		cfg:          cfg,
	}
}

// === FUNGSI BARU UNTUK UPDATE PROFIL ===
func (s *userService) UpdateProfile(userID uint, dataToUpdate *models.User) (*models.User, error) {
	currentUser, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}

	// Logika Keamanan: Hanya perbarui field yang diizinkan untuk diubah oleh pengguna.
	// Jabatan, Peran, dan Regu tidak disentuh.
	currentUser.NamaLengkap = dataToUpdate.NamaLengkap
	currentUser.NRP = dataToUpdate.NRP
	currentUser.Pangkat = dataToUpdate.Pangkat

	if err := s.userRepo.Update(currentUser); err != nil {
		return nil, err
	}

	logDetails := fmt.Sprintf("Pengguna '%s' (NRP: %s) memperbarui data profilnya.", currentUser.NamaLengkap, currentUser.NRP)
	s.auditService.LogActivity(userID, models.AuditUpdateUser, logDetails)

	return currentUser, nil
}
// === AKHIR FUNGSI BARU ===


func (s *userService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("pengguna tidak ditemukan")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.KataSandi), []byte(oldPassword))
	if err != nil {
		return errors.New("kata sandi saat ini yang Anda masukkan salah")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.cfg.BcryptCost)
	if err != nil {
		return err
	}
	user.KataSandi = string(hashedPassword)

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	logDetails := fmt.Sprintf("Pengguna '%s' (NRP: %s) mengubah kata sandinya sendiri.", user.NamaLengkap, user.NRP)
	s.auditService.LogActivity(userID, models.AuditUpdateUser, logDetails)

	return nil
}

// ... (sisa fungsi Create, Update (admin), Deactivate, dll. tidak berubah) ...
func (s *userService) Create(user *models.User, actorID uint) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.KataSandi), s.cfg.BcryptCost)
	if err != nil {
		return err
	}
	user.KataSandi = string(hashedPassword)
	if err := s.userRepo.Create(user); err != nil {
		return err
	}

	logDetails := fmt.Sprintf("Pengguna baru '%s' (NRP: %s) telah dibuat.", user.NamaLengkap, user.NRP)
	auditAction := models.AuditCreateUser

	if actorID == 0 {
		actorID = user.ID
		logDetails = fmt.Sprintf("Akun Super Admin pertama '%s' (NRP: %s) dibuat saat setup.", user.NamaLengkap, user.NRP)
		auditAction = models.AuditSystemSetup
	}
	s.auditService.LogActivity(actorID, auditAction, logDetails)

	return nil
}

func (s *userService) Update(user *models.User, newPassword string, actorID uint) error {
	oldUser, err := s.userRepo.FindByID(user.ID)
	if err != nil {
		return errors.New("pengguna tidak ditemukan untuk pembaruan")
	}

	if strings.TrimSpace(newPassword) != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.cfg.BcryptCost)
		if err != nil {
			return err
		}
		user.KataSandi = string(hashedPassword)
	} else {
		user.KataSandi = oldUser.KataSandi
	}

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	logDetails := fmt.Sprintf("Data pengguna '%s' (NRP: %s) telah diperbarui.", user.NamaLengkap, user.NRP)
	if newPassword != "" {
		logDetails += " Termasuk perubahan kata sandi."
	}
	s.auditService.LogActivity(actorID, models.AuditUpdateUser, logDetails)

	return nil
}

func (s *userService) Deactivate(id uint, actorID uint) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("pengguna tidak ditemukan")
	}

	if err := s.userRepo.Delete(id); err != nil {
		return err
	}

	logDetails := fmt.Sprintf("Pengguna '%s' (NRP: %s) telah dinonaktifkan.", user.NamaLengkap, user.NRP)
	s.auditService.LogActivity(actorID, models.AuditDeactivateUser, logDetails)

	return nil
}

func (s *userService) Activate(id uint, actorID uint) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("pengguna tidak ditemukan")
	}

	if err := s.userRepo.Restore(id); err != nil {
		return err
	}

	logDetails := fmt.Sprintf("Pengguna '%s' (NRP: %s) telah diaktifkan kembali.", user.NamaLengkap, user.NRP)
	s.auditService.LogActivity(actorID, models.AuditActivateUser, logDetails)

	return nil
}

func (s *userService) FindAll(statusFilter string) ([]models.User, error) {
	return s.userRepo.FindAll(statusFilter)
}

func (s *userService) FindByID(id uint) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *userService) FindOperators() ([]models.User, error) {
	return s.userRepo.FindOperators()
}