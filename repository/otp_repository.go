package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/transpara/backend/domain"
)

type otpRepository struct {
	db *sql.DB
}

// NewOTPRepository membuat instance otpRepository baru.
// Mengembalikan domain.OTPRepository (interface) — dependency injection.
func NewOTPRepository(db *sql.DB) domain.OTPRepository {
	return &otpRepository{db: db}
}

// Save menyimpan OTP token baru ke database.
// OTPHash sudah dalam bentuk bcrypt sebelum masuk ke sini.
func (r *otpRepository) Save(otp *domain.OTPToken) (*domain.OTPToken, error) {
	query := `
		INSERT INTO otp_tokens (user_id, otp_hash, expires_at)
		VALUES (?, ?, ?)`

	result, err := r.db.Exec(query, otp.UserID, otp.OTPHash, otp.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan OTP: %w", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil ID OTP baru: %w", err)
	}

	otp.ID = uint64(lastID)
	return otp, nil
}

// FindActiveByUserID mencari OTP yang masih aktif:
// - belum dipakai (is_used = 0)
// - belum kadaluarsa (expires_at > NOW())
// - percobaan belum melebihi batas maksimal
func (r *otpRepository) FindActiveByUserID(userID uint64) (*domain.OTPToken, error) {
	query := `
		SELECT id, user_id, otp_hash, expires_at, is_used, attempt_count, created_at
		FROM otp_tokens
		WHERE user_id    = ?
		  AND is_used    = 0
		  AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1`

	otp := &domain.OTPToken{}
	err := r.db.QueryRow(query, userID).Scan(
		&otp.ID,
		&otp.UserID,
		&otp.OTPHash,
		&otp.ExpiresAt,
		&otp.IsUsed,
		&otp.AttemptCount,
		&otp.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("OTP tidak ditemukan atau sudah kadaluarsa")
		}
		return nil, fmt.Errorf("gagal mencari OTP: %w", err)
	}

	return otp, nil
}

// IncrementAttempt menambah hitungan percobaan verifikasi yang salah.
func (r *otpRepository) IncrementAttempt(otpID uint64) error {
	query := `UPDATE otp_tokens SET attempt_count = attempt_count + 1 WHERE id = ?`
	_, err := r.db.Exec(query, otpID)
	if err != nil {
		return fmt.Errorf("gagal memperbarui attempt OTP: %w", err)
	}
	return nil
}

// MarkUsed menandai OTP sebagai sudah digunakan sehingga tidak bisa dipakai lagi.
func (r *otpRepository) MarkUsed(otpID uint64) error {
	query := `UPDATE otp_tokens SET is_used = 1 WHERE id = ?`
	_, err := r.db.Exec(query, otpID)
	if err != nil {
		return fmt.Errorf("gagal menandai OTP sebagai used: %w", err)
	}
	return nil
}

// DeleteByUserID menghapus semua OTP milik user (dipakai saat resend).
func (r *otpRepository) DeleteByUserID(userID uint64) error {
	query := `DELETE FROM otp_tokens WHERE user_id = ?`
	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("gagal menghapus OTP lama: %w", err)
	}
	return nil
}
