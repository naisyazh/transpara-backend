package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/transpara/backend/app/model"
	"github.com/transpara/backend/app/service"
)

// ExpenseRoute menangani request HTTP untuk pengajuan dana dan voting.
// Tidak ada business logic di sini, semua diteruskan ke use case.
type ExpenseRoute struct {
	expenseService service.ExpenseService
}

func NewExpenseRoute(expenseService service.ExpenseService) *ExpenseRoute {
	return &ExpenseRoute{expenseService: expenseService}
}

// Create — POST /api/v1/expense-requests (anggota dan bendahara)
func (h *ExpenseRoute) Create(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return respondFail(c, fiber.StatusUnauthorized, "Sesi tidak valid, silakan login kembali")
	}

	var req model.CreateExpenseRequest
	if err := c.BodyParser(&req); err != nil {
		return respondFail(c, fiber.StatusBadRequest, "Format request tidak valid, pastikan body berupa JSON")
	}

	result, err := h.expenseService.Create(userID, &req)
	if err != nil {
		return respondDomainError(c, err)
	}

	return respondOK(c, fiber.StatusCreated, "Pengajuan dana berhasil dibuat, menunggu voting anggota", result)
}

// List — GET /api/v1/expense-requests?status=pending|approved|rejected
func (h *ExpenseRoute) List(c *fiber.Ctx) error {
	result, err := h.expenseService.List(c.Query("status"))
	if err != nil {
		return respondDomainError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Daftar pengajuan dana berhasil diambil", result)
}

// GetByID — GET /api/v1/expense-requests/:id
func (h *ExpenseRoute) GetByID(c *fiber.Ctx) error {
	id, ok := parsePathID(c)
	if !ok {
		return respondFail(c, fiber.StatusBadRequest, "ID pengajuan tidak valid")
	}

	result, err := h.expenseService.GetByID(id)
	if err != nil {
		return respondDomainError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Detail pengajuan dana berhasil diambil", result)
}

// Vote — POST /api/v1/expense-requests/:id/vote
// Body: {"vote": "yes"} atau {"vote": "no"}
func (h *ExpenseRoute) Vote(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return respondFail(c, fiber.StatusUnauthorized, "Sesi tidak valid, silakan login kembali")
	}

	requestID, ok := parsePathID(c)
	if !ok {
		return respondFail(c, fiber.StatusBadRequest, "ID pengajuan tidak valid")
	}

	var req model.VoteRequest
	if err := c.BodyParser(&req); err != nil {
		return respondFail(c, fiber.StatusBadRequest, "Format request tidak valid, pastikan body berupa JSON")
	}

	result, err := h.expenseService.Vote(userID, requestID, &req)
	if err != nil {
		return respondDomainError(c, err)
	}

	message := "Suara Anda tercatat, voting masih berjalan"
	switch result.Status {
	case model.ExpenseStatusApproved:
		message = "Suara Anda tercatat, pengajuan disetujui"
	case model.ExpenseStatusRejected:
		message = "Suara Anda tercatat, pengajuan ditolak"
	}

	return respondOK(c, fiber.StatusOK, message, result)
}
