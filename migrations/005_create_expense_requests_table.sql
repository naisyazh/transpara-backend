-- Migration: 005_create_expense_requests_table.sql
-- Deskripsi : Membuat tabel expense_requests (pengajuan penggunaan uang)
--
-- Catatan:
--   - Anggota atau bendahara mengajukan dana: nominal, alasan, dan pos anggaran terkait
--   - status: pending (menunggu voting) -> approved / rejected (otomatis oleh backend)
--   - yes_votes dan no_votes adalah hitungan suara; baris suara asli ada di expense_votes
--   - Bendahara hanya boleh mencairkan dana jika status = approved (dicek di modul transaksi)
--   - Harus dijalankan SETELAH 004 karena ada foreign key ke budgets

CREATE TABLE IF NOT EXISTS expense_requests (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    budget_id    BIGINT UNSIGNED NOT NULL              COMMENT 'pos anggaran yang dipakai',
    requested_by BIGINT UNSIGNED NOT NULL              COMMENT 'id user pengaju',
    title        VARCHAR(150)    NOT NULL              COMMENT 'judul singkat pengajuan',
    amount       DECIMAL(15,2)   NOT NULL              COMMENT 'nominal dana yang diajukan',
    reason       TEXT            NOT NULL              COMMENT 'alasan pengajuan',
    status       ENUM('pending','approved','rejected') NOT NULL DEFAULT 'pending',
    yes_votes    INT UNSIGNED    NOT NULL DEFAULT 0    COMMENT 'jumlah suara setuju',
    no_votes     INT UNSIGNED    NOT NULL DEFAULT 0    COMMENT 'jumlah suara tolak',
    decided_at   DATETIME        NULL                  COMMENT 'waktu status berubah jadi approved/rejected',
    created_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    KEY idx_er_budget (budget_id),
    KEY idx_er_requested_by (requested_by),
    KEY idx_er_status (status),
    CONSTRAINT fk_er_budget FOREIGN KEY (budget_id)
        REFERENCES budgets (id),
    CONSTRAINT fk_er_user FOREIGN KEY (requested_by)
        REFERENCES users (id),
    CONSTRAINT chk_er_amount CHECK (amount > 0)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;