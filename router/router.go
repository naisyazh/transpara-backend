package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/transpara/backend/delivery"
	"github.com/transpara/backend/helper"
	"github.com/transpara/backend/middleware"
)

// SetupRoutes mendaftarkan semua route ke aplikasi Fiber.
// Pola URL: /api/v1/<modul>/<aksi>
func SetupRoutes(app *fiber.App, authHandler *handler.AuthHandler, jwtHelper *helper.JWTHelper) {
	api := app.Group("/api/v1")

	// ── Health Check ──────────────────────────────────────────────
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Server Transpara berjalan dengan baik",
		})
	})

	// ── Modul Autentikasi — Public (tidak butuh token) ────────────
	auth := api.Group("/auth")
	auth.Post("/register",   authHandler.Register)
	auth.Post("/verify-otp", authHandler.VerifyOTP)
	auth.Post("/resend-otp", authHandler.ResendOTP)
	auth.Post("/login",      authHandler.Login)

	// ── Contoh endpoint yang dilindungi JWT ───────────────────────
	// Endpoint ini membutuhkan header: Authorization: Bearer <token>
	// Middleware RequireAuth akan memvalidasi token sebelum masuk ke handler.
	//
	// Endpoint /me mengembalikan data user dari token (tanpa query DB).
	auth.Get("/me", middleware.RequireAuth(jwtHelper), func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Data user berhasil diambil",
			"data": fiber.Map{
				"user_id": c.Locals("userID"),
				"name":    c.Locals("userName"),
				"email":   c.Locals("userEmail"),
			},
		})
	})

	// ── Modul Transaksi (Developer 2) ─────────────────────────────
	// transactions := api.Group("/transactions")

	// ── Modul Anggaran dan Voting (Developer 3) ───────────────────
	// budgets := api.Group("/budgets")
}
