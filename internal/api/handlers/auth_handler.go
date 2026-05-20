package handlers

import (
	"expense-tracker/internal/api/middleware"
	"expense-tracker/internal/services"
	pkgerrors "expense-tracker/pkg/errors"
	"expense-tracker/pkg/response"
	"expense-tracker/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authSvc   services.AuthService
	validator *validator.Validator
}

func NewAuthHandler(authSvc services.AuthService, v *validator.Validator) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, validator: v}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var input services.RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.validator.Validate(&input); err != nil {
		return response.UnprocessableEntity(c, err.Error())
	}

	result, err := h.authSvc.Register(c.Context(), &input)
	if err != nil {
		switch {
		case pkgerrors.Is(err, pkgerrors.ErrConflict):
			return response.Conflict(c, "email already in use")
		default:
			return response.InternalServerError(c, "registration failed")
		}
	}
	return response.Created(c, result)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var input services.LoginInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.validator.Validate(&input); err != nil {
		return response.UnprocessableEntity(c, err.Error())
	}

	result, err := h.authSvc.Login(c.Context(), &input)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrInvalidCredentials) {
			return response.Unauthorized(c, "invalid email or password")
		}
		return response.InternalServerError(c, "login failed")
	}
	return response.OK(c, result)
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refreshToken" validate:"required"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.validator.Validate(&body); err != nil {
		return response.UnprocessableEntity(c, err.Error())
	}

	tokens, err := h.authSvc.RefreshTokens(c.Context(), body.RefreshToken)
	if err != nil {
		return response.Unauthorized(c, "invalid or expired refresh token")
	}
	return response.OK(c, tokens)
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)
	user, err := h.authSvc.GetUser(c.Context(), userID)
	if err != nil {
		return response.NotFound(c, "user not found")
	}
	return response.OK(c, user)
}

func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)

	var input services.UpdateProfileInput
	if err := c.BodyParser(&input); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.validator.Validate(&input); err != nil {
		return response.UnprocessableEntity(c, err.Error())
	}

	user, err := h.authSvc.UpdateProfile(c.Context(), userID, &input)
	if err != nil {
		switch {
		case pkgerrors.Is(err, pkgerrors.ErrNotFound):
			return response.NotFound(c, "user not found")
		default:
			return response.InternalServerError(c, "failed to update profile")
		}
	}
	return response.OK(c, user)
}
