// Package authhandler adapts authenticated application handlers to Fiber routes.
package authhandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UserHandler receives the application UUID resolved by the session middleware.
type UserHandler func(c *fiber.Ctx, userID uuid.UUID) error

// WithUser requires the identity established by middleware.Session before calling
// the controller. It does not authenticate request headers or replace ownership checks.
func WithUser(handler UserHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(uuid.UUID)
		if !ok || userID == uuid.Nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Authentication required")
		}
		return handler(c, userID)
	}
}
