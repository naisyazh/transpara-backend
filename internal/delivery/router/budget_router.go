package router

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/transpara/backend/internal/delivery"
	"github.com/transpara/backend/internal/delivery/middleware"
	"github.com/transpara/backend/pkg/helper"
)

// SetupBudgetRoutes mendaftarkan route modul Anggaran dan Voting.
// Dipisah dari SetupRoutes supaya tidak bentrok dengan perubahan modul lain.
//
// Semua route butuh login (RequireAuth). Membuat dan mengubah pos anggaran
// hanya boleh bendahara (RequireRole), anggota biasa mendapat 403 Forbidden.
func SetupBudgetRoutes(
	app *fiber.App,
	budgetDelivery *delivery.BudgetDelivery,
	expenseDelivery *delivery.ExpenseDelivery,
	jwtHelper *helper.JWTHelper,
	db *sql.DB,
) {
	api := app.Group("/api/v1")

	requireAuth := middleware.RequireAuth(jwtHelper)
	requireTreasurer := middleware.RequireRole(db, "bendahara")

	// ── Pos Anggaran ──────────────────────────────────────────────
	budgets := api.Group("/budgets")
	budgets.Get("/", requireAuth, budgetDelivery.List)
	budgets.Get("/:id", requireAuth, budgetDelivery.GetByID)
	budgets.Post("/", requireAuth, requireTreasurer, budgetDelivery.Create)
	budgets.Put("/:id", requireAuth, requireTreasurer, budgetDelivery.Update)

	// ── Pengajuan Dana dan Voting ─────────────────────────────────
	expenses := api.Group("/expense-requests")
	expenses.Get("/", requireAuth, expenseDelivery.List)
	expenses.Get("/:id", requireAuth, expenseDelivery.GetByID)
	expenses.Post("/", requireAuth, expenseDelivery.Create)
	expenses.Post("/:id/vote", requireAuth, expenseDelivery.Vote)
}