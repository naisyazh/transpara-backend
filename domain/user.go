package domain

import "time"

// User adalah representasi entitas user di dalam sistem.
//
// PENTING: PasswordHash tidak boleh pernah dikirim ke client.
// Gunakan UserResponse untuk semua output API.
type User struct {
	ID           uint64    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`          // tidak muncul di JSON output
	IsVerified   bool      `json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserRepository adalah kontrak (interface) operasi database untuk User.
// Use Case hanya mengenal interface ini, bukan implementasi MySQL-nya.
type UserRepository interface {
	Save(user *User) (*User, error)
	FindByEmail(email string) (*User, error)
	FindByID(id uint64) (*User, error)
	// UpdateVerified menandai akun sebagai sudah terverifikasi
	UpdateVerified(userID uint64) error
}

// ── DTO (Data Transfer Object) ────────────────────────────────────────────

// RegisterRequest — body yang dikirim client saat POST /auth/register
type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// VerifyOTPRequest — body yang dikirim client saat POST /auth/verify-otp
type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp"   validate:"required,len=6"`
}

// ResendOTPRequest — body yang dikirim client saat POST /auth/resend-otp
type ResendOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// LoginRequest — body yang dikirim client saat POST /auth/login
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UserResponse — data user yang aman dikirim ke client (tanpa password)
type UserResponse struct {
	ID         uint64    `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
}

// LoginResponse — response setelah login berhasil
type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	User        UserResponse `json:"user"`
}

// ToResponse mengonversi entitas User ke UserResponse yang aman untuk dikirim ke client.
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:         u.ID,
		Name:       u.Name,
		Email:      u.Email,
		IsVerified: u.IsVerified,
		CreatedAt:  u.CreatedAt,
	}
}
