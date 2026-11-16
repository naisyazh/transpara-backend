package helper

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims adalah isi payload yang disimpan di dalam token.
// Kita simpan UserID dan Email agar middleware bisa tahu siapa yang login
// tanpa perlu query ke database setiap request.
type JWTClaims struct {
	UserID uint64 `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWTHelper mengelola pembuatan dan verifikasi JWT token.
type JWTHelper struct {
	secretKey   []byte
	issuer      string
	expiredHour time.Duration
}

// NewJWTHelper membuat instance JWTHelper baru.
// secretKey dan issuer dibaca dari konfigurasi .env.
func NewJWTHelper(secretKey, issuer string, expiredHour int) *JWTHelper {
	return &JWTHelper{
		secretKey:   []byte(secretKey),
		issuer:      issuer,
		expiredHour: time.Duration(expiredHour) * time.Hour,
	}
}

// GenerateToken membuat JWT access token baru untuk user yang berhasil login.
// Token berisi UserID, Name, dan Email — tidak menyimpan password sama sekali.
func (j *JWTHelper) GenerateToken(userID uint64, name, email string) (string, error) {
	now := time.Now()

	claims := JWTClaims{
		UserID: userID,
		Name:   name,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expiredHour)),
		},
	}

	// Buat token dengan algoritma HS256 (HMAC SHA-256)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Tanda tangani token dengan secret key
	signedToken, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", errors.New("gagal membuat token")
	}

	return signedToken, nil
}

// VerifyToken memverifikasi token dan mengembalikan claims-nya.
// Dipanggil oleh middleware setiap ada request ke endpoint yang dilindungi.
func (j *JWTHelper) VerifyToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(t *jwt.Token) (interface{}, error) {
			// Pastikan algoritma yang dipakai adalah HMAC (HS256)
			// Ini mencegah serangan algorithm confusion
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("algoritma token tidak valid")
			}
			return j.secretKey, nil
		},
	)

	if err != nil {
		return nil, errors.New("token tidak valid atau sudah kadaluarsa")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token tidak valid")
	}

	return claims, nil
}
