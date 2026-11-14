-- Migration: 008_create_categories_table.sql
-- Deskripsi : Membuat tabel categories (kategori transaksi keuangan)
--
-- Catatan:
--   - type menentukan kategori ini untuk pemasukan (income) atau pengeluaran (expense)
--   - Tidak ada group_id karena sistem ini untuk satu kelompok
--   - UNIQUE (name, type) mencegah kategori kembar dalam tipe yang sama
--   - Harus dijalankan SEBELUM 009 karena transactions punya foreign key ke sini

CREATE TABLE IF NOT EXISTS categories (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name       VARCHAR(100)    NOT NULL              COMMENT 'nama kategori',
    type       ENUM('income','expense') NOT NULL     COMMENT 'income = pemasukan, expense = pengeluaran',
    created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_category_name_type (name, type)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

-- Kategori awal (boleh diubah sesuai kebutuhan kelompok)
INSERT IGNORE INTO categories (name, type) VALUES
    ('Iuran Anggota', 'income'),
    ('Donasi',        'income'),
    ('Lain-lain',     'income'),
    ('Konsumsi',      'expense'),
    ('Perlengkapan',  'expense'),
    ('Kegiatan',      'expense'),
    ('Lain-lain',     'expense');