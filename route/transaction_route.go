package route

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/transpara/backend/app/model"
	"github.com/transpara/backend/app/service"
)

// TransactionRoute menangani request HTTP untuk transaksi keuangan.
// Tidak ada business logic di sini, semua diteruskan ke service.
type TransactionRoute struct {
	transactionService service.TransactionService
}

func NewTransactionRoute(transactionService service.TransactionService) *TransactionRoute {
	return &TransactionRoute{transactionService: transactionService}
}

// Create — POST /api/v1/transactions (khusus bendahara, dijaga di router)
func (h *TransactionRoute) Create(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return respondFail(c, fiber.StatusUnauthorized, "Sesi tidak valid, silakan login kembali")
	}

	var req model.CreateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return respondFail(c, fiber.StatusBadRequest, "Format request tidak valid, pastikan body berupa JSON")
	}

	result, err := h.transactionService.Create(userID, &req)
	if err != nil {
		return respondDomainError(c, err)
	}

	return respondOK(c, fiber.StatusCreated, "Transaksi berhasil dicatat", result)
}

// GetByID — GET /api/v1/transactions/:id
func (h *TransactionRoute) GetByID(c *fiber.Ctx) error {
	id, ok := parsePathID(c)
	if !ok {
		return respondFail(c, fiber.StatusBadRequest, "ID transaksi tidak valid")
	}

	result, err := h.transactionService.GetByID(id)
	if err != nil {
		return respondDomainError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Detail transaksi berhasil diambil", result)
}

// List — GET /api/v1/transactions
// Filter opsional: ?type=income|expense&category_id=1&from=YYYY-MM-DD&to=YYYY-MM-DD
func (h *TransactionRoute) List(c *fiber.Ctx) error {
	var categoryID uint64
	if raw := c.Query("category_id"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id == 0 {
			return respondFail(c, fiber.StatusBadRequest, "category_id tidak valid")
		}
		categoryID = id
	}

	result, err := h.transactionService.List(
		c.Query("type"), categoryID, c.Query("from"), c.Query("to"),
	)
	if err != nil {
		return respondDomainError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Daftar transaksi berhasil diambil", result)
}

// Summary — GET /api/v1/transactions/summary
func (h *TransactionRoute) Summary(c *fiber.Ctx) error {
	result, err := h.transactionService.GetSummary()
	if err != nil {
		return respondDomainError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Ringkasan saldo berhasil diambil", result)
}