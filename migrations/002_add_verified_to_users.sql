-- Migration: 002_add_verified_to_users.sql
-- Deskripsi : Menambahkan kolom is_verified ke tabel users
--             untuk membedakan akun yang sudah/belum verifikasi OTP

ALTER TABLE users
    ADD COLUMN is_verified TINYINT(1) NOT NULL DEFAULT 0
        COMMENT '0 = belum verifikasi, 1 = sudah verifikasi'
        AFTER password_hash;
