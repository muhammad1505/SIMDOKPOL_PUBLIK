package controllers

import (
	//"log"
	"net/http"
	"simdokpol/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service services.AuthService
}

func NewAuthController(service services.AuthService) *AuthController {
	return &AuthController{service: service}
}

type LoginRequest struct {
	NRP      string `json:"nrp" binding:"required" example:"12345"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// --- TAMBAHKAN BLOK KOMENTAR SWAGGER DI SINI ---
// @Summary Login Pengguna
// @Description Melakukan otentikasi pengguna berdasarkan NRP dan kata sandi, lalu mengembalikan token JWT dalam HttpOnly cookie.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Data Login Pengguna"
// @Success 200 {object} map[string]string "Contoh: {\"message\": \"Login berhasil\"}"
// @Failure 400 {object} map[string]string "Contoh: {\"error\": \"NRP dan Kata Sandi diperlukan\"}"
// @Failure 401 {object} map[string]string "Contoh: {\"error\": \"NRP atau kata sandi salah\"}"
// @Router /login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "NRP dan Kata Sandi diperlukan"})
		return
	}

	token, err := c.service.Login(req.NRP, req.Password)
	if err != nil {
		// Tidak perlu log error di sini karena service sudah memberikan pesan yang aman untuk pengguna
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Set token di dalam httpOnly cookie untuk keamanan
	// Durasi cookie 24 jam (dalam detik)
	ctx.SetCookie("token", token, 3600*24, "/", "localhost", false, true)

	ctx.JSON(http.StatusOK, gin.H{"message": "Login berhasil"})
}

// (Fungsi Logout tidak perlu didokumentasikan di Swagger untuk saat ini)
func (c *AuthController) Logout(ctx *gin.Context) {
	// Cara menghapus cookie adalah dengan mengaturnya kembali
	// dengan waktu kedaluwarsa di masa lalu (MaxAge < 0)
	ctx.SetCookie("token", "", -1, "/", "localhost", false, true)
	ctx.JSON(http.StatusOK, gin.H{"message": "Logout berhasil"})
}