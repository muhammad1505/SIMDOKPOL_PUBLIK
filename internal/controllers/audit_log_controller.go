package controllers

import (
	"net/http"
	_ "simdokpol/internal/models" // <-- PERBAIKAN: Ditambahkan blank identifier '_'
	"simdokpol/internal/services"

	"github.com/gin-gonic/gin"
)

type AuditLogController struct {
	service services.AuditLogService
}

func NewAuditLogController(service services.AuditLogService) *AuditLogController {
	return &AuditLogController{service: service}
}

// @Summary Mendapatkan Semua Log Audit
// @Description Mengambil seluruh riwayat aktivitas yang tercatat di sistem. Hanya bisa diakses oleh Super Admin.
// @Tags Audit Log
// @Produce json
// @Success 200 {array} models.AuditLog
// @Failure 500 {object} map[string]string "Error: Gagal mengambil data log audit"
// @Security BearerAuth
// @Router /audit-logs [get]
func (c *AuditLogController) FindAll(ctx *gin.Context) {
	logs, err := c.service.FindAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data log audit"})
		return
	}
	ctx.JSON(http.StatusOK, logs)
}