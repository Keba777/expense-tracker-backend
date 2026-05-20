package routes

import (
	"expense-tracker/internal/api/handlers"
	"expense-tracker/internal/api/middleware"
	"expense-tracker/pkg/jwt"

	"github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Handlers struct {
	Auth        *handlers.AuthHandler
	Transaction *handlers.TransactionHandler
	Category    *handlers.CategoryHandler
	Report      *handlers.ReportHandler
}

func Register(app *fiber.App, h *Handlers, jwtManager *jwt.Manager) {
	app.Use(recover.New())
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api/v1")

	auth := api.Group("/auth")
	auth.Post("/register", h.Auth.Register)
	auth.Post("/login", h.Auth.Login)
	auth.Post("/refresh", h.Auth.Refresh)
	auth.Get("/me", middleware.Auth(jwtManager), h.Auth.Me)

	users := api.Group("/users", middleware.Auth(jwtManager))
	users.Put("/profile", h.Auth.UpdateProfile)

	protected := api.Use(middleware.Auth(jwtManager))

	txn := protected.Group("/transactions")
	txn.Get("/", h.Transaction.List)
	txn.Post("/", h.Transaction.Create)
	txn.Get("/summary", h.Transaction.Summary)
	txn.Get("/export", h.Transaction.Export)
	txn.Get("/:id", h.Transaction.GetByID)
	txn.Put("/:id", h.Transaction.Update)
	txn.Delete("/:id", h.Transaction.Delete)

	cats := protected.Group("/categories")
	cats.Get("/", h.Category.List)
	cats.Post("/", h.Category.Create)
	cats.Put("/:id", h.Category.Update)
	cats.Delete("/:id", h.Category.Delete)

	reports := protected.Group("/reports")
	reports.Get("/monthly", h.Report.Monthly)
	reports.Get("/weekly", h.Report.Weekly)
	reports.Get("/trends", h.Report.Trends)
	reports.Get("/category-breakdown", h.Report.CategoryBreakdown)
}
