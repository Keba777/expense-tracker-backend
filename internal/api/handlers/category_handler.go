package handlers

import (
	"context"
	"expense-tracker/internal/api/middleware"
	"expense-tracker/internal/models"
	"expense-tracker/internal/repository"
	pkgerrors "expense-tracker/pkg/errors"
	"expense-tracker/pkg/response"
	"expense-tracker/pkg/validator"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CategoryHandler struct {
	catRepo   repository.CategoryRepository
	validator *validator.Validator
}

func NewCategoryHandler(catRepo repository.CategoryRepository, v *validator.Validator) *CategoryHandler {
	return &CategoryHandler{catRepo: catRepo, validator: v}
}

type categoryInput struct {
	Name  string               `json:"name"  validate:"required,min=1,max=100"`
	Icon  string               `json:"icon"  validate:"required"`
	Color string               `json:"color" validate:"required,len=7"`
	Type  models.CategoryType  `json:"type"  validate:"required,oneof=income expense"`
}

func (h *CategoryHandler) List(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)
	cats, err := h.catRepo.FindByUserID(c.Context(), userID)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch categories")
	}
	return response.OK(c, cats)
}

func (h *CategoryHandler) Create(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	var input categoryInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.validator.Validate(&input); err != nil {
		return response.UnprocessableEntity(c, err.Error())
	}

	cat := &models.Category{
		UserID: userID,
		Name:   input.Name,
		Icon:   input.Icon,
		Color:  input.Color,
		Type:   input.Type,
	}

	if err := h.catRepo.Create(context.Background(), cat); err != nil {
		return response.InternalServerError(c, "failed to create category")
	}
	return response.Created(c, cat)
}

func (h *CategoryHandler) Update(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid category id")
	}

	cat, err := h.catRepo.FindByID(c.Context(), id, userID)
	if err != nil {
		return response.NotFound(c, "category not found")
	}

	var input categoryInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.validator.Validate(&input); err != nil {
		return response.UnprocessableEntity(c, err.Error())
	}

	cat.Name = input.Name
	cat.Icon = input.Icon
	cat.Color = input.Color
	cat.Type = input.Type

	if err := h.catRepo.Update(c.Context(), cat); err != nil {
		return response.InternalServerError(c, "failed to update category")
	}
	return response.OK(c, cat)
}

func (h *CategoryHandler) Delete(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid category id")
	}

	cat, err := h.catRepo.FindByID(c.Context(), id, userID)
	if err != nil {
		return response.NotFound(c, "category not found")
	}
	if cat.IsDefault {
		return response.Forbidden(c, "cannot delete a default category")
	}

	if err := h.catRepo.Delete(c.Context(), id, userID); err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return response.NotFound(c, "category not found")
		}
		return response.InternalServerError(c, "failed to delete category")
	}
	return response.NoContent(c)
}
