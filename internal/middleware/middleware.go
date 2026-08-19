package middleware

import (
	"github.com/amantyagi23/authbackend/pkg/jwt"
	"github.com/amantyagi23/authbackend/pkg/response"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const LocalsUserIDKey = "userID"

// RequireAuth validates the auth_token cookie and attaches the user ID
// to the request context. Use on any route that needs an authenticated user.
func EnsureAuthentication(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies("auth_token")
		if token == "" {
			return response.Unauthorized(c, "missing auth token")
		}

		claims, err := jwt.VerifyToken(token)
		if err != nil {
			return response.Unauthorized(c, "invalid or expired session")
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			log.Warn("auth middleware: invalid user_id in token claims", zap.String("raw", claims.UserID))
			return response.Unauthorized(c, "invalid token payload")
		}

		c.Locals(LocalsUserIDKey, userID)

		return c.Next()
	}
}

// UserIDFromContext pulls the authenticated user's ID out of Fiber locals.
// Call this from handlers registered behind RequireAuth.
func UserIDFromContext(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(LocalsUserIDKey).(uuid.UUID)
	return id, ok
}
