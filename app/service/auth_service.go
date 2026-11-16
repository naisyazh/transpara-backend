package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/transpara/backend/app/model"
	"github.com/transpara/backend/helper"
	"github.com/transpara/backend/helper/email"
	"golang.org/x/crypto/bcrypt"
)

// AuthService adalah kontrak (interface) untuk semua use case autentikasi.
type AuthService interface {
	Register(req *model.RegisterRequest) (*RegisterResult, error)
	VerifyOTP(req *model.VerifyOTPRequest) (*model.UserResponse, error)
	ResendOTP(req *model.ResendOTPRequest) (*ResendResult, error)
	Login(req *model.LoginRequest) (*model.LoginResponse, error)
}

// RegisterResult adalah hasil dari Register.
// OTPCode dikosongkan setelah email dikirim — tidak dikirim ke client.
type RegisterResult struct {
	User    *model.UserResponse `json:"user"`
	Message string              `json:"message"`
}

// ResendResult adalah hasil dari ResendOTP.
type ResendResult struct {
	Message string `json:"message"`
}

// authService adalah implementasi konkret dari AuthService.
type authService struct {
	userRepo     model.UserRepository
	otpRepo      model.OTPRepository
	jwtHelper    *helper.JWTHelper
	emailService *email.EmailService
	validate     *validator.Validate
}

// NewAuthService membuat instance authService baru (dependency injection).
func NewAuthService(
	userRepo model.UserRepository,
	otpRepo model.OTPRepository,
	jwtHelper *helper.JWTHelper,
	emailService *email.EmailService,
) AuthService {
	return &authService{
		userRepo:     userRepo,
		otpRepo:      otpRepo,
		jwtHelper:    jwtHelper,
		emailService: emailService,
		validate:     validator.New(),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Register — daftarkan user baru, buat OTP, kirim ke email
// ─────────────────────────────────────────────────────────────────────────────
func (uc *authService) Register(req *model.RegisterRequest) (*RegisterResult, error) {
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

	// 3. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("gagal memproses password, coba lagi")
	}

	// 4. Simpan user baru (is_verified = 0 secara default)
	newUser := &model.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}
	savedUser, err := uc.userRepo.Save(newUser)
	if err != nil {
		return nil, err
	}

	// 5. Generate OTP 6 digit
	otpCode, err := generateOTP()
	if err != nil {
		return nil, errors.New("gagal membuat OTP, coba lagi")
	}

	// 6. Hash OTP dan simpan ke database
	if err := uc.saveOTP(savedUser.ID, otpCode); err != nil {
		return nil, err
	}

	// 7. Kirim OTP ke email user — OTP TIDAK dikembalikan ke response API
	if err := uc.emailService.SendOTP(savedUser.Email, savedUser.Name, otpCode); err != nil {
		// Jika gagal kirim email, hapus user yang baru dibuat agar bisa register ulang
		return nil, fmt.Errorf("gagal mengirim email OTP ke %s, coba lagi", savedUser.Email)
	}

	return &RegisterResult{
		User:    savedUser.ToResponse(),
		Message: fmt.Sprintf("Registrasi berhasil. Kode OTP telah dikirim ke %s", savedUser.Email),
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// VerifyOTP — verifikasi OTP yang dimasukkan user
// ─────────────────────────────────────────────────────────────────────────────
func (uc *authService) VerifyOTP(req *model.VerifyOTPRequest) (*model.UserResponse, error) {
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

	otpToken, err := uc.otpRepo.FindActiveByUserID(user.ID)
	if err != nil {
		return nil, errors.New("OTP tidak ditemukan atau sudah kadaluarsa, minta OTP baru")
	}

	if otpToken.AttemptCount >= model.OTPMaxAttempt {
		return nil, fmt.Errorf(
			"terlalu banyak percobaan salah (%d/%d), minta OTP baru",
			otpToken.AttemptCount, model.OTPMaxAttempt,
		)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(otpToken.OTPHash), []byte(req.OTP)); err != nil {
		_ = uc.otpRepo.IncrementAttempt(otpToken.ID)
		remaining := model.OTPMaxAttempt - (otpToken.AttemptCount + 1)
		return nil, fmt.Errorf("OTP salah, sisa percobaan: %d", remaining)
	}

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
// ResendOTP — kirim ulang OTP baru ke email user
// ─────────────────────────────────────────────────────────────────────────────
func (uc *authService) ResendOTP(req *model.ResendOTPRequest) (*ResendResult, error) {
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

	// Hapus semua OTP lama
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

	// Kirim OTP ke email — OTP tidak dikembalikan ke response
	if err := uc.emailService.SendOTP(user.Email, user.Name, otpCode); err != nil {
		return nil, fmt.Errorf("gagal mengirim email OTP ke %s, coba lagi", user.Email)
	}

	return &ResendResult{
		Message: fmt.Sprintf("OTP baru telah dikirim ke %s", user.Email),
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Login — autentikasi user dan kembalikan JWT token
// ─────────────────────────────────────────────────────────────────────────────
func (uc *authService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
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
		return nil, errors.New("email atau password salah")
	}

	if !user.IsVerified {
		return nil, errors.New("akun belum diverifikasi, periksa email Anda untuk kode OTP")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("email atau password salah")
	}

	token, err := uc.jwtHelper.GenerateToken(user.ID, user.Name, user.Email)
	if err != nil {
		return nil, errors.New("gagal membuat token, coba lagi")
	}

	return &model.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		User:        *user.ToResponse(),
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helper functions
// ─────────────────────────────────────────────────────────────────────────────

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

func (uc *authService) saveOTP(userID uint64, otpCode string) error {
	otpHash, err := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("gagal memproses OTP, coba lagi")
	}

	otpToken := &model.OTPToken{
		UserID:    userID,
		OTPHash:   string(otpHash),
		ExpiresAt: time.Now().Add(model.OTPExpiredMinute),
	}

	_, err = uc.otpRepo.Save(otpToken)
	return err
}

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
