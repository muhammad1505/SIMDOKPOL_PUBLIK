package controllers

import (
	"errors"
	"log"
	"net/http"
	"simdokpol/internal/models"
	"simdokpol/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// DocumentRequest adalah DTO untuk membuat atau memperbarui dokumen.
type DocumentRequest struct {
	NamaLengkap        string `json:"nama_lengkap" binding:"required" example:"BUDI SANTOSO"`
	TempatLahir        string `json:"tempat_lahir" binding:"required" example:"JAKARTA"`
	TanggalLahir       string `json:"tanggal_lahir" binding:"required" example:"1990-01-15"`
	JenisKelamin       string `json:"jenis_kelamin" binding:"required" enums:"Laki-laki,Perempuan"`
	Agama              string `json:"agama" binding:"required" example:"Islam"`
	Pekerjaan          string `json:"pekerjaan" binding:"required" example:"Karyawan Swasta"`
	Alamat             string `json:"alamat" binding:"required" example:"JL. MERDEKA NO. 10, JAKARTA"`
	LokasiHilang       string `json:"lokasi_hilang" binding:"required" example:"Sekitar Pasar Senen"`
	PetugasPelaporID   uint   `json:"petugas_pelapor_id" binding:"required" example:"2"`
	PejabatPersetujuID uint   `json:"pejabat_persetuju_id" binding:"required" example:"1"`
	Items              []struct {
		NamaBarang string `json:"nama_barang" binding:"required" example:"KTP"`
		Deskripsi  string `json:"deskripsi" example:"NIK: 3171234567890001"`
	} `json:"items" binding:"required,min=1"`
}

type LostDocumentController struct {
	docService services.LostDocumentService
}

func NewLostDocumentController(docService services.LostDocumentService) *LostDocumentController {
	return &LostDocumentController{docService: docService}
}

// @Summary Mendapatkan Dokumen Berdasarkan ID
// @Description Mengambil detail satu surat keterangan hilang berdasarkan ID-nya. Hanya bisa diakses oleh Super Admin atau operator yang membuat dokumen tersebut.
// @Tags Documents
// @Produce json
// @Param id path int true "ID Dokumen"
// @Success 200 {object} models.LostDocument
// @Failure 400 {object} map[string]string "Error: ID tidak valid"
// @Failure 403 {object} map[string]string "Error: Akses ditolak"
// @Failure 404 {object} map[string]string "Error: Dokumen tidak ditemukan"
// @Security BearerAuth
// @Router /documents/{id} [get]
func (c *LostDocumentController) FindByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "ID dokumen tidak valid")
		return
	}

	loggedInUserID := ctx.GetUint("userID")

	document, err := c.docService.FindByID(uint(id), loggedInUserID)
	if err != nil {
		if errors.Is(err, services.ErrAccessDenied) {
			APIError(ctx, http.StatusForbidden, "Akses ditolak: Anda tidak memiliki izin untuk melihat dokumen ini.")
			return
		}
		APIError(ctx, http.StatusNotFound, "Dokumen tidak ditemukan")
		return
	}

	ctx.JSON(http.StatusOK, document)
}

// @Summary Pencarian Dokumen Global
// @Description Mencari dokumen (aktif dan arsip) berdasarkan Nomor Surat atau Nama Pemohon.
// @Tags Documents
// @Produce json
// @Param q query string true "Kata Kunci Pencarian"
// @Success 200 {array} models.LostDocument
// @Failure 500 {object} map[string]string "Error: Terjadi kesalahan pada server"
// @Security BearerAuth
// @Router /search [get]
func (c *LostDocumentController) SearchGlobal(ctx *gin.Context) {
	query := ctx.Query("q")
	documents, err := c.docService.SearchGlobal(query)
	if err != nil {
		log.Printf("ERROR: Gagal melakukan pencarian global: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal melakukan pencarian dokumen.")
		return
	}
	ctx.JSON(http.StatusOK, documents)
}

// @Summary Mendapatkan Semua Dokumen
// @Description Mengambil daftar semua surat keterangan hilang, bisa difilter berdasarkan status (aktif/arsip) dan query pencarian.
// @Tags Documents
// @Produce json
// @Param q query string false "Kata Kunci Pencarian (No. Surat / Nama)"
// @Param status query string false "Filter status dokumen" enums(active, archived) default(active)
// @Success 200 {array} models.LostDocument
// @Failure 500 {object} map[string]string "Error: Terjadi kesalahan pada server"
// @Security BearerAuth
// @Router /documents [get]
func (c *LostDocumentController) FindAll(ctx *gin.Context) {
	query := ctx.Query("q")
	status := ctx.DefaultQuery("status", "active")

	documents, err := c.docService.FindAll(query, status)
	if err != nil {
		log.Printf("ERROR: Gagal mengambil data dokumen: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal mengambil data dokumen.")
		return
	}
	ctx.JSON(http.StatusOK, documents)
}

