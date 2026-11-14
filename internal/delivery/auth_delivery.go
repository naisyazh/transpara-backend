package delivery

import (
	"github.com/gofiber/fiber/v2"
	"github.com/transpara/backend/internal/domain"
	"github.com/transpara/backend/internal/usecase"
)

// AuthDelivery menerima semua request HTTP untuk modul autentikasi.
// Tidak ada business logic di sini — semua diteruskan ke use case.
type AuthDelivery struct {
	authUseCase usecase.AuthUseCase
}

func NewAuthDelivery(authUseCase usecase.AuthUseCase) *AuthDelivery {
	return &AuthDelivery{authUseCase: authUseCase}
}

// Register — POST /api/v1/auth/register
func (d *AuthDelivery) Register(c *fiber.Ctx) error {
	var req domain.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid, pastikan body berupa JSON",
		})
	}

	result, err := d.authUseCase.Register(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": result.Message,
		"data":    result.User,
	})
}

// VerifyOTP — POST /api/v1/auth/verify-otp
func (d *AuthDelivery) VerifyOTP(c *fiber.Ctx) error {
	var req domain.VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid, pastikan body berupa JSON",
		})
	}

	userResponse, err := d.authUseCase.VerifyOTP(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Verifikasi berhasil, akun Anda sudah aktif",
		"data":    userResponse,
	})
}

// ResendOTP — POST /api/v1/auth/resend-otp
func (d *AuthDelivery) ResendOTP(c *fiber.Ctx) error {
	var req domain.ResendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid, pastikan body berupa JSON",
		})
	}

	result, err := d.authUseCase.ResendOTP(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": result.Message,
	})
}

// Login — POST /api/v1/auth/login
func (d *AuthDelivery) Login(c *fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid, pastikan body berupa JSON",
		})
	}

	result, err := d.authUseCase.Login(&req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Login berhasil",
		"data":    result,
	})
}