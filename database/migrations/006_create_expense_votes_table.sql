-- Migration: 006_create_expense_votes_table.sql
-- Deskripsi : Membuat tabel expense_votes untuk menyimpan suara setuju/tolak
--
-- Catatan:
--   - Satu baris = satu suara dari satu user untuk satu pengajuan
--   - UNIQUE (expense_request_id, user_id) mencegah satu user vote dua kali,
--     dijaga di level database sehingga aman walau dua request masuk bersamaan
--   - Harus dijalankan SETELAH 005 karena ada foreign key ke expense_requests

CREATE TABLE IF NOT EXISTS expense_votes (
    id                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    expense_request_id BIGINT UNSIGNED NOT NULL              COMMENT 'pengajuan yang di-vote',
    user_id            BIGINT UNSIGNED NOT NULL              COMMENT 'user yang memberi suara',
    vote               ENUM('yes','no') NOT NULL             COMMENT 'yes = setuju, no = tolak',
    created_at         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_vote_once (expense_request_id, user_id),
    KEY idx_ev_user_id (user_id),
    CONSTRAINT fk_ev_request FOREIGN KEY (expense_request_id)
        REFERENCES expense_requests (id) ON DELETE CASCADE,
    CONSTRAINT fk_ev_user FOREIGN KEY (user_id)
        REFERENCES users (id)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;