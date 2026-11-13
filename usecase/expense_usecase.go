package usecase

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/transpara/backend/domain"
)

// ExpenseUseCase adalah kontrak use case untuk pengajuan dana dan voting.
//
// GetByID juga dipakai Developer 2 (modul transaksi) untuk memblokir
// pencairan dana: tolak kalau Status != domain.ExpenseStatusApproved.
type ExpenseUseCase interface {
	Create(userID uint64, req *domain.CreateExpenseRequest) (*domain.ExpenseRequest, error)
	GetByID(id uint64) (*domain.ExpenseRequest, error)
	List(status string) ([]domain.ExpenseRequest, error)
	Vote(userID, requestID uint64, req *domain.VoteRequest) (*domain.ExpenseRequest, error)
}

type expenseUseCase struct {
	budgetRepo  domain.BudgetRepository
	expenseRepo domain.ExpenseRequestRepository
	validate    *validator.Validate
}

// NewExpenseUseCase membuat instance expenseUseCase baru (dependency injection).
func NewExpenseUseCase(
	budgetRepo domain.BudgetRepository,
	expenseRepo domain.ExpenseRequestRepository,
) ExpenseUseCase {
	return &expenseUseCase{
		budgetRepo:  budgetRepo,
		expenseRepo: expenseRepo,
		validate:    validator.New(),
	}
}

// Create — ajukan penggunaan dana. Anggota maupun bendahara boleh mengajukan.
func (uc *expenseUseCase) Create(userID uint64, req *domain.CreateExpenseRequest) (*domain.ExpenseRequest, error) {
	// 1. Validasi input
	if err := uc.validate.Struct(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return nil, errors.New(buildValidationMessage(ve[0]))
		}
		return nil, errors.New("input tidak valid")
	}

	// 2. Pos anggaran harus ada
	budget, err := uc.budgetRepo.FindByID(req.BudgetID)
	if err != nil {
		return nil, err
	}

	// 3. Tanggal hari ini harus berada di dalam periode pos (batas akhir ikut dihitung)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	if today.Before(budget.StartDate) || today.After(budget.EndDate) {
		return nil, domain.ErrOutsideBudgetPeriod
	}

	// 4. Nominal tidak boleh melebihi sisa dana.
	// Repository mengecek ulang di dalam transaksi sebagai pengaman race condition;
	// cek di sini supaya aturan bisnisnya terbaca jelas dan pesan errornya informatif.
	amount := roundRupiah(req.Amount)
	remaining, err := uc.budgetRepo.GetRemainingAmount(budget.ID)
	if err != nil {
		return nil, err
	}
	if amount > remaining {
		return nil, fmt.Errorf("%w (sisa dana: Rp%.2f)", domain.ErrAmountExceedsRemaining, remaining)
	}

	// 5. Simpan dengan status awal pending
	return uc.expenseRepo.Create(&domain.ExpenseRequest{
		BudgetID:    budget.ID,
		RequestedBy: userID,
		Title:       strings.TrimSpace(req.Title),
		Amount:      amount,
		Reason:      strings.TrimSpace(req.Reason),
		Status:      domain.ExpenseStatusPending,
	})
}

// GetByID — detail satu pengajuan (termasuk status dan hitungan suara).
func (uc *expenseUseCase) GetByID(id uint64) (*domain.ExpenseRequest, error) {
	return uc.expenseRepo.FindByID(id)
}

// List — daftar pengajuan. status boleh kosong, atau pending/approved/rejected.
func (uc *expenseUseCase) List(status string) ([]domain.ExpenseRequest, error) {
	status = strings.ToLower(strings.TrimSpace(status))

	switch status {
	case "", domain.ExpenseStatusPending, domain.ExpenseStatusApproved, domain.ExpenseStatusRejected:
		// status valid
	default:
		return nil, errors.New("status tidak valid, gunakan pending, approved, atau rejected")
	}

	return uc.expenseRepo.FindAll(status)
}

// Vote — anggota memberi suara setuju (yes) atau tolak (no).
func (uc *expenseUseCase) Vote(userID, requestID uint64, req *domain.VoteRequest) (*domain.ExpenseRequest, error) {
	// 1. Validasi input
	if err := uc.validate.Struct(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return nil, errors.New(buildValidationMessage(ve[0]))
		}
		return nil, errors.New("input tidak valid")
	}

	// 2. Pengecekan awal supaya error cepat dan jelas.
	// Penjaga sesungguhnya ada di dalam transaksi (status dicek ulang di decide,
	// dan UNIQUE KEY di database menolak vote ganda).
	current, err := uc.expenseRepo.FindByID(requestID)
	if err != nil {
		return nil, err
	}
	if current.Status != domain.ExpenseStatusPending {
		return nil, domain.ErrRequestNotPending
	}

	voted, err := uc.expenseRepo.HasVoted(requestID, userID)
	if err != nil {
		return nil, err
	}
	if voted {
		return nil, domain.ErrAlreadyVoted
	}

	// 3. Total anggota = dasar hitungan ambang 50%
	totalMembers, err := uc.expenseRepo.CountVotingMembers()
	if err != nil {
		return nil, err
	}
	if totalMembers <= 0 {
		return nil, errors.New("belum ada anggota terverifikasi untuk voting")
	}

	// 4. decide dijalankan repository DI DALAM transaksi, setelah baris terkunci
	decide := func(cur *domain.ExpenseRequest, yesAfter, noAfter int) (string, error) {
		if cur.Status != domain.ExpenseStatusPending {
			return "", domain.ErrRequestNotPending
		}
		return decideVoteStatus(yesAfter, noAfter, totalMembers), nil
	}

	return uc.expenseRepo.CastVote(requestID, userID, req.Vote, decide)
}

// decideVoteStatus menentukan status pengajuan dari hitungan suara.
//
//   - approved : suara setuju melewati 50% dari total anggota (yes*2 > total)
//   - rejected : suara tolak sudah mencapai 50% (no*2 >= total), sehingga
//     suara setuju tidak mungkin lagi melewati 50%
//   - pending  : selain itu, voting masih berjalan
//
// Contoh total 4 anggota: 3 setuju -> approved; 2 tolak -> rejected;
// 2 setuju + 1 tolak -> masih pending.
func decideVoteStatus(yes, no, total int) string {
	switch {
	case yes*2 > total:
		return domain.ExpenseStatusApproved
	case no*2 >= total:
		return domain.ExpenseStatusRejected
	default:
		return domain.ExpenseStatusPending
	}
}