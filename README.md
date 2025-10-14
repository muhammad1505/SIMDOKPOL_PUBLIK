# Sistem Informasi Manajemen Dokumen Kepolisian (SIMDOKPOL)

Repositori ini berisi kode sumber untuk SIMDOKPOL, sebuah aplikasi web yang dirancang untuk membantu unit kepolisian dalam manajemen dan penerbitan surat keterangan secara efisien. Dibangun dengan arsitektur modern, fitur komprehensif, dan kemampuan untuk berjalan 100% offline.

## ✨ Fitur Utama

-   **Alur Setup Awal Terpandu:** Saat dijalankan pertama kali, aplikasi akan menampilkan halaman setup multi-langkah untuk mengonfigurasi detail instansi (KOP surat, nama kantor) dan membuat akun Super Admin pertama.

-   **Manajemen Dokumen Lengkap:**
    * Sistem penuh untuk Membuat, Membaca, Memperbarui, dan Menghapus (CRUD) surat keterangan, termasuk _soft delete_.
    * Fitur **Buat Ulang (Duplikat)** untuk meregenerasi surat keterangan dengan data yang sudah ada, mempercepat proses perpanjangan.

-   **Manajemen Pengguna Berbasis Peran:** Sistem administrasi pengguna dengan dua tingkat hak akses (Super Admin & Operator) dan fitur untuk menonaktifkan serta mengaktifkan kembali akun pengguna.

-   **Dasbor Analitik Real-Time:** Tampilan ringkasan data dengan kartu statistik dan grafik interaktif yang dinamis untuk memonitor aktivitas operasional secara harian, bulanan, dan tahunan.

-   **Formulir Cerdas & Dinamis:**
    * Input tanggal yang konsisten (`DD-MM-YYYY`) menggunakan datepicker, terlepas dari pengaturan browser.
    * Input data barang hilang interaktif menggunakan modal dengan field yang menyesuaikan jenis barang.
    * Pemilihan **Penerima Laporan** dan **Penanggung Jawab** secara dinamis, dengan sistem rekomendasi otomatis berdasarkan Regu petugas yang login.

-   **Penomoran Surat & Konfigurasi Dinamis:**
    * Sistem penomoran surat cerdas yang menghormati nomor register awal dari setup dan otomatis reset di awal tahun baru.
    * KOP surat, nama instansi, dan tempat surat sepenuhnya dinamis, diatur melalui halaman setup awal.

-   **Fitur Backup & Restore Database:**
    * Super Admin dapat melakukan backup seluruh database ke dalam satu file `.db` dan mengunduhnya.
    * Fitur restore yang aman dengan konfirmasi ganda untuk memulihkan data dari file backup.
    * Path (lokasi folder) untuk menyimpan backup dapat diatur melalui UI.

-   **Modul Audit Log Komprehensif:** Setiap aksi penting (pembuatan/pembaruan/penghapusan dokumen dan pengguna) dicatat secara otomatis. Super Admin dapat melihat riwayat lengkap aktivitas sistem.

-   **Pratinjau Cetak Presisi Tinggi:** Halaman pratinjau cetak yang dirancang agar 100% cocok dengan format fisik surat resmi, termasuk jenis font (`Courier New`) dan layout yang padat.

-   **Otentikasi & Otorisasi Aman:** Sistem login berbasis JWT yang disimpan dalam _HttpOnly Cookie_, dilengkapi dengan _middleware_ untuk melindungi rute berdasarkan status login dan peran pengguna.

-   **Aset 100% Offline:** Semua aset (font, CSS, JavaScript) disimpan secara lokal, memastikan aplikasi dapat berjalan lancar di lingkungan tanpa koneksi internet.

## 🌟 Stabilitas & Penyempurnaan

