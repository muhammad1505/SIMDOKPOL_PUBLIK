package controllers

import (
	"errors"
	"log"
	"net/http"
	"simdokpol/internal/models"
	"simdokpol/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService services.UserService
}

func NewUserController(userService services.UserService) *UserController {
	return &UserController{userService: userService}
}

type CreateUserRequest struct {
	NamaLengkap string `json:"nama_lengkap" binding:"required" example:"NAMA LENGKAP PETUGAS"`
	NRP         string `json:"nrp" binding:"required" example:"98765"`
	KataSandi   string `json:"kata_sandi" binding:"required,min=8" example:"password123"`
	Pangkat     string `json:"pangkat" binding:"required" example:"BRIPDA"`
	Peran       string `json:"peran" binding:"required" enums:"OPERATOR,SUPER_ADMIN"`
	Jabatan     string `json:"jabatan" binding:"required" example:"ANGGOTA JAGA REGU"`
	Regu        string `json:"regu" example:"I"`
}

type UpdateUserRequest struct {
	NamaLengkap string `json:"nama_lengkap" binding:"required"`
	NRP         string `json:"nrp" binding:"required"`
	KataSandi   string `json:"kata_sandi"`
	Pangkat     string `json:"pangkat" binding:"required"`
	Peran       string `json:"peran" binding:"required"`
	Jabatan     string `json:"jabatan" binding:"required"`
	Regu        string `json:"regu"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required" example:"password_lama123"`
	NewPassword     string `json:"new_password" binding:"required,min=8" example:"password_baru123"`
	ConfirmPassword string `json:"confirm_password" binding:"required" example:"password_baru123"`
}

type UpdateProfileRequest struct {
	NamaLengkap string `json:"nama_lengkap" binding:"required" example:"NAMA SAYA"`
	NRP         string `json:"nrp" binding:"required" example:"12345"`
	Pangkat     string `json:"pangkat" binding:"required" example:"BRIPKA"`
}

// @Summary Memperbarui Profil Pengguna
// @Description Memperbarui data profil (Nama, NRP, Pangkat) untuk pengguna yang sedang login.
// @Tags Profile
// @Accept json
// @Produce json
// @Param profile body UpdateProfileRequest true "Data Profil Baru"
// @Success 200 {object} map[string]interface{} "Pesan sukses dan data user yang diperbarui"
// @Failure 400 {object} map[string]string "Error: Input tidak valid"
// @Failure 500 {object} map[string]string "Error: Terjadi kesalahan pada server"
// @Security BearerAuth
// @Router /profile [put]
func (c *UserController) UpdateProfile(ctx *gin.Context) {
	var req UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		APIError(ctx, http.StatusBadRequest, "Input tidak valid: "+err.Error())
		return
	}

	userID := ctx.GetUint("userID")

	dataToUpdate := &models.User{
		NamaLengkap: req.NamaLengkap,
		NRP:         req.NRP,
		Pangkat:     req.Pangkat,
	}

	updatedUser, err := c.userService.UpdateProfile(userID, dataToUpdate)
	if err != nil {
		log.Printf("ERROR: Gagal memperbarui profil untuk user ID %d: %v", userID, err)
		APIError(ctx, http.StatusInternalServerError, "Gagal memperbarui profil.")
		return
	}

	APIResponse(ctx, http.StatusOK, "Profil berhasil diperbarui.", gin.H{"user": updatedUser})
}

// @Summary Mengubah Kata Sandi Pengguna
// @Description Mengubah kata sandi untuk pengguna yang sedang login.
// @Tags Profile
// @Accept json
// @Produce json
// @Param passwords body ChangePasswordRequest true "Data Kata Sandi Lama dan Baru"
// @Success 200 {object} map[string]string "Pesan Sukses"
// @Failure 400 {object} map[string]string "Error: Input tidak valid"
// @Failure 409 {object} map[string]string "Error: Kata sandi lama salah"
// @Security BearerAuth
// @Router /profile/password [put]
func (c *UserController) ChangePassword(ctx *gin.Context) {
	var req ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		APIError(ctx, http.StatusBadRequest, "Semua kolom wajib diisi dan kata sandi baru minimal 8 karakter.")
		return
	}
	if req.NewPassword != req.ConfirmPassword {
		APIError(ctx, http.StatusBadRequest, "Konfirmasi kata sandi baru tidak cocok.")
		return
	}

	userID := ctx.GetUint("userID")

	err := c.userService.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		log.Printf("Gagal mengubah password untuk user ID %d: %v", userID, err)
		if errors.Is(err, services.ErrOldPasswordMismatch) {
			APIError(ctx, http.StatusConflict, err.Error())
		} else {
			APIError(ctx, http.StatusInternalServerError, "Gagal mengubah kata sandi.")
		}
		return
	}

	APIResponse(ctx, http.StatusOK, "Kata sandi berhasil diperbarui.", nil)
}

