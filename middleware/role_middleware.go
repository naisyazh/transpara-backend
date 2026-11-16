package middleware

import (
	"database/sql"
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
)

// RequireRole membatasi akses endpoint hanya untuk role tertentu.
//
// PENTING: harus dipasang SETELAH RequireAuth, karena userID diambil dari
// c.Locals("userID") yang diisi oleh RequireAuth.
//
// Role dibaca dari tabel users pada setiap request, bukan dari token JWT.
// Dengan begitu perubahan role lewat database langsung berlaku tanpa
// user perlu login ulang.
//
// Cara pakai di router:
//
//	budgets.Post("/", requireAuth, middleware.RequireRole(db, "bendahara"), handler.Create)
func RequireRole(db *sql.DB, allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(uint64)
		if !ok || userID == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Akses ditolak, silakan login terlebih dahulu",
			})
		}

		var role string
		err := db.QueryRow(`SELECT role FROM users WHERE id = ?`, userID).Scan(&role)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"message": "Akun tidak ditemukan, silakan login kembali",
				})
			}
			log.Printf("[ERROR] RequireRole: gagal membaca role user %d: %v", userID, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Terjadi kesalahan pada server",
			})
		}

		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Locals("userRole", role)
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Anda tidak memiliki hak akses untuk fitur ini",
		})
	}
}
