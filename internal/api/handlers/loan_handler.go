package handlers

import (
	"expense-tracker/internal/api/middleware"
	"expense-tracker/internal/models"
	"expense-tracker/internal/services"
	pkgerrors "expense-tracker/pkg/errors"
	"expense-tracker/pkg/response"
	"expense-tracker/pkg/validator"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type LoanHandler struct {
	loanSvc   services.LoanService
	validator *validator.Validator
}

func NewLoanHandler(loanSvc services.LoanService, v *validator.Validator) *LoanHandler {
	return &LoanHandler{loanSvc: loanSvc, validator: v}
}

func (h *LoanHandler) Create(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	var input services.CreateLoanInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.validator.Validate(&input); err != nil {
		return response.UnprocessableEntity(c, err.Error())
	}

	l, err := h.loanSvc.CreateLoan(c.Context(), userID, &input)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrInvalidInput) {
			return response.BadRequest(c, err.Error())
		}
		return response.InternalServerError(c, "failed to create loan")
	}
	return response.Created(c, l)
}

func (h *LoanHandler) List(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("perPage", "20"))

	filter := &models.LoanFilter{
		Direction: models.LoanDirection(c.Query("direction")),
		Status:    models.LoanStatus(c.Query("status")),
		Page:      page,
		PerPage:   perPage,
	}
	if personIDStr := c.Query("personId"); personIDStr != "" {
		if personID, err := uuid.Parse(personIDStr); err == nil {
			filter.PersonID = &personID
		}
	}

	loans, total, err := h.loanSvc.ListLoans(c.Context(), userID, filter)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch loans")
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	return response.OKWithMeta(c, loans, &response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *LoanHandler) GetByID(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid loan id")
	}

	l, err := h.loanSvc.GetLoan(c.Context(), id, userID)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return response.NotFound(c, "loan not found")
		}
		return response.InternalServerError(c, "failed to fetch loan")
	}
	return response.OK(c, l)
}

func (h *LoanHandler) Update(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid loan id")
	}

	var input services.UpdateLoanInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.validator.Validate(&input); err != nil {
		return response.UnprocessableEntity(c, err.Error())
	}

	l, err := h.loanSvc.UpdateLoan(c.Context(), id, userID, &input)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return response.NotFound(c, "loan not found")
		}
		if pkgerrors.Is(err, pkgerrors.ErrInvalidInput) {
			return response.BadRequest(c, err.Error())
		}
		return response.InternalServerError(c, "failed to update loan")
	}
	return response.OK(c, l)
}

func (h *LoanHandler) Delete(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid loan id")
	}

	if err := h.loanSvc.DeleteLoan(c.Context(), id, userID); err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return response.NotFound(c, "loan not found")
		}
		return response.InternalServerError(c, "failed to delete loan")
	}
	return response.NoContent(c)
}

func (h *LoanHandler) ListPayments(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	loanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid loan id")
	}

	payments, err := h.loanSvc.ListPayments(c.Context(), loanID, userID)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return response.NotFound(c, "loan not found")
		}
		return response.InternalServerError(c, "failed to fetch payments")
	}
	return response.OK(c, payments)
}

func (h *LoanHandler) AddPayment(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	loanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid loan id")
	}

	var input services.CreateLoanPaymentInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.validator.Validate(&input); err != nil {
		return response.UnprocessableEntity(c, err.Error())
	}

	l, err := h.loanSvc.AddPayment(c.Context(), loanID, userID, &input)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return response.NotFound(c, "loan not found")
		}
		if pkgerrors.Is(err, pkgerrors.ErrInvalidInput) {
			return response.BadRequest(c, "payment exceeds remaining balance")
		}
		return response.InternalServerError(c, "failed to record payment")
	}
	return response.Created(c, l)
}

func (h *LoanHandler) DeletePayment(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	loanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid loan id")
	}
	paymentID, err := uuid.Parse(c.Params("paymentId"))
	if err != nil {
		return response.BadRequest(c, "invalid payment id")
	}

	l, err := h.loanSvc.DeletePayment(c.Context(), loanID, paymentID, userID)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return response.NotFound(c, "loan or payment not found")
		}
		return response.InternalServerError(c, "failed to delete payment")
	}
	return response.OK(c, l)
}
