package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/transpara/backend/internal/domain"
)

type expenseRequestRepository struct {
	db *sql.DB
}

// NewExpenseRequestRepository membuat instance expenseRequestRepository baru.
func NewExpenseRequestRepository(db *sql.DB) domain.ExpenseRequestRepository {
	return &expenseRequestRepository{db: db}
}

const expenseColumns = `id, budget_id, requested_by, title, amount, reason, status,
	yes_votes, no_votes, decided_at, created_at, updated_at`

func scanExpense(s rowScanner) (*domain.ExpenseRequest, error) {
	e := &domain.ExpenseRequest{}
	var decidedAt sql.NullTime // NULL selama masih pending

	err := s.Scan(
		&e.ID, &e.BudgetID, &e.RequestedBy, &e.Title, &e.Amount, &e.Reason, &e.Status,
		&e.YesVotes, &e.NoVotes, &decidedAt, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if decidedAt.Valid {
		t := decidedAt.Time
		e.DecidedAt = &t
	}
	return e, nil
}

// Create menyimpan pengajuan baru.
//
// Berjalan dalam transaksi dengan baris pos anggaran dikunci (FOR UPDATE),
// supaya dua pengajuan yang masuk bersamaan tidak sama-sama lolos cek sisa dana.
func (r *expenseRequestRepository) Create(req *domain.ExpenseRequest) (*domain.ExpenseRequest, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback() // aman dipanggil setelah Commit, hasilnya diabaikan

	var lockedID uint64
	err = tx.QueryRow(`SELECT id FROM budgets WHERE id = ? FOR UPDATE`, req.BudgetID).Scan(&lockedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrBudgetNotFound
		}
		return nil, fmt.Errorf("gagal mengunci pos anggaran: %w", err)
	}

	remaining, err := remainingAmount(tx, req.BudgetID)
	if err != nil {
		return nil, err
	}
	if req.Amount > remaining {
		return nil, domain.ErrAmountExceedsRemaining
	}

	result, err := tx.Exec(`
		INSERT INTO expense_requests (budget_id, requested_by, title, amount, reason, status)
		VALUES (?, ?, ?, ?, ?, ?)`,
		req.BudgetID, req.RequestedBy, req.Title, req.Amount, req.Reason, req.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan pengajuan: %w", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil ID pengajuan baru: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("gagal menyimpan transaksi: %w", err)
	}

	return r.FindByID(uint64(lastID))
}

// FindByID mencari pengajuan berdasarkan ID.
func (r *expenseRequestRepository) FindByID(id uint64) (*domain.ExpenseRequest, error) {
	query := `SELECT ` + expenseColumns + ` FROM expense_requests WHERE id = ? LIMIT 1`

	e, err := scanExpense(r.db.QueryRow(query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrExpenseNotFound
		}
		return nil, fmt.Errorf("gagal mencari pengajuan: %w", err)
	}
	return e, nil
}

// FindAll mengembalikan daftar pengajuan, terbaru di atas.
// Jika status tidak kosong, hanya pengajuan dengan status itu yang diambil.
func (r *expenseRequestRepository) FindAll(status string) ([]domain.ExpenseRequest, error) {
	query := `SELECT ` + expenseColumns + ` FROM expense_requests`
	args := []any{}

	if status != "" {
		query += ` WHERE status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC, id DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar pengajuan: %w", err)
	}
	defer rows.Close()

	list := make([]domain.ExpenseRequest, 0) // supaya JSON jadi [] bukan null
	for rows.Next() {
		e, err := scanExpense(rows)
		if err != nil {
			return nil, fmt.Errorf("gagal membaca pengajuan: %w", err)
		}
		list = append(list, *e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gagal membaca daftar pengajuan: %w", err)
	}
	return list, nil
}

// HasVoted mengecek apakah user sudah memberi suara pada pengajuan ini.
func (r *expenseRequestRepository) HasVoted(requestID, userID uint64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM expense_votes
			WHERE expense_request_id = ? AND user_id = ?
		)`

	var voted bool
	if err := r.db.QueryRow(query, requestID, userID).Scan(&voted); err != nil {
		return false, fmt.Errorf("gagal mengecek status vote: %w", err)
	}
	return voted, nil
}

// CountVotingMembers menghitung total anggota yang berhak vote.
// Query ke tabel users ditaruh di sini supaya UserRepository milik
// Developer 1 tidak perlu diubah.
func (r *expenseRequestRepository) CountVotingMembers() (int, error) {
	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE is_verified = 1`).Scan(&total); err != nil {
		return 0, fmt.Errorf("gagal menghitung jumlah anggota: %w", err)
	}
	return total, nil
}

// CastVote menyimpan suara dalam satu transaksi.
//
// Urutan di dalam transaksi:
//  1. Kunci baris pengajuan (SELECT ... FOR UPDATE). Voter lain yang masuk
//     bersamaan akan menunggu di sini, jadi hitungan suara tidak tertimpa.
//  2. Hitung yes/no setelah suara ini.
//  3. Minta usecase menentukan status baru lewat decide (boleh menolak).
//  4. Insert suara. UNIQUE (expense_request_id, user_id) menjadi penjaga
//     terakhir kalau user yang sama mengirim dua kali bersamaan.
//  5. Update yes_votes, no_votes, dan status (+ decided_at jika sudah diputuskan).
func (r *expenseRequestRepository) CastVote(
	requestID, userID uint64, vote string, decide domain.DecideFunc,
) (*domain.ExpenseRequest, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback()

	// 1. Kunci baris pengajuan
	current, err := scanExpense(tx.QueryRow(
		`SELECT `+expenseColumns+` FROM expense_requests WHERE id = ? FOR UPDATE`, requestID,
	))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrExpenseNotFound
		}
		return nil, fmt.Errorf("gagal mengunci pengajuan: %w", err)
	}

	// 2. Hitungan suara setelah suara ini
	yes, no := current.YesVotes, current.NoVotes
	if vote == domain.VoteYes {
		yes++
	} else {
		no++
	}

	// 3. Usecase menentukan status baru (atau menolak, mis. status sudah final)
	newStatus, err := decide(current, yes, no)
	if err != nil {
		return nil, err
	}

	// 4. Simpan suara
	_, err = tx.Exec(
		`INSERT INTO expense_votes (expense_request_id, user_id, vote) VALUES (?, ?, ?)`,
		requestID, userID, vote,
	)
	if err != nil {
		if isDuplicateVoteError(err) {
			return nil, domain.ErrAlreadyVoted
		}
		return nil, fmt.Errorf("gagal menyimpan suara: %w", err)
	}

	// 5. Update hitungan dan status
	if newStatus == domain.ExpenseStatusPending {
		_, err = tx.Exec(
			`UPDATE expense_requests SET yes_votes = ?, no_votes = ? WHERE id = ?`,
			yes, no, requestID,
		)
	} else {
		_, err = tx.Exec(
			`UPDATE expense_requests
			 SET yes_votes = ?, no_votes = ?, status = ?, decided_at = NOW()
			 WHERE id = ?`,
			yes, no, newStatus, requestID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui hasil voting: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("gagal menyimpan transaksi voting: %w", err)
	}

	return r.FindByID(requestID)
}

// isDuplicateVoteError memeriksa error duplikat UNIQUE KEY MySQL (kode 1062).
func isDuplicateVoteError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}