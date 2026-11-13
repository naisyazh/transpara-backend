package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/transpara/backend/domain"
)

const sqlDateLayout = "2006-01-02"

// rowScanner dipakai bersama oleh *sql.Row dan *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

// queryRower dipakai bersama oleh *sql.DB dan *sql.Tx,
// supaya satu fungsi query bisa jalan di dalam maupun di luar transaksi.
type queryRower interface {
	QueryRow(query string, args ...any) *sql.Row
}

type budgetRepository struct {
	db *sql.DB
}

// NewBudgetRepository membuat instance budgetRepository baru (dependency injection).
func NewBudgetRepository(db *sql.DB) domain.BudgetRepository {
	return &budgetRepository{db: db}
}

const budgetColumns = `id, name, description, max_amount, start_date, end_date, created_by, created_at, updated_at`

func scanBudget(s rowScanner) (*domain.Budget, error) {
	b := &domain.Budget{}
	var description sql.NullString // kolom description boleh NULL

	err := s.Scan(
		&b.ID, &b.Name, &description, &b.MaxAmount,
		&b.StartDate, &b.EndDate, &b.CreatedBy, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	b.Description = description.String
	return b, nil
}

func toNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// Create menyimpan pos anggaran baru.
func (r *budgetRepository) Create(b *domain.Budget) (*domain.Budget, error) {
	query := `
		INSERT INTO budgets (name, description, max_amount, start_date, end_date, created_by)
		VALUES (?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		b.Name, toNullString(b.Description), b.MaxAmount,
		b.StartDate.Format(sqlDateLayout), b.EndDate.Format(sqlDateLayout), b.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan pos anggaran: %w", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil ID pos anggaran baru: %w", err)
	}

	return r.FindByID(uint64(lastID))
}

// Update mengubah data pos anggaran. Keberadaan pos dicek oleh usecase.
func (r *budgetRepository) Update(b *domain.Budget) (*domain.Budget, error) {
	query := `
		UPDATE budgets
		SET name = ?, description = ?, max_amount = ?, start_date = ?, end_date = ?
		WHERE id = ?`

	_, err := r.db.Exec(query,
		b.Name, toNullString(b.Description), b.MaxAmount,
		b.StartDate.Format(sqlDateLayout), b.EndDate.Format(sqlDateLayout), b.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui pos anggaran: %w", err)
	}

	return r.FindByID(b.ID)
}

// FindByID mencari pos anggaran berdasarkan ID.
func (r *budgetRepository) FindByID(id uint64) (*domain.Budget, error) {
	query := `SELECT ` + budgetColumns + ` FROM budgets WHERE id = ? LIMIT 1`

	b, err := scanBudget(r.db.QueryRow(query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrBudgetNotFound
		}
		return nil, fmt.Errorf("gagal mencari pos anggaran: %w", err)
	}
	return b, nil
}

// FindAll mengembalikan semua pos anggaran, periode terbaru di atas.
func (r *budgetRepository) FindAll() ([]domain.Budget, error) {
	query := `SELECT ` + budgetColumns + ` FROM budgets ORDER BY start_date DESC, id DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar pos anggaran: %w", err)
	}
	defer rows.Close()

	budgets := make([]domain.Budget, 0) // supaya JSON jadi [] bukan null
	for rows.Next() {
		b, err := scanBudget(rows)
		if err != nil {
			return nil, fmt.Errorf("gagal membaca pos anggaran: %w", err)
		}
		budgets = append(budgets, *b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gagal membaca daftar pos anggaran: %w", err)
	}
	return budgets, nil
}

// GetRemainingAmount menghitung sisa dana pos.
func (r *budgetRepository) GetRemainingAmount(budgetID uint64) (float64, error) {
	return remainingAmount(r.db, budgetID)
}

// remainingAmount = max_amount - total pengajuan pending dan approved.
// Menerima queryRower agar bisa dipakai di dalam transaksi (lihat ExpenseRequest Create).
func remainingAmount(q queryRower, budgetID uint64) (float64, error) {
	query := `
		SELECT b.max_amount - COALESCE((
			SELECT SUM(er.amount)
			FROM expense_requests er
			WHERE er.budget_id = b.id
			  AND er.status IN ('pending', 'approved')
		), 0)
		FROM budgets b
		WHERE b.id = ?`

	var remaining float64
	if err := q.QueryRow(query, budgetID).Scan(&remaining); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrBudgetNotFound
		}
		return 0, fmt.Errorf("gagal menghitung sisa dana: %w", err)
	}
	return remaining, nil
}