package handlers

import (
	"expense-tracker/internal/api/middleware"
	"expense-tracker/internal/models"
	"expense-tracker/internal/services"
	"expense-tracker/pkg/response"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ReportHandler struct {
	txSvc services.TransactionService
}

func NewReportHandler(txSvc services.TransactionService) *ReportHandler {
	return &ReportHandler{txSvc: txSvc}
}

func (h *ReportHandler) Monthly(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	now := time.Now()
	year, _ := strconv.Atoi(c.Query("year", strconv.Itoa(now.Year())))
	month, _ := strconv.Atoi(c.Query("month", strconv.Itoa(int(now.Month()))))

	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, -1)

	summary, err := h.txSvc.GetSummary(c.Context(), userID, from, to)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch monthly report")
	}

	breakdown, err := h.txSvc.GetCategoryBreakdown(c.Context(), userID, from, to, models.TransactionExpense)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch category breakdown")
	}

	daily, err := h.txSvc.GetDailyTotals(c.Context(), userID, from, to)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch daily totals")
	}

	return response.OK(c, fiber.Map{
		"period":    fiber.Map{"year": year, "month": month, "from": from, "to": to},
		"summary":   summary,
		"breakdown": breakdown,
		"daily":     daily,
	})
}

func (h *ReportHandler) Weekly(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	from := now.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)
	to := from.AddDate(0, 0, 6)

	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t
			to = from.AddDate(0, 0, 6)
		}
	}

	summary, err := h.txSvc.GetSummary(c.Context(), userID, from, to)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch weekly report")
	}

	daily, err := h.txSvc.GetDailyTotals(c.Context(), userID, from, to)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch daily totals")
	}

	breakdown, err := h.txSvc.GetCategoryBreakdown(c.Context(), userID, from, to, models.TransactionExpense)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch category breakdown")
	}

	return response.OK(c, fiber.Map{
		"period":    fiber.Map{"from": from, "to": to},
		"summary":   summary,
		"daily":     daily,
		"breakdown": breakdown,
	})
}

func (h *ReportHandler) Trends(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)
	months, _ := strconv.Atoi(c.Query("months", "6"))
	if months > 12 {
		months = 12
	}

	trends, err := h.txSvc.GetMonthlyTrends(c.Context(), userID, months)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch trends")
	}
	return response.OK(c, trends)
}

func (h *ReportHandler) CategoryBreakdown(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	txType := models.TransactionExpense
	if c.Query("type") == "income" {
		txType = models.TransactionIncome
	}

	now := time.Now()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, -1)

	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			to = t
		}
	}

	breakdown, err := h.txSvc.GetCategoryBreakdown(c.Context(), userID, from, to, txType)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch category breakdown")
	}
	return response.OK(c, breakdown)
}
