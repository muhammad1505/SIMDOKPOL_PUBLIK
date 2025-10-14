package controllers

import (
	"log"
	"net/http"
	"simdokpol/internal/models"
	"simdokpol/internal/services"

	"github.com/gin-gonic/gin"
)

type ConfigController struct {
	configService services.ConfigService
	userService   services.UserService
}

func NewConfigController(configService services.ConfigService, userService services.UserService) *ConfigController {
	return &ConfigController{
		configService: configService,
		userService:   userService,
	}
}

type SaveSetupRequest struct {
	KopBaris1          string `json:"kop_baris_1" binding:"required"`
	KopBaris2          string `json:"kop_baris_2" binding:"required"`
	KopBaris3          string `json:"kop_baris_3" binding:"required"`
	NamaKantor         string `json:"nama_kantor" binding:"required"`
	TempatSurat        string `json:"tempat_surat" binding:"required"`
	FormatNomorSurat   string `json:"format_nomor_surat" binding:"required"`
	NomorSuratTerakhir string `json:"nomor_surat_terakhir" binding:"required"`
	ZonaWaktu          string `json:"zona_waktu" binding:"required"`
	ArchiveDurationDays string `json:"archive_duration_days" binding:"required"` // Ditambahkan
	AdminNamaLengkap   string `json:"admin_nama_lengkap" binding:"required"`
	AdminNRP           string `json:"admin_nrp" binding:"required"`
	AdminPangkat       string `json:"admin_pangkat" binding:"required"`
	AdminPassword      string `json:"admin_password" binding:"required,min=8"`
}

func (c *ConfigController) SaveSetup(ctx *gin.Context) {
	isSetup, _ := c.configService.IsSetupComplete()
	if isSetup {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Aplikasi sudah dikonfigurasi."})
		return
	}

	var req SaveSetupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid: " + err.Error()})
		return
	}

	configData := map[string]string{
		"kop_baris_1":           req.KopBaris1,
		"kop_baris_2":           req.KopBaris2,
		"kop_baris_3":           req.KopBaris3,
		"nama_kantor":           req.NamaKantor,
		"tempat_surat":          req.TempatSurat,
		"format_nomor_surat":    req.FormatNomorSurat,
		"nomor_surat_terakhir":  req.NomorSuratTerakhir,
		"zona_waktu":            req.ZonaWaktu,
		"archive_duration_days": req.ArchiveDurationDays, // Ditambahkan
	}

	configData[services.IsSetupCompleteKey] = "true"

	if err := c.configService.SaveConfig(configData); err != nil {
		log.Printf("ERROR: Gagal menyimpan konfigurasi sistem saat setup: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan konfigurasi sistem."})
		return
	}

	superAdmin := &models.User{
		NamaLengkap: req.AdminNamaLengkap,
		NRP:         req.AdminNRP,
		Pangkat:     req.AdminPangkat,
		KataSandi:   req.AdminPassword,
		Peran:       models.RoleSuperAdmin,
		Jabatan:     models.RoleSuperAdmin,
	}

	if err := c.userService.Create(superAdmin, 0); err != nil {
		log.Printf("ERROR: Gagal membuat akun Super Admin saat setup: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat akun Super Admin."})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Konfigurasi berhasil disimpan. Silakan login menggunakan akun Super Admin yang baru dibuat."})
}

func (c *ConfigController) ShowSetupPage(ctx *gin.Context) {
	isSetup, _ := c.configService.IsSetupComplete()
	if isSetup {
		ctx.Redirect(http.StatusFound, "/login")
		return
	}
	ctx.HTML(http.StatusOK, "setup.html", gin.H{"Title": "Konfigurasi Awal"})
}