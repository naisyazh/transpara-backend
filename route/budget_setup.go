package route

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
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
	budgetRoute *BudgetRoute,
	expenseRoute *ExpenseRoute,
	jwtHelper *helper.JWTHelper,
	db *sql.DB,
) {
	api := app.Group("/api/v1")

	requireAuth := middleware.RequireAuth(jwtHelper)
	requireTreasurer := middleware.RequireRole(db, "bendahara")

	// ── Pos Anggaran ──────────────────────────────────────────────
	budgets := api.Group("/budgets")
	budgets.Get("/", requireAuth, budgetRoute.List)
	budgets.Get("/:id", requireAuth, budgetRoute.GetByID)
	budgets.Post("/", requireAuth, requireTreasurer, budgetRoute.Create)
	budgets.Put("/:id", requireAuth, requireTreasurer, budgetRoute.Update)

	// ── Pengajuan Dana dan Voting ─────────────────────────────────
	expenses := api.Group("/expense-requests")
	expenses.Get("/", requireAuth, expenseRoute.List)
	expenses.Get("/:id", requireAuth, expenseRoute.GetByID)
	expenses.Post("/", requireAuth, expenseRoute.Create)
	expenses.Post("/:id/vote", requireAuth, expenseRoute.Vote)
}
