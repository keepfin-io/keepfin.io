package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type holdingIndexFlusher interface {
	FlushUser(context.Context, uuid.UUID) error
}

// HoldingSearchSync waits for read-after-write visibility, after the DB handler
// commits. An index outage must not turn a committed purchase into a retryable
// financial error. The durable outbox and fresh database read cover that case.
func HoldingSearchSync(index holdingIndexFlusher) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := c.Next(); err != nil {
			return err
		}
		if c.Method() == fiber.MethodGet || c.Method() == fiber.MethodHead ||
			c.Response().StatusCode() < 200 || c.Response().StatusCode() >= 300 {
			return nil
		}
		userID, ok := c.Locals("user_id").(uuid.UUID)
		if !ok {
			return nil
		}
		ctx, cancel := context.WithTimeout(c.UserContext(), 2500*time.Millisecond)
		defer cancel()
		state := "ready"
		if err := index.FlushUser(ctx, userID); err != nil {
			state = "pending"
		}
		c.Set("X-Holdings-Search-Sync", state)
		return nil
	}
}
