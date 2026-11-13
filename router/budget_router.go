package router

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/transpara/backend/handler"
	"github.com/transpara/backend/helper"
	"github.com/transpara/backend/middleware"
)

// SetupBudgetRoutes mendaftarkan route modul Anggaran dan Voting.
// Dipisah dari SetupRoutes supaya tidak bentrok dengan perubahan modul lain.
//
// Semua route butuh login (RequireAuth). Membuat dan mengubah pos anggaran
// hanya boleh bendahara (RequireRole), anggota biasa mendapat 403 Forbidden.
func SetupBudgetRoutes(
	app *fiber.App,
	budgetHandler *handler.BudgetHandler,
	expenseHandler *handler.ExpenseHandler,
	jwtHelper *helper.JWTHelper,
	db *sql.DB,
) {
	api := app.Group("/api/v1")

	requireAuth := middleware.RequireAuth(jwtHelper)
	requireTreasurer := middleware.RequireRole(db, "bendahara")

	// ── Pos Anggaran ──────────────────────────────────────────────
	budgets := api.Group("/budgets")
	budgets.Get("/", requireAuth, budgetHandler.List)
	budgets.Get("/:id", requireAuth, budgetHandler.GetByID)
	budgets.Post("/", requireAuth, requireTreasurer, budgetHandler.Create)
	budgets.Put("/:id", requireAuth, requireTreasurer, budgetHandler.Update)

	// ── Pengajuan Dana dan Voting ─────────────────────────────────
	expenses := api.Group("/expense-requests")
	expenses.Get("/", requireAuth, expenseHandler.List)
	expenses.Get("/:id", requireAuth, expenseHandler.GetByID)
	expenses.Post("/", requireAuth, expenseHandler.Create)
	expenses.Post("/:id/vote", requireAuth, expenseHandler.Vote)
}