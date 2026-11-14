package delivery

import (
	"github.com/gofiber/fiber/v2"
	"github.com/transpara/backend/internal/domain"
	"github.com/transpara/backend/internal/usecase"
)

// BudgetDelivery menangani request HTTP untuk pos anggaran.
// Tidak ada business logic di sini, semua diteruskan ke use case.
type BudgetDelivery struct {
	budgetUseCase usecase.BudgetUseCase
}

func NewBudgetDelivery(budgetUseCase usecase.BudgetUseCase) *BudgetDelivery {
	return &BudgetDelivery{budgetUseCase: budgetUseCase}
}

// Create — POST /api/v1/budgets (khusus bendahara, dijaga di router)
func (h *BudgetDelivery) Create(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return respondFail(c, fiber.StatusUnauthorized, "Sesi tidak valid, silakan login kembali")
	}

	var req domain.BudgetRequest
	if err := c.BodyParser(&req); err != nil {
		return respondFail(c, fiber.StatusBadRequest, "Format request tidak valid, pastikan body berupa JSON")
	}

	result, err := h.budgetUseCase.Create(userID, &req)
	if err != nil {
		return respondDomainError(c, err)
	}

	return respondOK(c, fiber.StatusCreated, "Pos anggaran berhasil dibuat", result)
}

// Update — PUT /api/v1/budgets/:id (khusus bendahara, dijaga di router)
func (h *BudgetDelivery) Update(c *fiber.Ctx) error {
	id, ok := parsePathID(c)
	if !ok {
		return respondFail(c, fiber.StatusBadRequest, "ID pos anggaran tidak valid")
	}

	var req domain.BudgetRequest
	if err := c.BodyParser(&req); err != nil {
		return respondFail(c, fiber.StatusBadRequest, "Format request tidak valid, pastikan body berupa JSON")
	}

	result, err := h.budgetUseCase.Update(id, &req)
	if err != nil {
		return respondDomainError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Pos anggaran berhasil diperbarui", result)
}

// GetByID — GET /api/v1/budgets/:id
func (h *BudgetDelivery) GetByID(c *fiber.Ctx) error {
	id, ok := parsePathID(c)
	if !ok {
		return respondFail(c, fiber.StatusBadRequest, "ID pos anggaran tidak valid")
	}

	result, err := h.budgetUseCase.GetByID(id)
	if err != nil {
		return respondDomainError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Detail pos anggaran berhasil diambil", result)
}

// List — GET /api/v1/budgets
func (h *BudgetDelivery) List(c *fiber.Ctx) error {
	result, err := h.budgetUseCase.List()
	if err != nil {
		return respondDomainError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Daftar pos anggaran berhasil diambil", result)
}