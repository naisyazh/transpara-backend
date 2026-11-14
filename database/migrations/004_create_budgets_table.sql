-- Migration: 004_create_budgets_table.sql
-- Deskripsi : Membuat tabel budgets (pos anggaran) untuk modul anggaran & voting
--
-- Catatan:
--   - Satu baris = satu pos anggaran: batas dana maksimal + periode berlaku
--   - Hanya bendahara yang boleh membuat pos anggaran (dicek di backend Go)
--   - Tipe created_by sama dengan users.id (BIGINT UNSIGNED) agar foreign key valid

CREATE TABLE IF NOT EXISTS budgets (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name        VARCHAR(150)    NOT NULL              COMMENT 'nama pos anggaran / kegiatan',
    description TEXT            NULL                  COMMENT 'keterangan tambahan (opsional)',
    max_amount  DECIMAL(15,2)   NOT NULL              COMMENT 'batas maksimal dana untuk pos ini',
    start_date  DATE            NOT NULL              COMMENT 'tanggal mulai periode pos',
    end_date    DATE            NOT NULL              COMMENT 'tanggal akhir periode pos',
    created_by  BIGINT UNSIGNED NOT NULL              COMMENT 'id bendahara pembuat pos',
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    KEY idx_budgets_created_by (created_by),
    CONSTRAINT fk_budgets_user FOREIGN KEY (created_by)
        REFERENCES users (id),
    CONSTRAINT chk_budgets_amount CHECK (max_amount > 0),
    CONSTRAINT chk_budgets_period CHECK (end_date >= start_date)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;