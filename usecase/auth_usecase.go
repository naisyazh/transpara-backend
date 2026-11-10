package usecase

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/transpara/backend/domain"
	"github.com/transpara/backend/helper"
	"golang.org/x/crypto/bcrypt"
)

// AuthUseCase adalah kontrak (interface) untuk semua use case autentikasi.
type AuthUseCase interface {
	Register(req *domain.RegisterRequest) (*RegisterResult, error)
	VerifyOTP(req *domain.VerifyOTPRequest) (*domain.UserResponse, error)
	ResendOTP(req *domain.ResendOTPRequest) (*ResendResult, error)
	Login(req *domain.LoginRequest) (*domain.LoginResponse, error)
}

// RegisterResult adalah hasil dari Register — berisi data user dan OTP simulasi.
type RegisterResult struct {
	User    *domain.UserResponse `json:"user"`
	OTPCode string               `json:"otp_code"`
	Message string               `json:"message"`
}

// ResendResult adalah hasil dari ResendOTP.
type ResendResult struct {
	OTPCode string `json:"otp_code"`
	Message string `json:"message"`
}

// authUseCase adalah implementasi konkret dari AuthUseCase.
type authUseCase struct {
	userRepo  domain.UserRepository
	otpRepo   domain.OTPRepository
	jwtHelper *helper.JWTHelper
	validate  *validator.Validate
}

