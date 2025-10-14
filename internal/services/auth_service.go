package services

import (
	"errors"
	"os"
	"simdokpol/internal/repositories"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var JWTSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

type AuthService interface {
	Login(nrp string, password string) (string, error)
}

type authService struct {
	userRepo repositories.UserRepository
}

func NewAuthService(userRepo repositories.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Login(nrp string, password string) (string, error) {
	// 1. Cari pengguna berdasarkan NRP, termasuk yang non-aktif
	user, err := s.userRepo.FindByNRP(nrp)
	if err != nil {
		// Jika tidak ditemukan sama sekali, kembalikan error biasa
		if err == gorm.ErrRecordNotFound {
			return "", errors.New("NRP atau kata sandi salah")
		}
		return "", err
	}

	// 2. Periksa apakah akun tersebut non-aktif (soft deleted)
	if user.DeletedAt.Valid {
		return "", errors.New("Akun Anda tidak aktif. Silakan hubungi Super Admin")
	}

	// 3. Jika aktif, lanjutkan verifikasi kata sandi
	err = bcrypt.CompareHashAndPassword([]byte(user.KataSandi), []byte(password))
	if err != nil {
		return "", errors.New("NRP atau kata sandi salah")
	}

	// 4. Buat token jika semua verifikasi berhasil
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": user.ID,
		"role":   user.Peran,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString(JWTSecretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}