Proyek ini telah melalui beberapa iterasi perbaikan untuk meningkatkan kualitas dan keamanan:
-   **Perbaikan Celah Keamanan (IDOR):** Endpoint untuk melihat detail dokumen kini divalidasi berdasarkan hak akses pemilik atau Super Admin untuk mencegah kebocoran data.
-   **Perbaikan Bug Kritis:** Logika penomoran surat dan kalkulasi dasbor berdasarkan zona waktu telah diperbaiki untuk memastikan akurasi dan konsistensi data.
-   **Peningkatan User Experience (UX):** Tombol aksi di semua tabel kini lebih informatif (menampilkan caption saat hover), input tanggal yang konsisten, dan alur kerja duplikasi dokumen untuk efisiensi.

## 🛠️ Tumpukan Teknologi (Technology Stack)

-   **Backend:** Go (Golang)
-   **Web Framework:** Gin
-   **ORM:** GORM
-   **Database:** SQLite
-   **Frontend:** HTML5, CSS3, JavaScript
-   **Pustaka Frontend:** jQuery, Bootstrap, DataTables, Chart.js, SweetAlert2, Bootstrap Datepicker
-   **UI Template:** SB Admin 2
-   **Live Reloading:** Air

## 📂 Struktur Proyek

simdokpol/
├── cmd/                # Titik masuk aplikasi (main.go)
├── internal/           # Logika inti aplikasi
│   ├── controllers/
│   ├── middleware/
│   ├── models/
│   ├── repositories/
│   └── services/
├── web/                # Aset frontend (HTML Templates, CSS, JS)
├── backups/            # Folder default untuk menyimpan file backup
├── migrations/         # File migrasi skema database
├── .air.toml           # Konfigurasi untuk live-reload
├── .env.example        # Contoh file environment
├── go.mod              # Manajemen dependensi Go
└── README.md           # Dokumentasi Proyek


## 🚀 Memulai Proyek

### Prasyarat

-   **Go:** Versi 1.22 atau lebih baru.
-   **Git:** Untuk kloning repositori.
-   **Air:** Untuk live-reloading selama pengembangan.

### Instalasi & Menjalankan

1.  **Kloning Repositori:**
    ```bash
    git clone [URL_REPOSITORI_ANDA] simdokpol
    cd simdokpol
    ```
2.  **Konfigurasi Environment:**
    Salin file `.env.example` menjadi `.env` dan sesuaikan isinya jika perlu.
    ```bash
    cp .env.example .env
    ```
3.  **Instalasi Dependensi:**
    ```bash
    go mod tidy
    ```
4.  **Menjalankan di Mode Pengembangan:**
    ```bash
    air
    ```
    Aplikasi akan berjalan di `http://localhost:8080`.

**Catatan Penting:** Saat menjalankan pertama kali, aplikasi akan otomatis mengarahkan Anda ke halaman `http://localhost:8080/setup`. Ikuti langkah-langkah di sana untuk melakukan konfigurasi awal sistem dan membuat akun Super Admin pertama.

## 🛣️ Rencana Pengembangan (Roadmap)

-   [x] **Fondasi Backend & Arsitektur Bersih**
-   [x] **Alur Kerja Surat Keterangan Hilang (CRUD Lengkap)**
-   [x] **Otentikasi & Otorisasi Berbasis Peran**
-   [x] **Manajemen Pengguna (CRUD, Aktivasi/Deaktivasi)**
-   [x] **Dasbor Analitik dengan Grafik Dinamis**
-   [x] **Formulir Cerdas dengan Pemilihan Petugas Otomatis**
-   [x] **Alur Setup Awal Terpandu**
-   [x] **Fitur Pencarian Global**
-   [x] **Modul Audit Log**
-   [x] **Fitur Backup & Restore Database (dengan path dinamis)**
-   [x] **Konfigurasi Aplikasi Dinamis**
-   [x] **Fitur Duplikat Dokumen**
-   [x] **Pratinjau Cetak Presisi Tinggi & 100% Offline**
-   [ ] **Pembungkusan Desktop (Wails):** Transformasi aplikasi web menjadi aplikasi desktop mandiri.
-   [ ] **Penambahan Unit Test & Integration Test**