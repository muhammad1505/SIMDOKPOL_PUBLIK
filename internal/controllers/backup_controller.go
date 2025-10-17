package controllers

import (
	"log"
	"net/http"
	"path/filepath"
	"simdokpol/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
)

type BackupController struct {
	service services.BackupService
}

func NewBackupController(service services.BackupService) *BackupController {
	return &BackupController{service: service}
}

// @Summary Membuat Backup Database
// @Description Membuat salinan database saat ini dan mengirimkannya sebagai file unduhan. Hanya bisa diakses oleh Super Admin.
// @Tags Backup & Restore
// @Produce application/octet-stream
// @Success 200 {file} file "File backup database (.db)"
// @Failure 500 {object} map[string]string "Error: Gagal memproses backup"
// @Security BearerAuth
// @Router /backups [post]
func (c *BackupController) CreateBackup(ctx *gin.Context) {
	actorID := ctx.GetUint("userID")
	backupPath, err := c.service.CreateBackup(actorID)
	if err != nil {
		log.Printf("ERROR: Gagal membuat backup oleh user id %d: %v", actorID, err)
		APIError(ctx, http.StatusInternalServerError, "Gagal memproses backup.")
		return
	}

	// Mengirim file sebagai attachment untuk diunduh oleh browser.
	ctx.FileAttachment(backupPath, filepath.Base(backupPath))
}

// @Summary Melakukan Restore Database
// @Description Memulihkan database dari file .db yang diunggah. Semua data saat ini akan ditimpa. Hanya bisa diakses oleh Super Admin.
// @Tags Backup & Restore
// @Accept multipart/form-data
// @Produce json
// @Param restore-file formData file true "File backup .db yang akan di-restore"
// @Success 200 {object} map[string]string "Pesan Sukses"
// @Failure 400 {object} map[string]string "Error: Input tidak valid"
// @Failure 500 {object} map[string]string "Error: Gagal memulihkan database"
// @Security BearerAuth
// @Router /restore [post]
func (c *BackupController) RestoreBackup(ctx *gin.Context) {
	file, err := ctx.FormFile("restore-file")
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "Tidak ada file yang diunggah.")
		return
	}

	if !strings.HasSuffix(file.Filename, ".db") {
		APIError(ctx, http.StatusBadRequest, "Format file tidak valid. Harap unggah file .db")
		return
	}

	src, err := file.Open()
	if err != nil {
		log.Printf("ERROR: Gagal membuka file restore yang diunggah: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal memproses file yang diunggah.")
		return
	}
	defer src.Close()

	actorID := ctx.GetUint("userID")
	if err := c.service.RestoreBackup(src, actorID); err != nil {
		log.Printf("ERROR: Gagal melakukan restore oleh user id %d: %v", actorID, err)
		APIError(ctx, http.StatusInternalServerError, "Gagal memulihkan database.")
		return
	}

	APIResponse(ctx, http.StatusOK, "Database berhasil dipulihkan.", nil)
}