// NewAuthUseCase membuat instance authUseCase baru (dependency injection).
func NewAuthUseCase(
	userRepo domain.UserRepository,
	otpRepo domain.OTPRepository,
	jwtHelper *helper.JWTHelper,
) AuthUseCase {
	return &authUseCase{
		userRepo:  userRepo,
		otpRepo:   otpRepo,
		jwtHelper: jwtHelper,
		validate:  validator.New(),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Register — mendaftarkan user baru dan membuat OTP verifikasi
// ─────────────────────────────────────────────────────────────────────────────
func (uc *authUseCase) Register(req *domain.RegisterRequest) (*RegisterResult, error) {
	// 1. Validasi input
	if err := uc.validate.Struct(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return nil, errors.New(buildValidationMessage(ve[0]))
		}
		return nil, errors.New("input tidak valid")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	// 2. Cek email belum terdaftar
	existing, _ := uc.userRepo.FindByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("email sudah terdaftar, gunakan email lain")
	}

	// 3. Hash password dengan bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("gagal memproses password, coba lagi")
	}

	// 4. Simpan user baru (is_verified = 0 secara default di database)
	newUser := &domain.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}
	savedUser, err := uc.userRepo.Save(newUser)
	if err != nil {
		return nil, err
	}

	// 5. Generate OTP 6 digit kriptografis aman
	otpCode, err := generateOTP()
	if err != nil {
		return nil, errors.New("gagal membuat OTP, coba lagi")
	}

	// 6. Hash OTP dan simpan ke database
	if err := uc.saveOTP(savedUser.ID, otpCode); err != nil {
		return nil, err
	}

	return &RegisterResult{
		User:    savedUser.ToResponse(),
		OTPCode: otpCode,
		Message: "[SIMULASI] OTP berhasil dibuat. Gunakan otp_code ini untuk verifikasi.",
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// VerifyOTP — memverifikasi OTP yang dimasukkan user
// ─────────────────────────────────────────────────────────────────────────────
func (uc *authUseCase) VerifyOTP(req *domain.VerifyOTPRequest) (*domain.UserResponse, error) {
	// 1. Validasi input
	if err := uc.validate.Struct(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return nil, errors.New(buildValidationMessage(ve[0]))
		}
		return nil, errors.New("input tidak valid")
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	// 2. Cari user
	user, err := uc.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("email tidak terdaftar")
	}

	// 3. Cek belum terverifikasi
	if user.IsVerified {
		return nil, errors.New("akun sudah terverifikasi, silakan login")
	}

	// 4. Ambil OTP aktif
	otpToken, err := uc.otpRepo.FindActiveByUserID(user.ID)
	if err != nil {
		return nil, errors.New("OTP tidak ditemukan atau sudah kadaluarsa, minta OTP baru")
	}

	// 5. Cek batas percobaan
	if otpToken.AttemptCount >= domain.OTPMaxAttempt {
		return nil, fmt.Errorf(
			"terlalu banyak percobaan salah (%d/%d), minta OTP baru",
			otpToken.AttemptCount, domain.OTPMaxAttempt,
		)
	}

	// 6. Bandingkan OTP yang dikirim dengan hash di database
	if err := bcrypt.CompareHashAndPassword([]byte(otpToken.OTPHash), []byte(req.OTP)); err != nil {
		_ = uc.otpRepo.IncrementAttempt(otpToken.ID)
		remaining := domain.OTPMaxAttempt - (otpToken.AttemptCount + 1)
		return nil, fmt.Errorf("OTP salah, sisa percobaan: %d", remaining)
	}

	// 7. OTP benar — tandai used dan update status user menjadi verified
	if err := uc.otpRepo.MarkUsed(otpToken.ID); err != nil {
		return nil, errors.New("gagal memproses verifikasi, coba lagi")
	}
	if err := uc.userRepo.UpdateVerified(user.ID); err != nil {
		return nil, errors.New("gagal memperbarui status akun, coba lagi")
	}

	updatedUser, err := uc.userRepo.FindByID(user.ID)
	if err != nil {
		return nil, errors.New("verifikasi berhasil namun gagal mengambil data user")
	}

	return updatedUser.ToResponse(), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ResendOTP — kirim ulang OTP baru
// ─────────────────────────────────────────────────────────────────────────────
func (uc *authUseCase) ResendOTP(req *domain.ResendOTPRequest) (*ResendResult, error) {
	if err := uc.validate.Struct(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return nil, errors.New(buildValidationMessage(ve[0]))
		}
		return nil, errors.New("input tidak valid")
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, err := uc.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("email tidak terdaftar")
	}

	if user.IsVerified {
		return nil, errors.New("akun sudah terverifikasi, silakan login")
	}

	// Hapus semua OTP lama milik user ini
	if err := uc.otpRepo.DeleteByUserID(user.ID); err != nil {
		return nil, errors.New("gagal menghapus OTP lama, coba lagi")
	}

	// Generate dan simpan OTP baru
	otpCode, err := generateOTP()
	if err != nil {
		return nil, errors.New("gagal membuat OTP baru, coba lagi")
	}
	if err := uc.saveOTP(user.ID, otpCode); err != nil {
		return nil, err
	}

	return &ResendResult{
		OTPCode: otpCode,
		Message: "[SIMULASI] OTP baru berhasil dibuat. Gunakan otp_code ini untuk verifikasi.",
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Login — autentikasi user dan kembalikan JWT token asli
// ─────────────────────────────────────────────────────────────────────────────
func (uc *authUseCase) Login(req *domain.LoginRequest) (*domain.LoginResponse, error) {
	// 1. Validasi input
	if err := uc.validate.Struct(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return nil, errors.New(buildValidationMessage(ve[0]))
		}
		return nil, errors.New("input tidak valid")
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	// 2. Cari user berdasarkan email
	user, err := uc.userRepo.FindByEmail(req.Email)
	if err != nil {
		// Pesan generik — tidak bocorkan apakah email terdaftar atau tidak
		return nil, errors.New("email atau password salah")
	}

	// 3. Cek akun sudah diverifikasi
	if !user.IsVerified {
		return nil, errors.New("akun belum diverifikasi, periksa OTP Anda")
	}

	// 4. Bandingkan password dengan hash di database
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("email atau password salah")
	}

	// 5. Generate JWT access token
	token, err := uc.jwtHelper.GenerateToken(user.ID, user.Name, user.Email)
	if err != nil {
		return nil, errors.New("gagal membuat token, coba lagi")
	}

	return &domain.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		User:        *user.ToResponse(),
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helper functions
// ─────────────────────────────────────────────────────────────────────────────

// generateOTP membuat kode OTP 6 digit menggunakan crypto/rand (kriptografis aman).
func generateOTP() (string, error) {
	const digits = "0123456789"
	otp := make([]byte, 6)
	for i := range otp {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		otp[i] = digits[n.Int64()]
	}
	return string(otp), nil
}

// saveOTP meng-hash OTP dan menyimpannya ke database.
func (uc *authUseCase) saveOTP(userID uint64, otpCode string) error {
	otpHash, err := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("gagal memproses OTP, coba lagi")
	}

	otpToken := &domain.OTPToken{
		UserID:    userID,
		OTPHash:   string(otpHash),
		ExpiresAt: time.Now().Add(domain.OTPExpiredMinute),
	}

	_, err = uc.otpRepo.Save(otpToken)
	return err
}

// buildValidationMessage mengonversi error validasi ke pesan Bahasa Indonesia.
func buildValidationMessage(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	switch fe.Tag() {
	case "required":
		return "Field " + field + " wajib diisi"
	case "email":
		return "Format email tidak valid"
	case "min":
		if field == "password" {
			return "Password minimal " + fe.Param() + " karakter"
		}
		return "Field " + field + " minimal " + fe.Param() + " karakter"
	case "max":
		return "Field " + field + " maksimal " + fe.Param() + " karakter"
	case "len":
		return "Field " + field + " harus tepat " + fe.Param() + " karakter"
	default:
		return "Field " + field + " tidak valid"
	}
}
