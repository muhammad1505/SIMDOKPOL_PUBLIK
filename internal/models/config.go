package models

// Configuration menyimpan pengaturan sistem dalam format key-value.
type Configuration struct {
	Key   string `gorm:"primaryKey;size:255" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}