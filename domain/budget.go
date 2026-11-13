package domain

import (
	"errors"
	"time"
)

// Budget adalah representasi baris di tabel budgets (pos anggaran).
type Budget struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	MaxAmount   float64   `json:"max_amount"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	CreatedBy   uint64    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BudgetRepository adalah kontrak operasi database untuk tabel budgets.
type BudgetRepository interface {
	Create(budget *Budget) (*Budget, error)
	Update(budget *Budget) (*Budget, error)
	FindByID(id uint64) (*Budget, error)
	FindAll() ([]Budget, error)

	// GetRemainingAmount menghitung sisa dana pos:
	// max_amount dikurangi total pengajuan berstatus pending dan approved.
	// Pengajuan rejected tidak dihitung karena dananya tidak jadi dipakai.
	GetRemainingAmount(budgetID uint64) (float64, error)
}

// ── DTO ──────────────────────────────────────────────────────────────────

// BudgetRequest — body untuk POST /budgets dan PUT /budgets/:id.
// Tanggal dikirim sebagai string dengan format YYYY-MM-DD.
type BudgetRequest struct {
	Name        string  `json:"name"        validate:"required,min=3,max=150"`
	Description string  `json:"description" validate:"max=1000"`
	MaxAmount   float64 `json:"max_amount"  validate:"required,gt=0"`
	StartDate   string  `json:"start_date"  validate:"required"`
	EndDate     string  `json:"end_date"    validate:"required"`
}

// BudgetResponse — data pos anggaran yang dikirim ke client, lengkap dengan sisa dana.
type BudgetResponse struct {
	ID              uint64    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	MaxAmount       float64   `json:"max_amount"`
	RemainingAmount float64   `json:"remaining_amount"`
	StartDate       string    `json:"start_date"`
	EndDate         string    `json:"end_date"`
	CreatedBy       uint64    `json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
}

// ToResponse mengonversi Budget ke BudgetResponse.
// remaining dihitung terpisah lewat BudgetRepository.GetRemainingAmount.
func (b *Budget) ToResponse(remaining float64) *BudgetResponse {
	return &BudgetResponse{
		ID:              b.ID,
		Name:            b.Name,
		Description:     b.Description,
		MaxAmount:       b.MaxAmount,
		RemainingAmount: remaining,
		StartDate:       b.StartDate.Format("2006-01-02"),
		EndDate:         b.EndDate.Format("2006-01-02"),
		CreatedBy:       b.CreatedBy,
		CreatedAt:       b.CreatedAt,
	}
}

// ── Error khusus pos anggaran ────────────────────────────────────────────

var (
	ErrBudgetNotFound      = errors.New("pos anggaran tidak ditemukan")
	ErrBudgetPeriodInvalid = errors.New("tanggal akhir tidak boleh lebih awal dari tanggal mulai")
	ErrBudgetBelowUsed     = errors.New("batas dana baru lebih kecil dari dana yang sudah diajukan di pos ini")
)