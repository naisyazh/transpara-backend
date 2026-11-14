package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/transpara/backend/helper"
)

// RequireAuth adalah middleware JWT yang melindungi endpoint tertentu.
//
// Cara kerja:
// 1. Ambil header "Authorization: Bearer <token>"
// 2. Verifikasi token menggunakan JWTHelper
// 3. Simpan data user ke context Fiber (c.Locals)
// 4. Lanjutkan ke handler jika token valid
// 5. Tolak request dengan 401 jika token tidak ada atau tidak valid
//
// Cara pakai di router:
//
//	protected.Get("/profile", middleware.RequireAuth(jwtHelper), handler.Profile)
func RequireAuth(jwtHelper *helper.JWTHelper) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ambil header Authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Akses ditolak, token tidak ditemukan. Silakan login terlebih dahulu",
			})
		}

		// Format header harus: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Format token tidak valid. Gunakan: Authorization: Bearer <token>",
			})
		}

		tokenString := parts[1]

		// Verifikasi token — cek tanda tangan dan masa berlaku
		claims, err := jwtHelper.VerifyToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Token tidak valid atau sudah kadaluarsa, silakan login kembali",
			})
		}

		// Simpan data user ke context agar handler bisa akses tanpa query DB
		// Gunakan c.Locals("key", value) untuk menyimpan,
		// dan c.Locals("key") untuk mengambil di handler
		c.Locals("userID", claims.UserID)
		c.Locals("userName", claims.Name)
		c.Locals("userEmail", claims.Email)

		// Lanjutkan ke handler berikutnya
		return c.Next()
	}
}
