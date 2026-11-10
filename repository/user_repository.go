package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/transpara/backend/domain"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository membuat instance userRepository baru.
// Mengembalikan domain.UserRepository (interface) — dependency injection.
func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// Save menyimpan user baru ke tabel users.
func (r *userRepository) Save(user *domain.User) (*domain.User, error) {
	query := `INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)`

	result, err := r.db.Exec(query, user.Name, user.Email, user.PasswordHash)
	if err != nil {
		if isDuplicateEmailError(err) {
			return nil, errors.New("email sudah terdaftar, gunakan email lain")
		}
		return nil, fmt.Errorf("gagal menyimpan user: %w", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil ID user baru: %w", err)
	}

	return r.FindByID(uint64(lastID))
}

// FindByEmail mencari user berdasarkan alamat email.
func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	query := `
		SELECT id, name, email, password_hash, is_verified, created_at, updated_at
		FROM users
		WHERE email = ?
		LIMIT 1`

	user := &domain.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mencari user: %w", err)
	}

	return user, nil
}

// FindByID mencari user berdasarkan ID.
func (r *userRepository) FindByID(id uint64) (*domain.User, error) {
	query := `
		SELECT id, name, email, password_hash, is_verified, created_at, updated_at
		FROM users
		WHERE id = ?
		LIMIT 1`

	user := &domain.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mencari user: %w", err)
	}

	return user, nil
}

// UpdateVerified menandai akun user sebagai sudah terverifikasi.
func (r *userRepository) UpdateVerified(userID uint64) error {
	query := `UPDATE users SET is_verified = 1 WHERE id = ?`
	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("gagal memperbarui status verifikasi: %w", err)
	}
	return nil
}

// isDuplicateEmailError memeriksa error duplikat UNIQUE KEY MySQL (kode 1062).
func isDuplicateEmailError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}
