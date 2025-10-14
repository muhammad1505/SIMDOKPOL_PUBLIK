package middleware

import (
	"net/http"
	"simdokpol/internal/models"

	"github.com/gin-gonic/gin"
)

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("currentUser")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak. Pengguna tidak terautentikasi."})
			c.Abort()
			return
		}

		currentUser, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak. Tipe data pengguna tidak valid."})
			c.Abort()
			return
		}

		// Gunakan konstanta untuk peran
		if currentUser.Peran != models.RoleSuperAdmin {
			if c.Request.Header.Get("Accept") == "application/json" {
				c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak. Anda tidak memiliki hak akses yang cukup."})
			} else {
				c.Redirect(http.StatusFound, "/")
			}
			c.Abort()
			return
		}

		c.Next()
	}
}