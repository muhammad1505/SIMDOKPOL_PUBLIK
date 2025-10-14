package middleware

import (
	"fmt"
	"net/http"
	"simdokpol/internal/repositories" // <-- IMPORT BARU
	"simdokpol/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Middleware sekarang menerima UserRepository untuk mengambil data pengguna
func AuthMiddleware(userRepo repositories.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")

		if err != nil {
			if strings.Contains(err.Error(), "named cookie not present") {
				// Untuk request halaman, redirect ke login
				if !strings.HasPrefix(c.Request.URL.Path, "/api") {
					c.Redirect(http.StatusFound, "/login")
					c.Abort()
					return
				}
				// Untuk request API, kirim JSON error
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Diperlukan otorisasi"})
				c.Abort()
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "Request tidak valid"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("signing method tidak terduga: %v", token.Header["alg"])
			}
			return services.JWTSecretKey, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID := uint(claims["userID"].(float64))

			// Ambil data lengkap pengguna dan simpan di context
			user, err := userRepo.FindByID(userID)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Pengguna tidak ditemukan"})
				c.Abort()
				return
			}
			c.Set("userID", userID)
			c.Set("currentUser", user) // Simpan objek user lengkap

			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
			c.Abort()
		}
	}
}