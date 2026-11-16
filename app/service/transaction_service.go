package service

import (
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/transpara/backend/app/model"
)

const dateLayout = "2006-01-02"

// TransactionService adalah kontrak use case untuk transaksi keuangan.
type TransactionService interface {
	// Create mencatat pemasukan atau pengeluaran (khusus bendahara, dijaga di router).
	Create(userID uint64, req *model.CreateTransactionRequest) (*model.Transaction, error)
	GetByID(id uint64) (*model.Transaction, error)
	// List mengembalikan transaksi. Semua parameter boleh kosong (tanpa filter);
	// from dan to berformat YYYY-MM-DD.
	List(txType string, categoryID uint64, from, to string) ([]model.Transaction, error)
	// GetSummary mengembalikan total pemasukan, pengeluaran, dan saldo kas.
	GetSummary() (*model.BalanceSummary, error)
}

type transactionService struct {
	transactionRepo model.TransactionRepository
	categoryRepo    model.CategoryRepository
	expenseService  ExpenseService
	validate        *validator.Validate
}

// NewTransactionService membuat instance transactionService baru (dependency injection).
// expenseService dipakai untuk EnsureApproved sebelum dana dicairkan.
func NewTransactionService(
	transactionRepo model.TransactionRepository,
	categoryRepo model.CategoryRepository,
	expenseService ExpenseService,
) TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
		expenseService:  expenseService,
		validate:        validator.New(),
	}
}

// Create — catat transaksi.
//
//   - income : nominal dari input, tidak boleh terhubung ke pengajuan.
//   - expense: wajib terhubung ke pengajuan berstatus approved, dan nominalnya
//     diambil dari pengajuan (bukan dari input) supaya dana yang cair persis
//     sama dengan yang disetujui anggota.
func (uc *transactionService) Create(userID uint64, req *model.CreateTransactionRequest) (*model.Transaction, error) {
	// 1. Validasi input
	if err := uc.validate.Struct(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return nil, errors.New(buildValidationMessage(ve[0]))
		}
		return nil, errors.New("input tidak valid")
	}

	// 2. Tanggal transaksi
	date, err := time.ParseInLocation(dateLayout, strings.TrimSpace(req.TransactionDate), time.Local)
	if err != nil {
		return nil, model.ErrInvalidDate
	}

	// 3. Kategori harus ada dan tipenya sama dengan tipe transaksi
	category, err := uc.categoryRepo.FindByID(req.CategoryID)
	if err != nil {
		return nil, err
	}
	if category.Type != req.Type {
		return nil, model.ErrCategoryTypeMismatch
	}

	// 4. Aturan per tipe
	var amount float64
	var expenseRequestID *uint64

	switch req.Type {
	case model.TransactionTypeIncome:
		if req.ExpenseRequestID != nil {
			return nil, model.ErrIncomeHasExpenseRequest
		}
		if req.Amount <= 0 {
			return nil, model.ErrAmountRequired
		}
		amount = roundRupiah(req.Amount)

	case model.TransactionTypeExpense:
		if req.ExpenseRequestID == nil || *req.ExpenseRequestID == 0 {
			return nil, model.ErrExpenseRequestRequired
		}
		// Hanya pengajuan approved yang lolos. pending/rejected -> ErrRequestNotApproved,
		// id tidak ada -> ErrExpenseNotFound.
		approved, err := uc.expenseService.EnsureApproved(*req.ExpenseRequestID)
		if err != nil {
			return nil, err
		}
		amount = approved.Amount
		id := approved.ID
		expenseRequestID = &id
	}

	// 5. Simpan. Pencairan ganda ditolak UNIQUE KEY di database (ErrAlreadyDisbursed).
	return uc.transactionRepo.Create(&model.Transaction{
		CategoryID:       category.ID,
		CreatedBy:        userID,
		ExpenseRequestID: expenseRequestID,
		Type:             req.Type,
		Amount:           amount,
		Description:      strings.TrimSpace(req.Description),
		TransactionDate:  date,
	})
}

// GetByID — detail satu transaksi.
func (uc *transactionService) GetByID(id uint64) (*model.Transaction, error) {
	return uc.transactionRepo.FindByID(id)
}

// List — daftar transaksi dengan filter opsional.
func (uc *transactionService) List(txType string, categoryID uint64, from, to string) ([]model.Transaction, error) {
	filter := model.TransactionFilter{CategoryID: categoryID}

	txType = strings.ToLower(strings.TrimSpace(txType))
	switch txType {
	case "", model.TransactionTypeIncome, model.TransactionTypeExpense:
		filter.Type = txType
	default:
		return nil, errors.New("tipe tidak valid, gunakan income atau expense")
	}

	if from = strings.TrimSpace(from); from != "" {
		t, err := time.ParseInLocation(dateLayout, from, time.Local)
		if err != nil {
			return nil, model.ErrInvalidDate
		}
		filter.From = &t
	}
	if to = strings.TrimSpace(to); to != "" {
		t, err := time.ParseInLocation(dateLayout, to, time.Local)
		if err != nil {
			return nil, model.ErrInvalidDate
		}
		filter.To = &t
	}
	if filter.From != nil && filter.To != nil && filter.From.After(*filter.To) {
		return nil, errors.New("tanggal awal tidak boleh setelah tanggal akhir")
	}

	return uc.transactionRepo.FindAll(filter)
}

// GetSummary — saldo kas = total pemasukan dikurangi total pengeluaran.
func (uc *transactionService) GetSummary() (*model.BalanceSummary, error) {
	return uc.transactionRepo.GetSummary()
}