package middleware

import (
	"github.com/amantyagi23/authbackend/pkg/jwt"
	"github.com/gofiber/fiber/v2"
)

func Protected(tokenName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenStr := c.Cookies(tokenName)

		claims, err := jwt.VerifyToken(tokenStr)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		c.Locals("user", claims)

		return c.Next()
	}
}
