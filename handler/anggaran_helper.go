package handler

import (
	"errors"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/transpara/backend/domain"
)

// getAuthUserID mengambil ID user yang sedang login dari c.Locals("userID").
// Nilainya diisi middleware RequireAuth setelah token JWT diverifikasi.
func getAuthUserID(c *fiber.Ctx) (uint64, bool) {
	id, ok := c.Locals("userID").(uint64)
	return id, ok && id > 0
}

// parsePathID membaca parameter :id dari URL dan memastikan angka positif.
func parsePathID(c *fiber.Ctx) (uint64, bool) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	return id, err == nil && id > 0
}

// respondOK mengirim response sukses dengan format yang sama seperti AuthHandler.
func respondOK(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// respondFail mengirim response gagal dengan format yang sama seperti AuthHandler.
func respondFail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}

// respondDomainError menerjemahkan error dari usecase menjadi HTTP status.
//
//	404 : pos anggaran / pengajuan tidak ditemukan
//	409 : konflik state (sudah vote, voting sudah ditutup)
//	422 : aturan bisnis tidak terpenuhi (di luar periode, melebihi sisa dana, dst.)
//	400 : error validasi input (errors.New biasa dari usecase)
//	500 : error infrastruktur (error DB dibungkus fmt.Errorf("...: %w") oleh
//	      repository), pesan aslinya hanya masuk log, tidak dikirim ke client.
func respondDomainError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrBudgetNotFound),
		errors.Is(err, domain.ErrExpenseNotFound):
		return respondFail(c, fiber.StatusNotFound, err.Error())

	case errors.Is(err, domain.ErrAlreadyVoted),
		errors.Is(err, domain.ErrRequestNotPending):
		return respondFail(c, fiber.StatusConflict, err.Error())

	case errors.Is(err, domain.ErrOutsideBudgetPeriod),
		errors.Is(err, domain.ErrAmountExceedsRemaining),
		errors.Is(err, domain.ErrBudgetPeriodInvalid),
		errors.Is(err, domain.ErrBudgetBelowUsed),
		errors.Is(err, domain.ErrRequestNotApproved):
		return respondFail(c, fiber.StatusUnprocessableEntity, err.Error())

	case errors.Unwrap(err) != nil:
		log.Printf("[ERROR] %v", err)
		return respondFail(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}

	return respondFail(c, fiber.StatusBadRequest, err.Error())
}