-- Migration: 003_create_otp_tokens_table.sql
-- Deskripsi : Membuat tabel otp_tokens untuk menyimpan OTP verifikasi
--
-- Catatan keamanan:
--   - Kolom otp_hash menyimpan HASH dari OTP, bukan OTP aslinya
--   - OTP asli hanya ada di response API (mode simulasi) dan tidak disimpan
--   - attempt_count membatasi percobaan verifikasi yang salah (maks 5x)

CREATE TABLE IF NOT EXISTS otp_tokens (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id       BIGINT UNSIGNED NOT NULL,
    otp_hash      VARCHAR(255)    NOT NULL              COMMENT 'bcrypt hash dari OTP 6 digit',
    expires_at    DATETIME        NOT NULL              COMMENT 'waktu kadaluarsa OTP (5 menit)',
    is_used       TINYINT(1)      NOT NULL DEFAULT 0    COMMENT '0 = belum dipakai, 1 = sudah dipakai',
    attempt_count TINYINT         NOT NULL DEFAULT 0    COMMENT 'jumlah percobaan verifikasi yang salah',
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    KEY idx_otp_user_id (user_id),
    CONSTRAINT fk_otp_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;
