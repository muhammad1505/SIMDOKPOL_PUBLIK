package controllers

import (
	"log"
	"net/http"
	_ "simdokpol/internal/dto" // <-- PERBAIKAN: Ditambahkan blank identifier '_'
	"simdokpol/internal/models"
	"simdokpol/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
)

type SettingsController struct {
	configService services.ConfigService
	auditService  services.AuditLogService
}

func NewSettingsController(configService services.ConfigService, auditService services.AuditLogService) *SettingsController {
	return &SettingsController{
		configService: configService,
		auditService:  auditService,
	}
}

// @Summary Mendapatkan Semua Pengaturan Sistem
// @Description Mengambil semua data konfigurasi sistem yang sedang aktif. Hanya bisa diakses oleh Super Admin.
// @Tags Settings
// @Produce json
// @Success 200 {object} dto.AppConfig
// @Failure 500 {object} map[string]string "Error: Gagal mengambil data pengaturan"
// @Security BearerAuth
// @Router /settings [get]
func (c *SettingsController) GetSettings(ctx *gin.Context) {
	config, err := c.configService.GetConfig()
	if err != nil {
		log.Printf("ERROR: Gagal mengambil data pengaturan: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pengaturan."})
		return
	}
	ctx.JSON(http.StatusOK, config)
}

// @Summary Memperbarui Pengaturan Sistem
// @Description Menyimpan satu atau lebih data konfigurasi sistem. Hanya bisa diakses oleh Super Admin.
// @Tags Settings
// @Accept json
// @Produce json
// @Param settings body dto.AppConfig true "Data Pengaturan Baru"
// @Success 200 {object} map[string]string "Pesan Sukses"
// @Failure 400 {object} map[string]string "Error: Format data tidak valid"
// @Failure 500 {object} map[string]string "Error: Gagal menyimpan pengaturan"
// @Security BearerAuth
// @Router /settings [put]
func (c *SettingsController) UpdateSettings(ctx *gin.Context) {
	var settings map[string]string
	if err := ctx.ShouldBindJSON(&settings); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid"})
		return
	}

	if path, exists := settings["backup_path"]; exists {
		if strings.Contains(path, "..") {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Path tidak valid. Tidak boleh mengandung '..'"})
			return
		}
	}

	if err := c.configService.SaveConfig(settings); err != nil {
		log.Printf("ERROR: Gagal menyimpan pengaturan: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pengaturan."})
		return
	}

	actorID, _ := ctx.Get("userID")
	c.auditService.LogActivity(actorID.(uint), models.AuditSettingsUpdated, "Pengaturan sistem telah diperbarui.")

	ctx.JSON(http.StatusOK, gin.H{"message": "Pengaturan berhasil disimpan"})
}