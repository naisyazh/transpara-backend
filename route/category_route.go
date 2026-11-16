package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/transpara/backend/app/model"
	"github.com/transpara/backend/app/service"
)

// CategoryRoute menangani request HTTP untuk kategori transaksi.
// Tidak ada business logic di sini, semua diteruskan ke service.
type CategoryRoute struct {
	categoryService service.CategoryService
}

func NewCategoryRoute(categoryService service.CategoryService) *CategoryRoute {
	return &CategoryRoute{categoryService: categoryService}
}

// Create — POST /api/v1/categories (khusus bendahara, dijaga di router)
func (h *CategoryRoute) Create(c *fiber.Ctx) error {
	var req model.CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return respondFail(c, fiber.StatusBadRequest, "Format request tidak valid, pastikan body berupa JSON")
	}

	result, err := h.categoryService.Create(&req)
	if err != nil {
		return respondDomainError(c, err)
	}

	return respondOK(c, fiber.StatusCreated, "Kategori berhasil dibuat", result)
}

// List — GET /api/v1/categories, filter opsional ?type=income|expense
func (h *CategoryRoute) List(c *fiber.Ctx) error {
	result, err := h.categoryService.List(c.Query("type"))
	if err != nil {
		return respondDomainError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Daftar kategori berhasil diambil", result)
}