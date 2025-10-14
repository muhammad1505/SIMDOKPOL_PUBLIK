package models

import (
	"time"
	"gorm.io/gorm"
)

// User merepresentasikan model pengguna dalam sistem.
type User struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	NamaLengkap string         `gorm:"size:255;not null" json:"nama_lengkap"`
	NRP         string         `gorm:"size:20;not null;unique" json:"nrp"`
	KataSandi   string         `gorm:"size:255;not null" json:"-"` // Kata sandi tidak diekspos di JSON
	Pangkat     string         `gorm:"size:100" json:"pangkat"`
	Peran       string         `gorm:"size:50;not null;default:'OPERATOR'" json:"peran"` // SUPER_ADMIN, OPERATOR
	Jabatan     string         `gorm:"size:100" json:"jabatan"` // KANIT SPKT, ANGGOTA JAGA REGU
	Regu        string         `gorm:"size:10" json:"regu"` // I, II, III
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Resident merepresentasikan model penduduk/pemohon.
type Resident struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	NIK          string         `gorm:"size:16;not null;unique" json:"nik"`
	NamaLengkap  string         `gorm:"size:255;not null" json:"nama_lengkap"`
	TempatLahir  string         `gorm:"size:100;not null" json:"tempat_lahir"`
	TanggalLahir time.Time      `gorm:"not null" json:"tanggal_lahir"`
	JenisKelamin string         `gorm:"size:20;not null" json:"jenis_kelamin"`
	Agama        string         `gorm:"size:50;not null" json:"agama"`
	Pekerjaan    string         `gorm:"size:100;not null" json:"pekerjaan"`
	Alamat       string         `gorm:"type:text;not null" json:"alamat"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// LostDocument diperbarui dengan field OperatorID dan LastUpdatedByID
type LostDocument struct {
	ID                 uint           `gorm:"primarykey" json:"id"`
	NomorSurat         string         `gorm:"size:255;not null;unique" json:"nomor_surat"`
	TanggalLaporan     time.Time      `gorm:"not null" json:"tanggal_laporan"`
	Status             string         `gorm:"size:50;not null;default:'DITERBITKAN'" json:"status"`
	LokasiHilang       string         `gorm:"type:text" json:"lokasi_hilang"`
	
	ResidentID         uint           `gorm:"not null" json:"resident_id"`
	Resident           Resident       `gorm:"foreignKey:ResidentID" json:"resident"`
	
	LostItems          []LostItem     `gorm:"foreignKey:LostDocumentID" json:"lost_items"`
	
	// Petugas yang namanya akan tercetak di surat
	PetugasPelaporID   uint           `gorm:"not null" json:"petugas_pelapor_id"`
	PetugasPelapor     User           `gorm:"foreignKey:PetugasPelaporID" json:"petugas_pelapor"`
	
	PejabatPersetujuID *uint          `json:"pejabat_persetuju_id"`
	PejabatPersetuju   User           `gorm:"foreignKey:PejabatPersetujuID" json:"pejabat_persetuju"`
	
	// --- FIELD BARU UNTUK AUDIT ---
	// Operator adalah pengguna yang login dan melakukan aksi Create
	OperatorID         uint           `gorm:"not null" json:"operator_id"`
	Operator           User           `gorm:"foreignKey:OperatorID" json:"operator"`
	
	// LastUpdatedByID adalah pengguna yang login dan melakukan aksi Update terakhir
	LastUpdatedByID    *uint          `json:"last_updated_by_id"`
	LastUpdatedBy      User           `gorm:"foreignKey:LastUpdatedByID" json:"last_updated_by"`

	TanggalPersetujuan *time.Time     `json:"tanggal_persetujuan"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

// LostItem merepresentasikan barang yang hilang.
type LostItem struct {
	ID             uint   `gorm:"primarykey" json:"id"`
	LostDocumentID uint   `gorm:"not null" json:"lost_document_id"`
	NamaBarang     string `gorm:"size:255;not null" json:"nama_barang"`
	Deskripsi      string `gorm:"type:text" json:"deskripsi"`
}

// AuditLog untuk mencatat aktivitas penting.
type AuditLog struct {
	ID        uint      `gorm:"primarykey"`
	UserID    uint      `gorm:"not null"`
	User      User      `gorm:"foreignKey:UserID"`
	Aksi      string    `gorm:"size:255;not null"`
	Detail    string    `gorm:"type:text"`
	Timestamp time.Time `gorm:"not null"`
}