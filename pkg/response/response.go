package response

import "github.com/gofiber/fiber/v2"

type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"perPage"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

func OK(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(envelope{
		Success: true,
		Data:    data,
	})
}

func OKWithMeta(c *fiber.Ctx, data interface{}, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(envelope{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

func Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(envelope{
		Success: true,
		Data:    data,
	})
}

func Message(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusOK).JSON(envelope{
		Success: true,
		Message: message,
	})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func BadRequest(c *fiber.Ctx, err string) error {
	return c.Status(fiber.StatusBadRequest).JSON(envelope{
		Success: false,
		Error:   err,
	})
}

func Unauthorized(c *fiber.Ctx, err string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(envelope{
		Success: false,
		Error:   err,
	})
}

func Forbidden(c *fiber.Ctx, err string) error {
	return c.Status(fiber.StatusForbidden).JSON(envelope{
		Success: false,
		Error:   err,
	})
}

func NotFound(c *fiber.Ctx, err string) error {
	return c.Status(fiber.StatusNotFound).JSON(envelope{
		Success: false,
		Error:   err,
	})
}

func Conflict(c *fiber.Ctx, err string) error {
	return c.Status(fiber.StatusConflict).JSON(envelope{
		Success: false,
		Error:   err,
	})
}

func UnprocessableEntity(c *fiber.Ctx, err string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(envelope{
		Success: false,
		Error:   err,
	})
}

func InternalServerError(c *fiber.Ctx, err string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(envelope{
		Success: false,
		Error:   err,
	})
}
