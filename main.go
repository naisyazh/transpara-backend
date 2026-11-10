package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/transpara/backend/config"
	"github.com/transpara/backend/database"
	"github.com/transpara/backend/handler"
	"github.com/transpara/backend/helper"
	"github.com/transpara/backend/repository"
	"github.com/transpara/backend/router"
	"github.com/transpara/backend/usecase"
)

func main() {
	// ── 1. Konfigurasi ────────────────────────────────────────────
	cfg := config.Load()

	// ── 2. Koneksi Database ───────────────────────────────────────
	db := database.NewMySQLConnection(cfg)
	defer db.Close()

	// ── 3. JWT Helper ─────────────────────────────────────────────
	// Dibuat di sini agar bisa dipakai oleh use case (generate token)
	// dan middleware (verify token)
	jwtHelper := helper.NewJWTHelper(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTExpiredHour)

	// ── 4. Dependency Injection ───────────────────────────────────
	// Rakit dari bawah ke atas: Repository → Use Case → Handler

	// Repository
	userRepo := repository.NewUserRepository(db)
	otpRepo  := repository.NewOTPRepository(db)

	// Use Case (terima jwtHelper untuk generate token saat login)
	authUseCase := usecase.NewAuthUseCase(userRepo, otpRepo, jwtHelper)

	// Handler
	authHandler := handler.NewAuthHandler(authUseCase)

	// ── 5. Setup Aplikasi Fiber ───────────────────────────────────
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

	// ── 6. Routes (jwtHelper diteruskan untuk middleware) ─────────
	router.SetupRoutes(app, authHandler, jwtHelper)

	// ── 7. 404 Handler ────────────────────────────────────────────
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Endpoint tidak ditemukan",
		})
	})

	// ── 8. Jalankan Server ────────────────────────────────────────
	log.Printf("=========================================")
	log.Printf("  %s berjalan di port %s", cfg.AppName, cfg.AppPort)
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
