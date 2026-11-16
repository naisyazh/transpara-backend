package route

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/transpara/backend/helper"
	"github.com/transpara/backend/middleware"
)

// SetupTransactionRoutes mendaftarkan route modul Transaksi Keuangan.
// Dipisah dari SetupRoutes supaya tidak bentrok dengan perubahan modul lain.
//
// Semua route butuh login (RequireAuth). Membuat kategori dan mencatat
// transaksi hanya boleh bendahara (RequireRole), anggota biasa mendapat 403.
func SetupTransactionRoutes(
	app *fiber.App,
	transactionRoute *TransactionRoute,
	categoryRoute *CategoryRoute,
	jwtHelper *helper.JWTHelper,
	db *sql.DB,
) {
	api := app.Group("/api/v1")

	requireAuth := middleware.RequireAuth(jwtHelper)
	requireTreasurer := middleware.RequireRole(db, "bendahara")

	// ── Kategori ──────────────────────────────────────────────────
	categories := api.Group("/categories")
	categories.Get("/", requireAuth, categoryRoute.List)
	categories.Post("/", requireAuth, requireTreasurer, categoryRoute.Create)

	// ── Transaksi ─────────────────────────────────────────────────
	// "/summary" harus didaftarkan SEBELUM "/:id", kalau tidak Fiber
	// menganggap "summary" sebagai nilai :id.
	transactions := api.Group("/transactions")
	transactions.Get("/", requireAuth, transactionRoute.List)
	transactions.Get("/summary", requireAuth, transactionRoute.Summary)
	transactions.Get("/:id", requireAuth, transactionRoute.GetByID)
	transactions.Post("/", requireAuth, requireTreasurer, transactionRoute.Create)
}