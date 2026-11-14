-- Migration: 009_create_transactions_table.sql
-- Deskripsi : Membuat tabel transactions (pemasukan dan pengeluaran kas)
--
-- Catatan:
--   - Pengeluaran (expense) WAJIB terhubung ke pengajuan berstatus approved
--     lewat expense_request_id. Statusnya dicek di service (EnsureApproved).
--   - Pemasukan (income) tidak punya pengajuan, expense_request_id harus NULL
--   - UNIQUE pada expense_request_id: satu pengajuan hanya bisa dicairkan sekali
--     (NULL boleh banyak, jadi tidak mengganggu pemasukan)
--   - Nominal pengeluaran diambil dari pengajuan, bukan diinput bebas
--   - Harus dijalankan SETELAH 005 dan 008 karena ada foreign key ke
--     expense_requests dan categories

CREATE TABLE IF NOT EXISTS transactions (
    id                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    category_id        BIGINT UNSIGNED NOT NULL              COMMENT 'kategori transaksi',
    created_by         BIGINT UNSIGNED NOT NULL              COMMENT 'id user (bendahara) yang mencatat',
    expense_request_id BIGINT UNSIGNED NULL                  COMMENT 'pengajuan yang dicairkan, wajib untuk expense',
    type               ENUM('income','expense') NOT NULL     COMMENT 'income = pemasukan, expense = pengeluaran',
    amount             DECIMAL(15,2)   NOT NULL              COMMENT 'nominal transaksi',
    description        VARCHAR(255)    NOT NULL              COMMENT 'keterangan transaksi',
    transaction_date   DATE            NOT NULL              COMMENT 'tanggal transaksi terjadi',
    created_at         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_tx_expense_request (expense_request_id),
    KEY idx_tx_category (category_id),
    KEY idx_tx_created_by (created_by),
    KEY idx_tx_type_date (type, transaction_date),
    CONSTRAINT fk_tx_category FOREIGN KEY (category_id)
        REFERENCES categories (id),
    CONSTRAINT fk_tx_user FOREIGN KEY (created_by)
        REFERENCES users (id),
    CONSTRAINT fk_tx_expense_request FOREIGN KEY (expense_request_id)
        REFERENCES expense_requests (id),
    CONSTRAINT chk_tx_amount CHECK (amount > 0),
    CONSTRAINT chk_tx_type_request CHECK (
        (type = 'expense' AND expense_request_id IS NOT NULL) OR
        (type = 'income'  AND expense_request_id IS NULL)
    )
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;