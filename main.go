package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/transpara/backend/internal/delivery"
	"github.com/transpara/backend/internal/delivery/router"
	"github.com/transpara/backend/internal/infrastructure/email"
	"github.com/transpara/backend/internal/repository"
	"github.com/transpara/backend/internal/usecase"
	"github.com/transpara/backend/pkg/config"
	"github.com/transpara/backend/pkg/database"
	"github.com/transpara/backend/pkg/helper"
)

func main() {
	// ── 1. Konfigurasi ────────────────────────────────────────────
	cfg := config.Load()

	// ── 2. Koneksi Database ───────────────────────────────────────
	db := database.NewMySQLConnection(cfg)
	defer db.Close()

	// ── 3. JWT Helper ─────────────────────────────────────────────
	jwtHelper := helper.NewJWTHelper(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTExpiredHour)

	// ── 4. Email Service ──────────────────────────────────────────
	emailService := email.NewEmailService(cfg)

	// ── 5. Dependency Injection ───────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	otpRepo := repository.NewOTPRepository(db)
	authUseCase := usecase.NewAuthUseCase(userRepo, otpRepo, jwtHelper, emailService)
	authDelivery := delivery.NewAuthDelivery(authUseCase)

	// ── 6. Setup Aplikasi Fiber ───────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName: cfg.AppName,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := "Terjadi kesalahan pada server"
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				message = e.Message
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": message,
			})
		},
	})

	// ── 7. Routes ─────────────────────────────────────────────────
	router.SetupRoutes(app, authDelivery, jwtHelper)

	// ── 8. Endpoint Tidak Ditemukan (404) ─────────────────────────
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Endpoint tidak ditemukan",
		})
	})

	// ── 9. Jalankan Server ────────────────────────────────────────
	log.Printf("=========================================")
	log.Printf("  %s berjalan di port %s", cfg.AppName, cfg.AppPort)
	log.Printf("  SMTP: %s", cfg.SMTPUsername)
	log.Printf("=========================================")
	log.Printf("  POST /api/v1/auth/register")
	log.Printf("  POST /api/v1/auth/verify-otp")
	log.Printf("  POST /api/v1/auth/resend-otp")
	log.Printf("  POST /api/v1/auth/login")
	log.Printf("  GET  /api/v1/auth/me        [JWT required]")
	log.Printf("=========================================")

	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}