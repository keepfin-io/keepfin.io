package middleware

import (
	"keepnode/pkg/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func JWT() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var token string

		// First, try to get token from cookie
		token = c.Cookies("auth_token")

		// If not in cookie, try Authorization header (for backwards compatibility)
		if token == "" {
			authHeader := c.Get("Authorization")
			if authHeader != "" {
				// Remove "Bearer " prefix if present
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// No token found
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": true,
				"msg":   "Missing or invalid token",
			})
		}

		// Parse and validate token
		parsedToken, err := utils.ParseJWT(token)
		if err != nil || !parsedToken.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": true,
				"msg":   "Invalid token",
			})
		}

		// Store user info in context for later use
		c.Locals("user", parsedToken)

		return c.Next()
	}
}
