package middleware

import (
	"github.com/amantyagi23/authbackend/pkg/jwt"
	"github.com/amantyagi23/authbackend/pkg/response"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const LocalsUserIDKey = "userID"

func EnsureAuthentication(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies("auth_token")

		if token == "" {
			return response.Unauthorized(c, "missing access token")
		}

		claims, err := jwt.VerifyAccessToken(token)
		if err != nil {
			return response.Unauthorized(c, "invalid or expired access token")
		}

		userID, err := uuid.Parse(claims.UserID.String())

		if err != nil {
			log.Warn(
				"auth middleware: invalid subject in token",
				zap.String("subject", claims.Subject),
			)

			return response.Unauthorized(c, "invalid token payload")
		}

		c.Locals(LocalsUserIDKey, userID)

		return c.Next()
	}
}

func UserIDFromContext(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(LocalsUserIDKey).(uuid.UUID)
	return id, ok
}
