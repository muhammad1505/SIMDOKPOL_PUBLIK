package controllers

import (
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
		APIError(ctx, http.StatusBadRequest, "NRP dan Kata Sandi diperlukan")
		return
	}

	token, err := c.service.Login(req.NRP, req.Password)
	if err != nil {
		APIError(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	// Untuk production, 'secure' harus true dan 'domain' disesuaikan
	ctx.SetCookie("token", token, 3600*24, "/", "localhost", false, true)

	APIResponse(ctx, http.StatusOK, "Login berhasil", nil)
}

// Logout tidak memerlukan dokumentasi Swagger
func (c *AuthController) Logout(ctx *gin.Context) {
	ctx.SetCookie("token", "", -1, "/", "localhost", false, true)
	APIResponse(ctx, http.StatusOK, "Logout berhasil", nil)
}