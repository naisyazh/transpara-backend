package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/transpara/backend/app/model"
)

type transactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository membuat instance transactionRepository baru.
func NewTransactionRepository(db *sql.DB) model.TransactionRepository {
	return &transactionRepository{db: db}
}

// transactionSelect mengambil kolom transaksi plus nama kategori (JOIN).
const transactionSelect = `
	SELECT t.id, t.category_id, c.name, t.created_by, t.expense_request_id,
	       t.type, t.amount, t.description, t.transaction_date,
	       t.created_at, t.updated_at
	FROM transactions t
	JOIN categories c ON c.id = t.category_id`

const dateLayout = "2006-01-02"

func scanTransaction(s rowScanner) (*model.Transaction, error) {
	t := &model.Transaction{}
	var expenseRequestID sql.NullInt64 // NULL untuk pemasukan

	err := s.Scan(
		&t.ID, &t.CategoryID, &t.CategoryName, &t.CreatedBy, &expenseRequestID,
		&t.Type, &t.Amount, &t.Description, &t.TransactionDate,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if expenseRequestID.Valid {
		id := uint64(expenseRequestID.Int64)
		t.ExpenseRequestID = &id
	}
	return t, nil
}

// Create menyimpan transaksi baru.
//
// Pencairan ganda dijaga UNIQUE KEY uq_tx_expense_request di database:
// jika dua request mencairkan pengajuan yang sama bersamaan, hanya satu
// yang lolos dan satunya menerima ErrAlreadyDisbursed. Karena itu tidak
// perlu transaksi atau penguncian baris tambahan di sini.
func (r *transactionRepository) Create(t *model.Transaction) (*model.Transaction, error) {
	result, err := r.db.Exec(`
		INSERT INTO transactions
			(category_id, created_by, expense_request_id, type, amount, description, transaction_date)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		t.CategoryID, t.CreatedBy, t.ExpenseRequestID, t.Type, t.Amount,
		t.Description, t.TransactionDate.Format(dateLayout),
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, model.ErrAlreadyDisbursed
		}
		return nil, fmt.Errorf("gagal menyimpan transaksi: %w", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil ID transaksi baru: %w", err)
	}

	return r.FindByID(uint64(lastID))
}

// FindByID mencari transaksi berdasarkan ID.
func (r *transactionRepository) FindByID(id uint64) (*model.Transaction, error) {
	t, err := scanTransaction(r.db.QueryRow(transactionSelect+` WHERE t.id = ? LIMIT 1`, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrTransactionNotFound
		}
		return nil, fmt.Errorf("gagal mencari transaksi: %w", err)
	}
	return t, nil
}

// FindAll mengembalikan daftar transaksi, terbaru di atas.
// Setiap field filter yang terisi menambah satu kondisi WHERE.
func (r *transactionRepository) FindAll(filter model.TransactionFilter) ([]model.Transaction, error) {
	query := transactionSelect
	conditions := []string{}
	args := []any{}

	if filter.Type != "" {
		conditions = append(conditions, `t.type = ?`)
		args = append(args, filter.Type)
	}
	if filter.CategoryID != 0 {
		conditions = append(conditions, `t.category_id = ?`)
		args = append(args, filter.CategoryID)
	}
	if filter.From != nil {
		conditions = append(conditions, `t.transaction_date >= ?`)
		args = append(args, filter.From.Format(dateLayout))
	}
	if filter.To != nil {
		conditions = append(conditions, `t.transaction_date <= ?`)
		args = append(args, filter.To.Format(dateLayout))
	}

	for i, cond := range conditions {
		if i == 0 {
			query += ` WHERE ` + cond
		} else {
			query += ` AND ` + cond
		}
	}
	query += ` ORDER BY t.transaction_date DESC, t.id DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar transaksi: %w", err)
	}
	defer rows.Close()

	list := make([]model.Transaction, 0) // supaya JSON jadi [] bukan null
	for rows.Next() {
		t, err := scanTransaction(rows)
		if err != nil {
			return nil, fmt.Errorf("gagal membaca transaksi: %w", err)
		}
		list = append(list, *t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gagal membaca daftar transaksi: %w", err)
	}
	return list, nil
}

// GetSummary menghitung total pemasukan, total pengeluaran, dan saldo
// langsung di database supaya tidak ada selisih pembulatan float di Go.
func (r *transactionRepository) GetSummary() (*model.BalanceSummary, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income'  THEN amount END), 0),
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount END), 0),
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE -amount END), 0)
		FROM transactions`

	s := &model.BalanceSummary{}
	if err := r.db.QueryRow(query).Scan(&s.TotalIncome, &s.TotalExpense, &s.Balance); err != nil {
		return nil, fmt.Errorf("gagal menghitung saldo: %w", err)
	}
	return s, nil
}

// isDuplicateKeyError memeriksa error duplikat UNIQUE KEY MySQL/MariaDB (kode 1062).
func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}