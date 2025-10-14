package services

import (
	"errors"
	"simdokpol/internal/mocks"
	"simdokpol/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestAuthService_Login(t *testing.T) {
	// Setup: Buat password hash sekali untuk semua test case
	password := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	// Setup: Buat mock user untuk digunakan dalam test
	mockUser := &models.User{
		ID:        1,
		NRP:       "12345",
		KataSandi: string(hashedPassword),
		Peran:     models.RoleOperator,
		DeletedAt: gorm.DeletedAt{}, // Akun aktif
	}

	mockInactiveUser := &models.User{
		ID:        2,
		NRP:       "54321",
		KataSandi: string(hashedPassword),
		Peran:     models.RoleOperator,
		DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true}, // Akun non-aktif
	}

	// Atur secret key untuk JWT (hanya untuk testing)
	JWTSecretKey = []byte("test-secret")

	// Definisikan semua test case
	testCases := []struct {
		name          string
		nrp           string
		password      string
		setupMock     func(mockRepo *mocks.UserRepository)
		expectToken   bool
		expectedError string
	}{
		{
			name:        "Login Berhasil",
			nrp:         "12345",
			password:    "password123",
			setupMock: func(mockRepo *mocks.UserRepository) {
				// Harapkan method FindByNRP dipanggil dengan NRP "12345"
				// dan kembalikan mockUser tanpa error
				mockRepo.On("FindByNRP", "12345").Return(mockUser, nil)
			},
			expectToken:   true,
			expectedError: "",
		},
		{
			name:        "Gagal - Kata Sandi Salah",
			nrp:         "12345",
			password:    "password-salah",
			setupMock: func(mockRepo *mocks.UserRepository) {
				mockRepo.On("FindByNRP", "12345").Return(mockUser, nil)
			},
			expectToken:   false,
			expectedError: "NRP atau kata sandi salah",
		},
		{
			name:        "Gagal - Pengguna Tidak Ditemukan",
			nrp:         "00000",
			password:    "password123",
			setupMock: func(mockRepo *mocks.UserRepository) {
				// Harapkan FindByNRP mengembalikan error gorm.ErrRecordNotFound
				mockRepo.On("FindByNRP", "00000").Return(nil, gorm.ErrRecordNotFound)
			},
			expectToken:   false,
			expectedError: "NRP atau kata sandi salah",
		},
		{
			name:        "Gagal - Akun Tidak Aktif",
			nrp:         "54321",
			password:    "password123",
			setupMock: func(mockRepo *mocks.UserRepository) {
				mockRepo.On("FindByNRP", "54321").Return(mockInactiveUser, nil)
			},
			expectToken:   false,
			expectedError: "Akun Anda tidak aktif. Silakan hubungi Super Admin",
		},
		{
			name:        "Gagal - Error Database Lainnya",
			nrp:         "12345",
			password:    "password123",
			setupMock: func(mockRepo *mocks.UserRepository) {
				// Simulasikan error internal server
				mockRepo.On("FindByNRP", "12345").Return(nil, errors.New("koneksi database error"))
			},
			expectToken:   false,
			expectedError: "koneksi database error",
		},
	}

	// Jalankan semua test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Buat instance mock repository baru untuk setiap test
			mockUserRepo := new(mocks.UserRepository)
			
			// 2. Setup mock sesuai definisi test case
			tc.setupMock(mockUserRepo)

			// 3. Buat instance AuthService dengan mock repository
			authService := NewAuthService(mockUserRepo)

			// 4. Panggil method Login yang ingin di-test
			token, err := authService.Login(tc.nrp, tc.password)

			// 5. Lakukan assertion (pemeriksaan hasil)
			if tc.expectToken {
				assert.NoError(t, err, "Seharusnya tidak ada error")
				assert.NotEmpty(t, token, "Token seharusnya tidak kosong")
			} else {
				assert.Error(t, err, "Seharusnya ada error")
				assert.Empty(t, token, "Token seharusnya kosong")
				assert.Equal(t, tc.expectedError, err.Error(), "Pesan error tidak sesuai")
			}
			
			// 6. Verifikasi bahwa method yang di-mock benar-benar dipanggil
			mockUserRepo.AssertExpectations(t)
		})
	}
}