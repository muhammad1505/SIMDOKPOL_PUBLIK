/**
 * FILE HEADER: internal/controllers/helpers.go
 *
 * PURPOSE:
 * Menyediakan fungsi-fungsi helper untuk standarisasi respons API JSON.
 * Dengan menggunakan helper ini, semua respons (baik sukses maupun error)
 * akan memiliki format yang konsisten di seluruh aplikasi.
 */
package controllers

import "github.com/gin-gonic/gin"

// APIResponse mengirimkan respons JSON standar untuk operasi yang sukses.
//
// PARAMETERS:
// - ctx (*gin.Context): Konteks request Gin.
// - statusCode (int): Kode status HTTP (misalnya, 200, 201).
// - message (string): Pesan singkat yang mendeskripsikan hasil.
// - data (interface{}): Payload data opsional yang akan disertakan dalam respons.
func APIResponse(ctx *gin.Context, statusCode int, message string, data interface{}) {
	response := gin.H{"message": message}
	if data != nil {
		response["data"] = data
	}
	ctx.JSON(statusCode, response)
}

// APIError mengirimkan respons JSON standar untuk error.
//
// PARAMETERS:
// - ctx (*gin.Context): Konteks request Gin.
// - statusCode (int): Kode status HTTP error (misalnya, 400, 403, 404, 500).
// - errorMessage (string): Pesan error yang aman untuk ditampilkan ke klien.
func APIError(ctx *gin.Context, statusCode int, errorMessage string) {
	ctx.JSON(statusCode, gin.H{"error": errorMessage})
}