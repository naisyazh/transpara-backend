package model

import (
	"errors"
	"time"
)

// Status pengajuan dan pilihan suara (sesuai ENUM di migrasi 005 dan 006).
const (
	ExpenseStatusPending  = "pending"
	ExpenseStatusApproved = "approved"
	ExpenseStatusRejected = "rejected"

	VoteYes = "yes"
	VoteNo  = "no"
)

// ExpenseRequest adalah representasi baris di tabel expense_requests.
type ExpenseRequest struct {
	ID          uint64     `json:"id"`
	BudgetID    uint64     `json:"budget_id"`
	RequestedBy uint64     `json:"requested_by"`
	Title       string     `json:"title"`
	Amount      float64    `json:"amount"`
	Reason      string     `json:"reason"`
	Status      string     `json:"status"`
	YesVotes    int        `json:"yes_votes"`
	NoVotes     int        `json:"no_votes"`
	DecidedAt   *time.Time `json:"decided_at"` // null selama masih pending
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ExpenseVote adalah representasi baris di tabel expense_votes.
type ExpenseVote struct {
	ID               uint64    `json:"id"`
	ExpenseRequestID uint64    `json:"expense_request_id"`
	UserID           uint64    `json:"user_id"`
	Vote             string    `json:"vote"`
	CreatedAt        time.Time `json:"created_at"`
}

// DecideFunc dipanggil repository DI DALAM transaksi voting, setelah baris
// pengajuan terkunci. yesAfter/noAfter adalah hitungan suara SETELAH suara
// baru dihitung. Mengembalikan status baru, atau error untuk membatalkan
// (rollback) seluruh voting.
//
// Logika bisnis (ambang 50%) tetap di usecase, sedangkan repository
// hanya mengurus transaksi dan penguncian.
type DecideFunc func(current *ExpenseRequest, yesAfter, noAfter int) (string, error)

// ExpenseRequestRepository adalah kontrak operasi database untuk
// pengajuan dana dan suaranya.
type ExpenseRequestRepository interface {
	Create(req *ExpenseRequest) (*ExpenseRequest, error)
	FindByID(id uint64) (*ExpenseRequest, error)
	// FindAll mengembalikan semua pengajuan; status kosong berarti tanpa filter.
	FindAll(status string) ([]ExpenseRequest, error)

	// HasVoted mengecek apakah user sudah memberi suara pada pengajuan ini.
	HasVoted(requestID, userID uint64) (bool, error)

	// CountVotingMembers menghitung total anggota yang berhak vote
	// (user dengan is_verified = 1, bendahara ikut dihitung).
	CountVotingMembers() (int, error)

	// CastVote menyimpan suara dalam SATU transaksi: kunci baris pengajuan,
	// tentukan status baru lewat decide, insert vote, lalu update tally/status.
	CastVote(requestID, userID uint64, vote string, decide DecideFunc) (*ExpenseRequest, error)
}

// ── DTO ──────────────────────────────────────────────────────────────────

// CreateExpenseRequest — body untuk POST /expense-requests.
type CreateExpenseRequest struct {
	BudgetID uint64  `json:"budget_id" validate:"required"`
	Title    string  `json:"title"     validate:"required,min=3,max=150"`
	Amount   float64 `json:"amount"    validate:"required,gt=0"`
	Reason   string  `json:"reason"    validate:"required,min=5"`
}

// VoteRequest — body untuk POST /expense-requests/:id/vote.
type VoteRequest struct {
	Vote string `json:"vote" validate:"required,oneof=yes no"`
}

// ── Error khusus pengajuan dan voting ────────────────────────────────────

var (
	ErrExpenseNotFound        = errors.New("pengajuan dana tidak ditemukan")
	ErrOutsideBudgetPeriod    = errors.New("pengajuan di luar periode pos anggaran")
	ErrAmountExceedsRemaining = errors.New("nominal melebihi sisa dana pos anggaran")
	ErrRequestNotPending      = errors.New("pengajuan sudah diputuskan, voting sudah ditutup")
	ErrRequestNotApproved     = errors.New("pengajuan belum disetujui, dana belum bisa dicairkan")
	ErrAlreadyVoted           = errors.New("Anda sudah memberikan suara untuk pengajuan ini")
)
