package handlers

import (
	"expense-tracker/internal/api/middleware"
	"expense-tracker/internal/models"
	"expense-tracker/internal/repository"
	"expense-tracker/internal/services"
	pkgerrors "expense-tracker/pkg/errors"
	"expense-tracker/pkg/response"
	"expense-tracker/pkg/validator"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PersonHandler struct {
	personRepo repository.PersonRepository
	loanSvc    services.LoanService
	validator  *validator.Validator
}

func NewPersonHandler(personRepo repository.PersonRepository, loanSvc services.LoanService, v *validator.Validator) *PersonHandler {
	return &PersonHandler{personRepo: personRepo, loanSvc: loanSvc, validator: v}
}

type personInput struct {
	Name  string  `json:"name"  validate:"required,min=1,max=100"`
	Phone *string `json:"phone" validate:"omitempty,max=30"`
	Notes *string `json:"notes" validate:"omitempty,max=1000"`
}

func (h *PersonHandler) List(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	people, err := h.loanSvc.ListPeopleWithBalances(c.Context(), userID)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch people")
	}
	return response.OK(c, people)
}

func (h *PersonHandler) Create(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	var input personInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.validator.Validate(&input); err != nil {
		return response.UnprocessableEntity(c, err.Error())
	}

	p := &models.Person{
		UserID: userID,
		Name:   input.Name,
		Phone:  input.Phone,
		Notes:  input.Notes,
	}
	if err := h.personRepo.Create(c.Context(), p); err != nil {
		return response.InternalServerError(c, "failed to create person")
	}
	return response.Created(c, p)
}

func (h *PersonHandler) GetByID(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid person id")
	}

	p, err := h.loanSvc.GetPersonWithBalance(c.Context(), id, userID)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return response.NotFound(c, "person not found")
		}
		return response.InternalServerError(c, "failed to fetch person")
	}
	return response.OK(c, p)
}

func (h *PersonHandler) Update(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid person id")
	}

	p, err := h.personRepo.FindByID(c.Context(), id, userID)
	if err != nil {
		return response.NotFound(c, "person not found")
	}

	var input personInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.validator.Validate(&input); err != nil {
		return response.UnprocessableEntity(c, err.Error())
	}

	p.Name = input.Name
	p.Phone = input.Phone
	p.Notes = input.Notes

	if err := h.personRepo.Update(c.Context(), p); err != nil {
		return response.InternalServerError(c, "failed to update person")
	}
	return response.OK(c, p)
}

func (h *PersonHandler) Delete(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid person id")
	}

	if _, err := h.personRepo.FindByID(c.Context(), id, userID); err != nil {
		return response.NotFound(c, "person not found")
	}

	hasOutstanding, err := h.loanSvc.HasOutstandingLoans(c.Context(), id, userID)
	if err != nil {
		return response.InternalServerError(c, "failed to check outstanding loans")
	}
	if hasOutstanding {
		return response.Conflict(c, "person has outstanding loans")
	}

	if err := h.personRepo.Delete(c.Context(), id, userID); err != nil {
		return response.InternalServerError(c, "failed to delete person")
	}
	return response.NoContent(c)
}
