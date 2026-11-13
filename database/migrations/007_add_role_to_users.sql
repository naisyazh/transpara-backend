-- Migration: 007_add_role_to_users.sql
-- Deskripsi : Menambahkan kolom role ke tabel users.
--             User baru otomatis berrole 'anggota'.
--             Role 'bendahara' hanya bisa disetel lewat database.

ALTER TABLE users
    ADD COLUMN role ENUM('anggota', 'bendahara') NOT NULL DEFAULT 'anggota'
        COMMENT 'anggota = user biasa, bendahara = pengelola keuangan'
        AFTER is_verified;