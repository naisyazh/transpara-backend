package service

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/transpara/backend/app/model"
)

// CategoryService adalah kontrak use case untuk kategori transaksi.
type CategoryService interface {
	Create(req *model.CreateCategoryRequest) (*model.Category, error)
	// List mengembalikan kategori; txType boleh kosong, income, atau expense.
	List(txType string) ([]model.Category, error)
}

type categoryService struct {
	categoryRepo model.CategoryRepository
	validate     *validator.Validate
}

// NewCategoryService membuat instance categoryService baru (dependency injection).
func NewCategoryService(categoryRepo model.CategoryRepository) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
		validate:     validator.New(),
	}
}

// Create — buat kategori baru (khusus bendahara, dijaga di router).
func (uc *categoryService) Create(req *model.CreateCategoryRequest) (*model.Category, error) {
	if err := uc.validate.Struct(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return nil, errors.New(buildValidationMessage(ve[0]))
		}
		return nil, errors.New("input tidak valid")
	}

	return uc.categoryRepo.Create(&model.Category{
		Name: strings.TrimSpace(req.Name),
		Type: req.Type,
	})
}

// List — daftar kategori, boleh difilter berdasarkan tipe.
func (uc *categoryService) List(txType string) ([]model.Category, error) {
	txType = strings.ToLower(strings.TrimSpace(txType))

	switch txType {
	case "", model.TransactionTypeIncome, model.TransactionTypeExpense:
		// tipe valid
	default:
		return nil, errors.New("tipe tidak valid, gunakan income atau expense")
	}

	return uc.categoryRepo.FindAll(txType)
}