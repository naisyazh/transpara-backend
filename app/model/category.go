package model

import (
	"errors"
	"time"
)

// Tipe transaksi dan kategori (sesuai ENUM di migrasi 008 dan 009).
const (
	TransactionTypeIncome  = "income"
	TransactionTypeExpense = "expense"
)

// Category adalah representasi baris di tabel categories.
type Category struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CategoryRepository adalah kontrak operasi database untuk kategori.
type CategoryRepository interface {
	Create(c *Category) (*Category, error)
	FindByID(id uint64) (*Category, error)
	// FindAll mengembalikan semua kategori; txType kosong berarti tanpa filter.
	FindAll(txType string) ([]Category, error)
}

// ── DTO ──────────────────────────────────────────────────────────────────

// CreateCategoryRequest — body untuk POST /categories.
type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
	Type string `json:"type" validate:"required,oneof=income expense"`
}

// ── Error khusus kategori ────────────────────────────────────────────────

var (
	ErrCategoryNotFound  = errors.New("kategori tidak ditemukan")
	ErrCategoryDuplicate = errors.New("kategori dengan nama dan tipe yang sama sudah ada")
)