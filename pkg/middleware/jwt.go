package middleware

import (
	"keepnode/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

func JWT() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": true,
				"msg":   "Missing or invalid token",
			})
		}

		parsedToken, err := utils.ParseJWT(token)
		if err != nil || !parsedToken.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": true,
				"msg":   "Invalid token",
			})
		}

		return c.Next()
	}
}