// @Summary Menghapus Dokumen
// @Description Menghapus (soft delete) sebuah surat keterangan hilang. Hanya bisa diakses oleh Super Admin atau operator yang membuatnya.
// @Tags Documents
// @Produce json
// @Param id path int true "ID Dokumen"
// @Success 200 {object} map[string]string "Pesan Sukses"
// @Failure 400 {object} map[string]string "Error: ID tidak valid"
// @Failure 403 {object} map[string]string "Error: Akses ditolak"
// @Failure 500 {object} map[string]string "Error: Gagal menghapus dokumen"
// @Security BearerAuth
// @Router /documents/{id} [delete]
func (c *LostDocumentController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "ID dokumen tidak valid")
		return
	}

	loggedInUserID := ctx.GetUint("userID")

	if err := c.docService.DeleteLostDocument(uint(id), loggedInUserID); err != nil {
		if errors.Is(err, services.ErrAccessDenied) {
			APIError(ctx, http.StatusForbidden, err.Error())
			return
		}
		log.Printf("ERROR: Gagal menghapus dokumen id %d: %v", id, err)
		APIError(ctx, http.StatusInternalServerError, "Gagal menghapus dokumen.")
		return
	}

	APIResponse(ctx, http.StatusOK, "Dokumen berhasil dihapus", nil)
}

// @Summary Memperbarui Dokumen
// @Description Memperbarui data sebuah surat keterangan hilang. Hanya bisa diakses oleh Super Admin atau operator yang membuatnya.
// @Tags Documents
// @Accept json
// @Produce json
// @Param id path int true "ID Dokumen"
// @Param document body DocumentRequest true "Data Dokumen yang Diperbarui"
// @Success 200 {object} models.LostDocument
// @Failure 400 {object} map[string]string "Error: Input tidak valid"
// @Failure 403 {object} map[string]string "Error: Akses ditolak"
// @Failure 500 {object} map[string]string "Error: Gagal memperbarui dokumen"
// @Security BearerAuth
// @Router /documents/{id} [put]
func (c *LostDocumentController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "ID dokumen tidak valid")
		return
	}

	var req DocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		APIError(ctx, http.StatusBadRequest, "Input tidak valid: "+err.Error())
		return
	}

	tglLahir, err := time.Parse("2006-01-02", req.TanggalLahir)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "Format Tanggal Lahir salah, gunakan YYYY-MM-DD")
		return
	}

	loggedInUserID := ctx.GetUint("userID")

	residentData := models.Resident{
		NamaLengkap:  req.NamaLengkap,
		TempatLahir:  req.TempatLahir,
		TanggalLahir: tglLahir,
		JenisKelamin: req.JenisKelamin,
		Agama:        req.Agama,
		Pekerjaan:    req.Pekerjaan,
		Alamat:       req.Alamat,
	}

	var lostItems []models.LostItem
	for _, item := range req.Items {
		lostItems = append(lostItems, models.LostItem{NamaBarang: item.NamaBarang, Deskripsi: item.Deskripsi})
	}

	updatedDoc, err := c.docService.UpdateLostDocument(uint(id), residentData, lostItems, req.LokasiHilang, req.PetugasPelaporID, req.PejabatPersetujuID, loggedInUserID)
	if err != nil {
		if errors.Is(err, services.ErrAccessDenied) {
			APIError(ctx, http.StatusForbidden, err.Error())
			return
		}
		log.Printf("ERROR: Gagal memperbarui dokumen id %d: %v", id, err)
		APIError(ctx, http.StatusInternalServerError, "Gagal memperbarui dokumen.")
		return
	}

	ctx.JSON(http.StatusOK, updatedDoc)
}

// @Summary Membuat Dokumen Baru
// @Description Membuat surat keterangan hilang baru.
// @Tags Documents
// @Accept json
// @Produce json
// @Param document body DocumentRequest true "Data Dokumen Baru"
// @Success 201 {object} models.LostDocument
// @Failure 400 {object} map[string]string "Error: Input tidak valid"
// @Failure 500 {object} map[string]string "Error: Gagal membuat dokumen"
// @Security BearerAuth
// @Router /documents [post]
func (c *LostDocumentController) Create(ctx *gin.Context) {
	var req DocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		APIError(ctx, http.StatusBadRequest, "Input tidak valid: "+err.Error())
		return
	}

	tglLahir, err := time.Parse("2006-01-02", req.TanggalLahir)
	if err != nil {
		APIError(ctx, http.StatusBadRequest, "Format Tanggal Lahir salah, gunakan YYYY-MM-DD")
		return
	}

	operatorID := ctx.GetUint("userID")

	residentData := models.Resident{
		NamaLengkap:  req.NamaLengkap,
		TempatLahir:  req.TempatLahir,
		TanggalLahir: tglLahir,
		JenisKelamin: req.JenisKelamin,
		Agama:        req.Agama,
		Pekerjaan:    req.Pekerjaan,
		Alamat:       req.Alamat,
	}

	var lostItems []models.LostItem
	for _, item := range req.Items {
		lostItems = append(lostItems, models.LostItem{NamaBarang: item.NamaBarang, Deskripsi: item.Deskripsi})
	}

	createdDoc, err := c.docService.CreateLostDocument(residentData, lostItems, operatorID, req.LokasiHilang, req.PetugasPelaporID, req.PejabatPersetujuID)
	if err != nil {
		log.Printf("ERROR: Gagal membuat dokumen: %v", err)
		APIError(ctx, http.StatusInternalServerError, "Gagal membuat dokumen.")
		return
	}

	ctx.JSON(http.StatusCreated, createdDoc)
}