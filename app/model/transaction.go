package model

import (
	"errors"
	"time"
)

// Transaction adalah representasi baris di tabel transactions.
type Transaction struct {
	ID               uint64    `json:"id"`
	CategoryID       uint64    `json:"category_id"`
	CategoryName     string    `json:"category_name"` // hasil JOIN ke categories
	CreatedBy        uint64    `json:"created_by"`
	ExpenseRequestID *uint64   `json:"expense_request_id"` // null untuk income
	Type             string    `json:"type"`
	Amount           float64   `json:"amount"`
	Description      string    `json:"description"`
	TransactionDate  time.Time `json:"transaction_date"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TransactionFilter dipakai untuk menyaring daftar transaksi.
// Field yang kosong (zero value) berarti tidak difilter.
type TransactionFilter struct {
	Type       string
	CategoryID uint64
	From       *time.Time // tanggal awal, ikut dihitung
	To         *time.Time // tanggal akhir, ikut dihitung
}

// BalanceSummary adalah ringkasan kas: total pemasukan, pengeluaran, dan saldo.
type BalanceSummary struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	Balance      float64 `json:"balance"` // TotalIncome - TotalExpense
}

// TransactionRepository adalah kontrak operasi database untuk transaksi.
type TransactionRepository interface {
	// Create menyimpan transaksi baru. Jika expense_request_id sudah pernah
	// dipakai, mengembalikan ErrAlreadyDisbursed (dijaga UNIQUE KEY).
	Create(t *Transaction) (*Transaction, error)
	FindByID(id uint64) (*Transaction, error)
	FindAll(filter TransactionFilter) ([]Transaction, error)
	// GetSummary menghitung total pemasukan, pengeluaran, dan saldo.
	GetSummary() (*BalanceSummary, error)
}

// ── DTO ──────────────────────────────────────────────────────────────────

// CreateTransactionRequest — body untuk POST /transactions.
//
//   - income : amount wajib diisi, expense_request_id harus kosong.
//   - expense: expense_request_id wajib diisi, amount diabaikan karena
//     nominal diambil dari pengajuan yang sudah disetujui.
//
// Aturan per tipe dicek di service. transaction_date berformat YYYY-MM-DD.
type CreateTransactionRequest struct {
	Type             string  `json:"type"               validate:"required,oneof=income expense"`
	CategoryID       uint64  `json:"category_id"        validate:"required"`
	ExpenseRequestID *uint64 `json:"expense_request_id"`
	Amount           float64 `json:"amount"             validate:"omitempty,gt=0"`
	Description      string  `json:"description"        validate:"required,min=3,max=255"`
	TransactionDate  string  `json:"transaction_date"   validate:"required"`
}

// ── Error khusus transaksi ───────────────────────────────────────────────

var (
	ErrTransactionNotFound     = errors.New("transaksi tidak ditemukan")
	ErrAlreadyDisbursed        = errors.New("pengajuan ini sudah pernah dicairkan")
	ErrExpenseRequestRequired  = errors.New("pengeluaran wajib terhubung ke pengajuan dana yang disetujui")
	ErrIncomeHasExpenseRequest = errors.New("pemasukan tidak boleh terhubung ke pengajuan dana")
	ErrAmountRequired          = errors.New("nominal wajib diisi untuk pemasukan")
	ErrCategoryTypeMismatch    = errors.New("tipe kategori tidak sesuai dengan tipe transaksi")
	ErrInvalidDate             = errors.New("format tanggal tidak valid, gunakan YYYY-MM-DD")
)