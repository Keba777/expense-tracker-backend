package middleware

import (
	"expense-tracker/pkg/jwt"
	"expense-tracker/pkg/response"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const userIDKey = "userID"
const userEmailKey = "userEmail"
const userPlanKey = "userPlan"

func Auth(jwtManager *jwt.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Unauthorized(c, "missing authorization header")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return response.Unauthorized(c, "invalid authorization header format")
		}

		claims, err := jwtManager.ValidateAccess(parts[1])
		if err != nil {
			return response.Unauthorized(c, "invalid or expired token")
		}

		c.Locals(userIDKey, claims.UserID)
		c.Locals(userEmailKey, claims.Email)
		c.Locals(userPlanKey, claims.Plan)

		return c.Next()
	}
}

func UserIDFromCtx(c *fiber.Ctx) uuid.UUID {
	id, _ := c.Locals(userIDKey).(uuid.UUID)
	return id
}

func UserPlanFromCtx(c *fiber.Ctx) string {
	plan, _ := c.Locals(userPlanKey).(string)
	return plan
}
