-- Perintah untuk membuat semua tabel (Migrasi NAIK)

CREATE TABLE `users` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `nama_lengkap` text NOT NULL,
    `nrp` text NOT NULL UNIQUE,
    `kata_sandi` text NOT NULL,
    `pangkat` text,
    `peran` text NOT NULL DEFAULT 'OPERATOR',
    `jabatan` text,
    `regu` text,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime
);
CREATE INDEX `idx_users_deleted_at` ON `users`(`deleted_at`);

CREATE TABLE `residents` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `nik` text NOT NULL UNIQUE,
    `nama_lengkap` text NOT NULL,
    `tempat_lahir` text NOT NULL,
    `tanggal_lahir` datetime NOT NULL,
    `jenis_kelamin` text NOT NULL,
    `agama` text NOT NULL,
    `pekerjaan` text NOT NULL,
    `alamat` text NOT NULL,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime
);
CREATE INDEX `idx_residents_deleted_at` ON `residents`(`deleted_at`);

CREATE TABLE `lost_documents` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `nomor_surat` text NOT NULL UNIQUE,
    `tanggal_laporan` datetime NOT NULL,
    `status` text NOT NULL DEFAULT 'DITERBITKAN',
    `lokasi_hilang` text,
    `resident_id` integer NOT NULL,
    `petugas_pelapor_id` integer NOT NULL,
    `pejabat_persetuju_id` integer,
    `operator_id` integer NOT NULL,
    `last_updated_by_id` integer,
    `tanggal_persetujuan` datetime,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime,
    FOREIGN KEY (`pejabat_persetuju_id`) REFERENCES `users`(`id`),
    FOREIGN KEY (`operator_id`) REFERENCES `users`(`id`),
    FOREIGN KEY (`last_updated_by_id`) REFERENCES `users`(`id`),
    FOREIGN KEY (`resident_id`) REFERENCES `residents`(`id`),
    FOREIGN KEY (`petugas_pelapor_id`) REFERENCES `users`(`id`)
);
CREATE INDEX `idx_lost_documents_deleted_at` ON `lost_documents`(`deleted_at`);

CREATE TABLE `lost_items` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `lost_document_id` integer NOT NULL,
    `nama_barang` text NOT NULL,
    `deskripsi` text,
    FOREIGN KEY (`lost_document_id`) REFERENCES `lost_documents`(`id`)
);

CREATE TABLE `audit_logs` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `user_id` integer NOT NULL,
    `aksi` text NOT NULL,
    `detail` text,
    `timestamp` datetime NOT NULL,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
);

CREATE TABLE `configurations` (
    `key` text PRIMARY KEY,
    `value` text
);