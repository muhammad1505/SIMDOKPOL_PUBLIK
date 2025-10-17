/**
 * FILE HEADER: internal/services/errors.go
 *
 * PURPOSE:
 * Mendefinisikan error-error standar yang digunakan di seluruh lapisan service.
 * Menggunakan variabel error global memungkinkan perbandingan error yang aman
 * (menggunakan `errors.Is`) di lapisan controller, menghindari penggunaan "magic strings".
 */
package services

import "errors"

var (
	// ErrAccessDenied dikembalikan ketika seorang pengguna mencoba melakukan aksi
	// yang tidak diizinkan oleh hak aksesnya.
	ErrAccessDenied = errors.New("akses ditolak")

	// ErrNotFound dikembalikan ketika sebuah record atau entitas tidak dapat
	// ditemukan di database.
	ErrNotFound = errors.New("data tidak ditemukan")

	// ErrInvalidCredentials dikembalikan saat proses login gagal karena
	// NRP atau kata sandi tidak cocok.
	ErrInvalidCredentials = errors.New("NRP atau kata sandi salah")

	// ErrAccountInactive dikembalikan saat pengguna yang mencoba login
	// memiliki akun yang berstatus non-aktif (soft-deleted).
	ErrAccountInactive = errors.New("akun Anda tidak aktif, silakan hubungi Super Admin")

	// ErrOldPasswordMismatch dikembalikan saat mengubah kata sandi tetapi
	// kata sandi lama yang dimasukkan tidak cocok.
	ErrOldPasswordMismatch = errors.New("kata sandi saat ini yang Anda masukkan salah")
)