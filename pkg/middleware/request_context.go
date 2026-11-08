package middleware

import (
	"github.com/gofiber/fiber/v2"
	"keepnode/pkg/apiresponse"
)

func RequestContext() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiresponse.RequestID(c)
		return c.Next()
	}
}
