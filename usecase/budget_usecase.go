package usecase

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/transpara/backend/domain"
)

const budgetDateLayout = "2006-01-02"

// BudgetUseCase adalah kontrak use case untuk pos anggaran.
// Pembatasan "hanya bendahara" dilakukan di router (Tahap 6), bukan di sini.
type BudgetUseCase interface {
	Create(userID uint64, req *domain.BudgetRequest) (*domain.BudgetResponse, error)
	Update(id uint64, req *domain.BudgetRequest) (*domain.BudgetResponse, error)
	GetByID(id uint64) (*domain.BudgetResponse, error)
	List() ([]domain.BudgetResponse, error)
}

type budgetUseCase struct {
	budgetRepo domain.BudgetRepository
	validate   *validator.Validate
}

// NewBudgetUseCase membuat instance budgetUseCase baru (dependency injection).
func NewBudgetUseCase(budgetRepo domain.BudgetRepository) BudgetUseCase {
	return &budgetUseCase{
		budgetRepo: budgetRepo,
		validate:   validator.New(),
	}
}

// Create — buat pos anggaran baru. userID adalah bendahara pembuatnya.
func (uc *budgetUseCase) Create(userID uint64, req *domain.BudgetRequest) (*domain.BudgetResponse, error) {
	budget, err := uc.buildBudget(req)
	if err != nil {
		return nil, err
	}
	budget.CreatedBy = userID

	saved, err := uc.budgetRepo.Create(budget)
	if err != nil {
		return nil, err
	}

	// Pos baru belum punya pengajuan, jadi sisa dana = batas dana.
	return saved.ToResponse(saved.MaxAmount), nil
}

// Update — ubah pos anggaran. Batas dana baru tidak boleh lebih kecil
// dari dana yang sudah diajukan (pending + approved) di pos ini.
func (uc *budgetUseCase) Update(id uint64, req *domain.BudgetRequest) (*domain.BudgetResponse, error) {
	existing, err := uc.budgetRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	updated, err := uc.buildBudget(req)
	if err != nil {
		return nil, err
	}
	updated.ID = id

	remaining, err := uc.budgetRepo.GetRemainingAmount(id)
	if err != nil {
		return nil, err
	}
	used := roundRupiah(existing.MaxAmount - remaining)
	if updated.MaxAmount < used {
		return nil, domain.ErrBudgetBelowUsed
	}

	saved, err := uc.budgetRepo.Update(updated)
	if err != nil {
		return nil, err
	}
	return saved.ToResponse(roundRupiah(saved.MaxAmount - used)), nil
}

// GetByID — detail satu pos anggaran beserta sisa dananya.
func (uc *budgetUseCase) GetByID(id uint64) (*domain.BudgetResponse, error) {
	budget, err := uc.budgetRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	remaining, err := uc.budgetRepo.GetRemainingAmount(id)
	if err != nil {
		return nil, err
	}
	return budget.ToResponse(remaining), nil
}

// List — semua pos anggaran beserta sisa dananya.
// Satu query sisa dana per pos: tidak masalah untuk jumlah pos kelompok yang kecil.
func (uc *budgetUseCase) List() ([]domain.BudgetResponse, error) {
	budgets, err := uc.budgetRepo.FindAll()
	if err != nil {
		return nil, err
	}

	result := make([]domain.BudgetResponse, 0, len(budgets))
	for i := range budgets {
		remaining, err := uc.budgetRepo.GetRemainingAmount(budgets[i].ID)
		if err != nil {
			return nil, err
		}
		result = append(result, *budgets[i].ToResponse(remaining))
	}
	return result, nil
}

// buildBudget memvalidasi input (nominal dan tanggal) lalu menyusun entitas Budget.
func (uc *budgetUseCase) buildBudget(req *domain.BudgetRequest) (*domain.Budget, error) {
	if err := uc.validate.Struct(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return nil, errors.New(buildValidationMessage(ve[0]))
		}
		return nil, errors.New("input tidak valid")
	}

	start, err := time.ParseInLocation(budgetDateLayout, strings.TrimSpace(req.StartDate), time.Local)
	if err != nil {
		return nil, errors.New("format start_date harus YYYY-MM-DD, contoh: 2026-11-01")
	}
	end, err := time.ParseInLocation(budgetDateLayout, strings.TrimSpace(req.EndDate), time.Local)
	if err != nil {
		return nil, errors.New("format end_date harus YYYY-MM-DD, contoh: 2026-11-30")
	}
	if end.Before(start) {
		return nil, domain.ErrBudgetPeriodInvalid
	}

	return &domain.Budget{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		MaxAmount:   roundRupiah(req.MaxAmount),
		StartDate:   start,
		EndDate:     end,
	}, nil
}

// roundRupiah membulatkan ke 2 desimal agar perbandingan nominal tidak
// kena selisih floating point.
func roundRupiah(v float64) float64 {
	return math.Round(v*100) / 100
}