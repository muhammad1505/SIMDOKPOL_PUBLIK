package models

// Konstanta untuk Peran Pengguna
const (
	RoleSuperAdmin = "SUPER_ADMIN"
	RoleOperator   = "OPERATOR"
)

// Konstanta untuk Aksi Audit Log
const (
	AuditCreateUser      = "BUAT PENGGUNA"
	AuditUpdateUser      = "UPDATE PENGGUNA"
	AuditDeactivateUser  = "NONAKTIFKAN PENGGUNA"
	AuditActivateUser    = "AKTIFKAN PENGGUNA"
	AuditCreateDocument  = "BUAT DOKUMEN"
	AuditUpdateDocument  = "UPDATE DOKUMEN"
	AuditDeleteDocument  = "HAPUS DOKUMEN"
	AuditSystemSetup     = "SETUP SISTEM"
	AuditBackupCreated   = "BUAT BACKUP"
	AuditRestoreFromFile = "PULIHKAN DARI FILE"
	AuditSettingsUpdated = "PERBARUI PENGATURAN"
)