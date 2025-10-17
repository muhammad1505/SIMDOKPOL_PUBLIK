/**
 * FILE HEADER: internal/controllers/user_controller_test.go
 *
 * PURPOSE:
 * File ini berisi unit test untuk UserController. Tujuannya adalah untuk
 * memverifikasi bahwa endpoint-endpoint manajemen pengguna berfungsi dengan benar,
 * termasuk validasi input, otorisasi (hak akses), dan penanganan error.
 *
 * PENDEKDATAN:
 * - Mocking: UserService di-mock untuk mengisolasi controller dari business logic.
 * - Middleware Simulation: Otorisasi (AuthMiddleware) disimulasikan
 * dengan menyuntikkan data 'currentUser' ke dalam konteks Gin untuk setiap request tes.
 * - Table-Driven Tests: Menggunakan struct test case untuk menguji berbagai skenario
 * dengan rapi dan efisien.
 */
package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"simdokpol/internal/middleware"
	"simdokpol/internal/models"
	"simdokpol/internal/services/mocks"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupTestRouter membuat instance Gin baru dan menerapkan middleware yang relevan untuk pengujian.
func setupTestRouter(mockUserService *mocks.UserService) (*gin.Engine, *mocks.UserService) {
	gin.SetMode(gin.TestMode)

	if mockUserService == nil {
		mockUserService = new(mocks.UserService)
	}

	userController := NewUserController(mockUserService)

	router := gin.New()
	// Middleware ini menyuntikkan user ke konteks, mensimulasikan AuthMiddleware
	router.Use(func(c *gin.Context) {
		// User akan di-set di setiap test case individu
		c.Next()
	})

	// Rute publik/non-admin
	router.PUT("/api/profile", userController.UpdateProfile)

	// Grup rute yang dilindungi oleh AdminAuthMiddleware
	adminRoutes := router.Group("/api")
	adminRoutes.Use(middleware.AdminAuthMiddleware())
	{
		adminRoutes.POST("/users", userController.Create)
	}

	return router, mockUserService
}

