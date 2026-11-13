package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/transpara/backend/domain"
	"github.com/transpara/backend/usecase"
)

// ExpenseHandler menangani request HTTP untuk pengajuan dana dan voting.
// Tidak ada business logic di sini, semua diteruskan ke use case.
type ExpenseHandler struct {
	expenseUseCase usecase.ExpenseUseCase
}

func NewExpenseHandler(expenseUseCase usecase.ExpenseUseCase) *ExpenseHandler {
	return &ExpenseHandler{expenseUseCase: expenseUseCase}
}

// Create — POST /api/v1/expense-requests (anggota dan bendahara)
func (h *ExpenseHandler) Create(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return respondFail(c, fiber.StatusUnauthorized, "Sesi tidak valid, silakan login kembali")
	}

	var req domain.CreateExpenseRequest
	if err := c.BodyParser(&req); err != nil {
		return respondFail(c, fiber.StatusBadRequest, "Format request tidak valid, pastikan body berupa JSON")
	}

	result, err := h.expenseUseCase.Create(userID, &req)
	if err != nil {
		return respondDomainError(c, err)
	}

	return respondOK(c, fiber.StatusCreated, "Pengajuan dana berhasil dibuat, menunggu voting anggota", result)
}

// List — GET /api/v1/expense-requests?status=pending|approved|rejected
func (h *ExpenseHandler) List(c *fiber.Ctx) error {
	result, err := h.expenseUseCase.List(c.Query("status"))
	if err != nil {
		return respondDomainError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Daftar pengajuan dana berhasil diambil", result)
}

// GetByID — GET /api/v1/expense-requests/:id
func (h *ExpenseHandler) GetByID(c *fiber.Ctx) error {
	id, ok := parsePathID(c)
	if !ok {
		return respondFail(c, fiber.StatusBadRequest, "ID pengajuan tidak valid")
	}

	result, err := h.expenseUseCase.GetByID(id)
	if err != nil {
		return respondDomainError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Detail pengajuan dana berhasil diambil", result)
}

// Vote — POST /api/v1/expense-requests/:id/vote
// Body: {"vote": "yes"} atau {"vote": "no"}
func (h *ExpenseHandler) Vote(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return respondFail(c, fiber.StatusUnauthorized, "Sesi tidak valid, silakan login kembali")
	}

	requestID, ok := parsePathID(c)
	if !ok {
		return respondFail(c, fiber.StatusBadRequest, "ID pengajuan tidak valid")
	}

	var req domain.VoteRequest
	if err := c.BodyParser(&req); err != nil {
		return respondFail(c, fiber.StatusBadRequest, "Format request tidak valid, pastikan body berupa JSON")
	}

	result, err := h.expenseUseCase.Vote(userID, requestID, &req)
	if err != nil {
		return respondDomainError(c, err)
	}

	message := "Suara Anda tercatat, voting masih berjalan"
	switch result.Status {
	case domain.ExpenseStatusApproved:
		message = "Suara Anda tercatat, pengajuan disetujui"
	case domain.ExpenseStatusRejected:
		message = "Suara Anda tercatat, pengajuan ditolak"
	}

	return respondOK(c, fiber.StatusOK, message, result)
}