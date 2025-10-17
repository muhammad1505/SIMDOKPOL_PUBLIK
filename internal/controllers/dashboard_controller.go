package controllers

import (
	"log"
	"net/http"
	"simdokpol/internal/services"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	service services.DashboardService
}

func NewDashboardController(service services.DashboardService) *DashboardController {
	return &DashboardController{service: service}
}

// @Summary Mendapatkan Notifikasi Dokumen Kedaluwarsa
// @Description Mengambil daftar dokumen yang akan segera masuk masa arsip untuk pengguna yang sedang login.
// @Tags Dashboard & Stats
// @Produce json
// @Success 200 {array} models.LostDocument
// @Security BearerAuth
// @Router /notifications/expiring-documents [get]
func (c *DashboardController) GetExpiringDocuments(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	// Jendela notifikasi diatur ke 3 hari sebelum kedaluwarsa.
	// Angka ini bisa dibuat dinamis dari konfigurasi jika diperlukan.
	notificationWindowDays := 3
	documents, err := c.service.GetExpiringDocumentsForUser(userID, notificationWindowDays)
	if err != nil {
		log.Printf("ERROR: Gagal mengambil notifikasi dokumen kedaluwarsa untuk user ID %d: %v", userID, err)
		// Kembalikan array kosong agar tidak merusak UI frontend
		ctx.JSON(http.StatusOK, []string{})
		return
	}

	ctx.JSON(http.StatusOK, documents)
}

// @Summary Mendapatkan Statistik Utama Dasbor
// @Description Mengambil data statistik utama untuk kartu-kartu di dasbor (laporan hari ini, bulan ini, tahun ini, dan total pengguna).
// @Tags Dashboard & Stats
// @Produce json
// @Success 200 {object} services.DashboardStatsDTO
// @Failure 500 {object} map[string]string "Error: Gagal mengambil data statistik"
// @Security BearerAuth
// @Router /stats [get]
func (c *DashboardController) GetStats(ctx *gin.Context) {
	stats, err := c.service.GetDashboardStats()
	if err != nil {
		log.Printf("ERROR: Gagal mengambil statistik dasbor: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal mengambil data statistik")
		return
	}
	ctx.JSON(http.StatusOK, stats)
}

// @Summary Mendapatkan Data Grafik Bulanan
// @Description Mengambil data label dan jumlah penerbitan surat per bulan untuk grafik area di dasbor.
// @Tags Dashboard & Stats
// @Produce json
// @Success 200 {object} services.ChartDataDTO
// @Failure 500 {object} map[string]string "Error: Gagal mengambil data grafik"
// @Security BearerAuth
// @Router /stats/monthly-issuance [get]
func (c *DashboardController) GetMonthlyChart(ctx *gin.Context) {
	chartData, err := c.service.GetMonthlyIssuanceChartData()
	if err != nil {
		log.Printf("ERROR: Gagal mengambil data grafik bulanan: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal mengambil data grafik")
		return
	}
	ctx.JSON(http.StatusOK, chartData)
}

// @Summary Mendapatkan Data Komposisi Barang
// @Description Mengambil data label dan jumlah untuk grafik pai komposisi barang hilang di dasbor.
// @Tags Dashboard & Stats
// @Produce json
// @Success 200 {object} services.PieChartDataDTO
// @Failure 500 {object} map[string]string "Error: Gagal mengambil data komposisi barang"
// @Security BearerAuth
// @Router /stats/item-composition [get]
func (c *DashboardController) GetItemCompositionChart(ctx *gin.Context) {
	pieData, err := c.service.GetItemCompositionPieChartData()
	if err != nil {
		log.Printf("ERROR: Gagal mengambil data komposisi barang: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal mengambil data komposisi barang")
		return
	}
	ctx.JSON(http.StatusOK, pieData)
}