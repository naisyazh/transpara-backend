package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/transpara/backend/app/model"
)

// ExpenseService adalah kontrak use case untuk pengajuan dana dan voting.
//
// EnsureApproved dipakai Developer 2 (modul transaksi) sebelum mencairkan
// dana: hanya pengajuan berstatus approved yang lolos.
type ExpenseService interface {
	Create(userID uint64, req *model.CreateExpenseRequest) (*model.ExpenseRequest, error)
	GetByID(id uint64) (*model.ExpenseRequest, error)
	// EnsureApproved mengembalikan pengajuan jika statusnya approved.
	// Status pending atau rejected mengembalikan model.ErrRequestNotApproved.
	EnsureApproved(id uint64) (*model.ExpenseRequest, error)
	List(status string) ([]model.ExpenseRequest, error)
	Vote(userID, requestID uint64, req *model.VoteRequest) (*model.ExpenseRequest, error)
}

type expenseService struct {
	budgetRepo  model.BudgetRepository
	expenseRepo model.ExpenseRequestRepository
	validate    *validator.Validate
}

// NewExpenseService membuat instance expenseService baru (dependency injection).
func NewExpenseService(
	budgetRepo model.BudgetRepository,
	expenseRepo model.ExpenseRequestRepository,
) ExpenseService {
	return &expenseService{
		budgetRepo:  budgetRepo,
		expenseRepo: expenseRepo,
		validate:    validator.New(),
	}
}

// Create — ajukan penggunaan dana. Anggota maupun bendahara boleh mengajukan.
func (uc *expenseService) Create(userID uint64, req *model.CreateExpenseRequest) (*model.ExpenseRequest, error) {
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
		return nil, model.ErrOutsideBudgetPeriod
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
		return nil, fmt.Errorf("%w (sisa dana: Rp%.2f)", model.ErrAmountExceedsRemaining, remaining)
	}

	// 5. Simpan dengan status awal pending
	return uc.expenseRepo.Create(&model.ExpenseRequest{
		BudgetID:    budget.ID,
		RequestedBy: userID,
		Title:       strings.TrimSpace(req.Title),
		Amount:      amount,
		Reason:      strings.TrimSpace(req.Reason),
		Status:      model.ExpenseStatusPending,
	})
}

// GetByID — detail satu pengajuan (termasuk status dan hitungan suara).
func (uc *expenseService) GetByID(id uint64) (*model.ExpenseRequest, error) {
	return uc.expenseRepo.FindByID(id)
}

// EnsureApproved — dipakai modul transaksi sebelum bendahara mencairkan dana.
// Mengembalikan pengajuan jika statusnya approved. Jika pending atau rejected
// dikembalikan model.ErrRequestNotApproved, jika id tidak ada dikembalikan
// model.ErrExpenseNotFound.
func (uc *expenseService) EnsureApproved(id uint64) (*model.ExpenseRequest, error) {
	req, err := uc.expenseRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if req.Status != model.ExpenseStatusApproved {
		return nil, model.ErrRequestNotApproved
	}
	return req, nil
}

// List — daftar pengajuan. status boleh kosong, atau pending/approved/rejected.
func (uc *expenseService) List(status string) ([]model.ExpenseRequest, error) {
	status = strings.ToLower(strings.TrimSpace(status))

	switch status {
	case "", model.ExpenseStatusPending, model.ExpenseStatusApproved, model.ExpenseStatusRejected:
		// status valid
	default:
		return nil, errors.New("status tidak valid, gunakan pending, approved, atau rejected")
	}

	return uc.expenseRepo.FindAll(status)
}

// Vote — anggota memberi suara setuju (yes) atau tolak (no).
func (uc *expenseService) Vote(userID, requestID uint64, req *model.VoteRequest) (*model.ExpenseRequest, error) {
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
	if current.Status != model.ExpenseStatusPending {
		return nil, model.ErrRequestNotPending
	}

	voted, err := uc.expenseRepo.HasVoted(requestID, userID)
	if err != nil {
		return nil, err
	}
	if voted {
		return nil, model.ErrAlreadyVoted
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
	decide := func(cur *model.ExpenseRequest, yesAfter, noAfter int) (string, error) {
		if cur.Status != model.ExpenseStatusPending {
			return "", model.ErrRequestNotPending
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
		return model.ExpenseStatusApproved
	case no*2 >= total:
		return model.ExpenseStatusRejected
	default:
		return model.ExpenseStatusPending
	}
}
