package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/transpara/backend/domain"
	"github.com/transpara/backend/usecase"
)

// AuthHandler menangani semua request HTTP untuk modul autentikasi.
// Handler hanya membaca request, memanggil use case, dan menulis response.
// Tidak ada business logic di sini.
type AuthHandler struct {
	authUseCase usecase.AuthUseCase
}

// NewAuthHandler membuat instance AuthHandler baru (dependency injection).
func NewAuthHandler(authUseCase usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUseCase: authUseCase}
}

// Register — POST /api/v1/auth/register
// Mendaftarkan user baru dan mengembalikan OTP simulasi.
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req domain.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid, pastikan body berupa JSON",
		})
	}

	result, err := h.authUseCase.Register(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": result.Message,
		"data": fiber.Map{
			"user":     result.User,
			"otp_code": result.OTPCode,
		},
	})
}

// VerifyOTP — POST /api/v1/auth/verify-otp
// Memverifikasi OTP yang dimasukkan user.
func (h *AuthHandler) VerifyOTP(c *fiber.Ctx) error {
	var req domain.VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid, pastikan body berupa JSON",
		})
	}

	userResponse, err := h.authUseCase.VerifyOTP(&req)
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
// Membuat dan mengirim ulang OTP baru untuk user yang belum terverifikasi.
func (h *AuthHandler) ResendOTP(c *fiber.Ctx) error {
	var req domain.ResendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid, pastikan body berupa JSON",
		})
	}

	result, err := h.authUseCase.ResendOTP(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": result.Message,
		"data": fiber.Map{
			"otp_code": result.OTPCode,
		},
	})
}

// Login — POST /api/v1/auth/login
// Autentikasi user yang sudah terverifikasi.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid, pastikan body berupa JSON",
		})
	}

	result, err := h.authUseCase.Login(&req)
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