// @Summary Membuat Pengguna Baru
// @Description Membuat akun pengguna baru (Operator atau Super Admin). Hanya bisa diakses oleh Super Admin.
// @Tags Users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "Data Pengguna Baru"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]string "Error: Input tidak valid"
// @Failure 500 {object} map[string]string "Error: Terjadi kesalahan pada server"
// @Security BearerAuth
// @Router /users [post]
func (c *UserController) Create(ctx *gin.Context) {
	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		APIError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	actorID := ctx.GetUint("userID")

	user := models.User{
		NamaLengkap: req.NamaLengkap,
		NRP:         req.NRP,
		KataSandi:   req.KataSandi,
		Pangkat:     req.Pangkat,
		Peran:       req.Peran,
		Jabatan:     req.Jabatan,
		Regu:        req.Regu,
	}

	if err := c.userService.Create(&user, actorID); err != nil {
		log.Printf("ERROR: Gagal membuat pengguna: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal membuat pengguna.")
		return
	}
	ctx.JSON(http.StatusCreated, user)
}

func (c *UserController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "ID Pengguna tidak valid")
		return
	}
	var req UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		APIError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if req.KataSandi != "" && len(req.KataSandi) < 8 {
		APIError(ctx, http.StatusBadRequest, "Kata sandi baru minimal 8 karakter")
		return
	}
	actorID := ctx.GetUint("userID")

	user := models.User{
		ID:          uint(id),
		NamaLengkap: req.NamaLengkap,
		NRP:         req.NRP,
		Pangkat:     req.Pangkat,
		Peran:       req.Peran,
		Jabatan:     req.Jabatan,
		Regu:        req.Regu,
	}

	if err := c.userService.Update(&user, req.KataSandi, actorID); err != nil {
		log.Printf("ERROR: Gagal memperbarui pengguna id %d: %v", id, err)
		APIError(ctx, http.StatusInternalServerError, "Gagal memperbarui pengguna.")
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (c *UserController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "ID Pengguna tidak valid")
		return
	}
	actorID := ctx.GetUint("userID")

	if err := c.userService.Deactivate(uint(id), actorID); err != nil {
		log.Printf("ERROR: Gagal menonaktifkan pengguna id %d: %v", id, err)
		APIError(ctx, http.StatusInternalServerError, "Gagal menonaktifkan pengguna.")
		return
	}
	APIResponse(ctx, http.StatusOK, "Pengguna berhasil dinonaktifkan", nil)
}

func (c *UserController) Activate(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "ID Pengguna tidak valid")
		return
	}
	actorID := ctx.GetUint("userID")

	if err := c.userService.Activate(uint(id), actorID); err != nil {
		log.Printf("ERROR: Gagal mengaktifkan pengguna id %d: %v", id, err)
		APIError(ctx, http.StatusInternalServerError, "Gagal mengaktifkan pengguna.")
		return
	}
	APIResponse(ctx, http.StatusOK, "Pengguna berhasil diaktifkan", nil)
}

// @Summary Mendapatkan Semua Pengguna
// @Description Mengambil daftar semua pengguna (aktif atau non-aktif). Hanya bisa diakses oleh Super Admin.
// @Tags Users
// @Produce json
// @Param status query string false "Filter status pengguna" enums(active, inactive) default(active)
// @Success 200 {array} models.User
// @Failure 500 {object} map[string]string "Error: Gagal mengambil data pengguna"
// @Security BearerAuth
// @Router /users [get]
func (c *UserController) FindAll(ctx *gin.Context) {
	statusFilter := ctx.DefaultQuery("status", "active")
	users, err := c.userService.FindAll(statusFilter)
	if err != nil {
		log.Printf("ERROR: Gagal mengambil data semua pengguna: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal mengambil data pengguna.")
		return
	}
	ctx.JSON(http.StatusOK, users)
}

func (c *UserController) FindByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "ID Pengguna tidak valid")
		return
	}
	user, err := c.userService.FindByID(uint(id))
	if err != nil {
		APIError(ctx, http.StatusNotFound, "Pengguna tidak ditemukan")
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (c *UserController) FindOperators(ctx *gin.Context) {
	operators, err := c.userService.FindOperators()
	if err != nil {
		log.Printf("ERROR: Gagal mengambil data operator: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal mengambil data operator.")
		return
	}
	ctx.JSON(http.StatusOK, operators)
}