package handlers

import (
	"expense-tracker/internal/api/middleware"
	"expense-tracker/internal/models"
	"expense-tracker/internal/services"
	pkgerrors "expense-tracker/pkg/errors"
	"expense-tracker/pkg/response"
	"expense-tracker/pkg/validator"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TransactionHandler struct {
	txSvc     services.TransactionService
	validator *validator.Validator
}

func NewTransactionHandler(txSvc services.TransactionService, v *validator.Validator) *TransactionHandler {
	return &TransactionHandler{txSvc: txSvc, validator: v}
}

func (h *TransactionHandler) Create(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	var input services.CreateTransactionInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.validator.Validate(&input); err != nil {
		return response.UnprocessableEntity(c, err.Error())
	}

	t, err := h.txSvc.Create(c.Context(), userID, &input)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrInvalidInput) {
			return response.BadRequest(c, err.Error())
		}
		return response.InternalServerError(c, "failed to create transaction")
	}
	return response.Created(c, t)
}

func (h *TransactionHandler) List(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("perPage", "20"))

	filter := &models.TransactionFilter{
		Type:    models.TransactionType(c.Query("type")),
		Search:  c.Query("search"),
		Page:    page,
		PerPage: perPage,
	}

	if catIDStr := c.Query("categoryId"); catIDStr != "" {
		catID, err := uuid.Parse(catIDStr)
		if err == nil {
			filter.CategoryID = &catID
		}
	}
	if from := c.Query("from"); from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			filter.FromDate = &t
		}
	}
	if to := c.Query("to"); to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			filter.ToDate = &t
		}
	}

	transactions, total, err := h.txSvc.List(c.Context(), userID, filter)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch transactions")
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	return response.OKWithMeta(c, transactions, &response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *TransactionHandler) GetByID(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid transaction id")
	}

	t, err := h.txSvc.GetByID(c.Context(), id, userID)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return response.NotFound(c, "transaction not found")
		}
		return response.InternalServerError(c, "failed to fetch transaction")
	}
	return response.OK(c, t)
}

func (h *TransactionHandler) Update(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid transaction id")
	}

	var input services.UpdateTransactionInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.validator.Validate(&input); err != nil {
		return response.UnprocessableEntity(c, err.Error())
	}

	t, err := h.txSvc.Update(c.Context(), id, userID, &input)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return response.NotFound(c, "transaction not found")
		}
		if pkgerrors.Is(err, pkgerrors.ErrInvalidInput) {
			return response.BadRequest(c, err.Error())
		}
		return response.InternalServerError(c, "failed to update transaction")
	}
	return response.OK(c, t)
}

func (h *TransactionHandler) Delete(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid transaction id")
	}

	if err := h.txSvc.Delete(c.Context(), id, userID); err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return response.NotFound(c, "transaction not found")
		}
		return response.InternalServerError(c, "failed to delete transaction")
	}
	return response.NoContent(c)
}

func (h *TransactionHandler) Summary(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

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

	summary, err := h.txSvc.GetSummary(c.Context(), userID, from, to)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch summary")
	}
	return response.OK(c, summary)
}
