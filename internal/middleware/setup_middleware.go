package middleware

import (
	"net/http"
	"simdokpol/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
)

// SetupMiddleware memeriksa apakah aplikasi sudah di-setup.
// Jika belum, semua request akan dialihkan ke halaman setup.
func SetupMiddleware(configService services.ConfigService) gin.HandlerFunc {
	return func(c *gin.Context) {
		isSetup, err := configService.IsSetupComplete()
		if err != nil {
			// Jika ada error saat cek konfigurasi, tampilkan halaman error.
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"Title": "Error Konfigurasi",
				"ErrorMessage": "Tidak dapat memverifikasi konfigurasi aplikasi: " + err.Error(),
			})
			c.Abort()
			return
		}

		// Jika setup belum selesai
		if !isSetup {
			// Izinkan akses hanya ke halaman setup, API-nya, dan aset statis
			allowedPaths := []string{"/setup", "/api/setup", "/static/"}
			isAllowed := false
			for _, path := range allowedPaths {
				if strings.HasPrefix(c.Request.URL.Path, path) {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				// Alihkan semua request lain ke halaman setup
				c.Redirect(http.StatusFound, "/setup")
				c.Abort()
				return
			}
		}

		// Jika setup sudah selesai, lanjutkan request
		c.Next()
	}
}