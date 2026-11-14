package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/transpara/backend/app/model"
	"github.com/transpara/backend/app/service"
)

// AuthRoute menerima semua request HTTP untuk modul autentikasi.
// Tidak ada business logic di sini — semua diteruskan ke use case.
type AuthRoute struct {
	authService service.AuthService
}

func NewAuthRoute(authService service.AuthService) *AuthRoute {
	return &AuthRoute{authService: authService}
}

// Register — POST /api/v1/auth/register
func (d *AuthRoute) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid, pastikan body berupa JSON",
		})
	}

	result, err := d.authService.Register(&req)
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
func (d *AuthRoute) VerifyOTP(c *fiber.Ctx) error {
	var req model.VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid, pastikan body berupa JSON",
		})
	}

	userResponse, err := d.authService.VerifyOTP(&req)
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
func (d *AuthRoute) ResendOTP(c *fiber.Ctx) error {
	var req model.ResendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid, pastikan body berupa JSON",
		})
	}

	result, err := d.authService.ResendOTP(&req)
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
func (d *AuthRoute) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid, pastikan body berupa JSON",
		})
	}

	result, err := d.authService.Login(&req)
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