// TestUserController_Create menguji endpoint POST /api/users
func TestUserController_Create(t *testing.T) {
	// --- SETUP DATA UMUM ---
	adminUser := &models.User{ID: 1, NamaLengkap: "Admin", Peran: models.RoleSuperAdmin}
	operatorUser := &models.User{ID: 2, NamaLengkap: "Operator", Peran: models.RoleOperator}
	
	validRequestBody := CreateUserRequest{
		NamaLengkap: "USER BARU",
		NRP:         "99999",
		KataSandi:   "password123",
		Pangkat:     "BRIPDA",
		Peran:       models.RoleOperator,
		Jabatan:     "ANGGOTA JAGA REGU",
		Regu:        "I",
	}

	// --- DEFINISI TEST CASE ---
	testCases := []struct {
		name               string
		userInContext      *models.User
		requestBody        interface{}
		mockSetup          func(*mocks.UserService)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:          "Success - Super Admin Creates User",
			userInContext: adminUser,
			requestBody:   validRequestBody,
			mockSetup: func(mockSvc *mocks.UserService) {
				mockSvc.On("Create", mock.AnythingOfType("*models.User"), adminUser.ID).Return(nil).Once()
			},
			expectedStatusCode: http.StatusCreated,
			expectedBody: `{"nama_lengkap":"USER BARU","nrp":"99999","pangkat":"BRIPDA","peran":"OPERATOR","jabatan":"ANGGOTA JAGA REGU","regu":"I"}`,
		},
		{
			name:          "Failure - Operator Tries to Create User (Authorization Error)",
			userInContext: operatorUser,
			requestBody:   validRequestBody,
			mockSetup:          func(mockSvc *mocks.UserService) {},
			expectedStatusCode: http.StatusForbidden,
			expectedBody:       `{"error":"Akses ditolak. Anda tidak memiliki hak akses yang cukup."}`,
		},
		{
			name:          "Failure - Invalid Request Body (Missing NamaLengkap)",
			userInContext: adminUser,
			requestBody: gin.H{
				"nrp":        "99999",
				"kata_sandi": "password123",
				"pangkat":    "BRIPDA",
				"peran":      "OPERATOR",
				"jabatan":    "ANGGOTA JAGA REGU",
			},
			mockSetup:          func(mockSvc *mocks.UserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody: `{"error":"Key: 'CreateUserRequest.NamaLengkap' Error:Field validation for 'NamaLengkap' failed on the 'required' tag"}`,
		},
		{
			name:          "Failure - Service Layer Returns Error",
			userInContext: adminUser,
			requestBody:   validRequestBody,
			mockSetup: func(mockSvc *mocks.UserService) {
				mockSvc.On("Create", mock.AnythingOfType("*models.User"), adminUser.ID).Return(errors.New("NRP sudah terdaftar")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"error":"Gagal membuat pengguna."}`,
		},
	}

	// --- EKSEKUSI TEST ---
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockUserService := new(mocks.UserService)
			router, _ := setupTestRouter(mockUserService)
			tc.mockSetup(mockUserService)

			jsonBody, err := json.Marshal(tc.requestBody)
			assert.NoError(t, err)

			req, _ := http.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			router.Use(func(c *gin.Context) {
				if tc.userInContext != nil {
					c.Set("currentUser", tc.userInContext)
					c.Set("userID", tc.userInContext.ID)
				}
			})
			
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatusCode, recorder.Code)
			
			if tc.expectedStatusCode == http.StatusCreated {
				var responseUser models.User
				err := json.Unmarshal(recorder.Body.Bytes(), &responseUser)
				assert.NoError(t, err)
				assert.Equal(t, "USER BARU", responseUser.NamaLengkap)
				assert.Equal(t, "99999", responseUser.NRP)
			} else {
				assert.JSONEq(t, tc.expectedBody, recorder.Body.String())
			}

			mockUserService.AssertExpectations(t)
		})
	}
}


// TestUserController_UpdateProfile menguji endpoint PUT /api/profile
func TestUserController_UpdateProfile(t *testing.T) {
	// --- SETUP DATA UMUM ---
	loggedInUser := &models.User{ID: 5, NamaLengkap: "User Lama", NRP: "55555", Pangkat: "BRIPTU"}
	
	validRequestBody := UpdateProfileRequest{
		NamaLengkap: "USER BARU",
		NRP:         "55555-NEW",
		Pangkat:     "BRIPKA",
	}

	// --- DEFINISI TEST CASE ---
	testCases := []struct {
		name               string
		userInContext      *models.User
		requestBody        interface{}
		mockSetup          func(*mocks.UserService)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:          "Success - User Updates Own Profile",
			userInContext: loggedInUser,
			requestBody:   validRequestBody,
			mockSetup: func(mockSvc *mocks.UserService) {
				// Simulasikan service mengembalikan data user yang telah diperbarui
				updatedUser := &models.User{
					ID:          loggedInUser.ID,
					NamaLengkap: validRequestBody.NamaLengkap,
					NRP:         validRequestBody.NRP,
					Pangkat:     validRequestBody.Pangkat,
				}
				mockSvc.On("UpdateProfile", loggedInUser.ID, mock.AnythingOfType("*models.User")).Return(updatedUser, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"message":"Profil berhasil diperbarui.", "data": {"user": {"ID":5, "nama_lengkap":"USER BARU", "nrp":"55555-NEW", "pangkat":"BRIPKA", "peran":"", "jabatan":"", "regu":"", "created_at":"0001-01-01T00:00:00Z", "updated_at":"0001-01-01T00:00:00Z"}}}`,
		},
		{
			name:          "Failure - Invalid Request Body (Missing Pangkat)",
			userInContext: loggedInUser,
			requestBody: gin.H{
				"nama_lengkap": "USER BARU",
				"nrp":          "55555-NEW",
			},
			mockSetup:          func(mockSvc *mocks.UserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"error":"Input tidak valid: Key: 'UpdateProfileRequest.Pangkat' Error:Field validation for 'Pangkat' failed on the 'required' tag"}`,
		},
		{
			name:          "Failure - Service Layer Returns Error",
			userInContext: loggedInUser,
			requestBody:   validRequestBody,
			mockSetup: func(mockSvc *mocks.UserService) {
				mockSvc.On("UpdateProfile", loggedInUser.ID, mock.AnythingOfType("*models.User")).Return(nil, errors.New("database connection error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"error":"Gagal memperbarui profil."}`,
		},
	}

	// --- EKSEKUSI TEST ---
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockUserService := new(mocks.UserService)
			// Rute /api/profile tidak memerlukan middleware admin, jadi kita bisa pakai router biasa
			router, _ := setupTestRouter(mockUserService) 
			tc.mockSetup(mockUserService)

			jsonBody, err := json.Marshal(tc.requestBody)
			assert.NoError(t, err)

			req, _ := http.NewRequest(http.MethodPut, "/api/profile", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			router.Use(func(c *gin.Context) {
				if tc.userInContext != nil {
					c.Set("currentUser", tc.userInContext)
					c.Set("userID", tc.userInContext.ID)
				}
			})
			
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatusCode, recorder.Code)
			assert.JSONEq(t, tc.expectedBody, recorder.Body.String())

			mockUserService.AssertExpectations(t)
		})
	}
}