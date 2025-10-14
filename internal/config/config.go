package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// Config menampung semua variabel konfigurasi aplikasi.
type Config struct {
	JWTSecretKey string
	DBDSN        string
	BcryptCost   int // Biaya bcrypt yang sudah dihitung
}

// determineBcryptCost menjalankan benchmark kecil untuk menemukan biaya bcrypt yang optimal.
func determineBcryptCost() int {
	log.Println("[CONFIG] Menjalankan benchmark bcrypt untuk menentukan biaya optimal...")
	targetDuration := 150 * time.Millisecond // Target waktu kita: di bawah 150ms

	for cost := bcrypt.DefaultCost; cost >= bcrypt.MinCost; cost-- {
		startTime := time.Now()
		_, err := bcrypt.GenerateFromPassword([]byte("dummy-password-for-benchmark"), cost)
		duration := time.Since(startTime)

		if err == nil && duration < targetDuration {
			log.Printf("[CONFIG] Biaya bcrypt optimal diatur ke: %d (waktu: %v)\n", cost, duration)
			return cost
		}
	}

	// Fallback jika bahkan MinCost lebih lambat dari target
	log.Printf("[CONFIG] Perangkat terdeteksi sangat lambat. Menggunakan biaya bcrypt minimum: %d\n", bcrypt.MinCost)
	return bcrypt.MinCost
}

// Load memuat konfigurasi dari file .env dan environment variables.
func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: tidak dapat memuat file .env. Menggunakan environment variables sistem.")
	}

	// Jalankan benchmark untuk menentukan biaya bcrypt
	chosenBcryptCost := determineBcryptCost()

	cfg := &Config{
		JWTSecretKey: os.Getenv("JWT_SECRET_KEY"),
		DBDSN:        os.Getenv("DB_DSN"),
		BcryptCost:   chosenBcryptCost,
	}
	
	if cfg.JWTSecretKey == "" {
		log.Fatal("FATAL: JWT_SECRET_KEY tidak di-set di environment atau file .env")
	}
	if cfg.DBDSN == "" {
		log.Fatal("FATAL: DB_DSN tidak di-set di environment atau file .env")
	}

	return cfg, nil
}