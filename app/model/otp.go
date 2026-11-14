package model

import "time"

// OTPToken adalah representasi baris di tabel otp_tokens.
// Field OTPHash menyimpan hasil bcrypt dari OTP asli — OTP asli tidak disimpan.
type OTPToken struct {
	ID           uint64    `json:"id"`
	UserID       uint64    `json:"user_id"`
	OTPHash      string    `json:"-"` // hash bcrypt, tidak dikirim ke client
	ExpiresAt    time.Time `json:"expires_at"`
	IsUsed       bool      `json:"is_used"`
	AttemptCount int       `json:"attempt_count"`
	CreatedAt    time.Time `json:"created_at"`
}

// OTPRepository adalah kontrak (interface) operasi database untuk OTPToken.
type OTPRepository interface {
	// Save menyimpan OTP token baru ke database
	Save(otp *OTPToken) (*OTPToken, error)

	// FindActiveByUserID mencari OTP yang masih aktif (belum dipakai & belum kadaluarsa)
	// untuk user tertentu
	FindActiveByUserID(userID uint64) (*OTPToken, error)

	// IncrementAttempt menambah hitungan percobaan yang salah
	IncrementAttempt(otpID uint64) error

	// MarkUsed menandai OTP sebagai sudah digunakan
	MarkUsed(otpID uint64) error

	// DeleteByUserID menghapus semua OTP lama milik user (untuk resend)
	DeleteByUserID(userID uint64) error
}

// Konstanta batas percobaan dan masa berlaku OTP
const (
	OTPMaxAttempt    = 5               // maksimal 5x percobaan salah
	OTPExpiredMinute = 5 * time.Minute // OTP berlaku 5 menit
)
