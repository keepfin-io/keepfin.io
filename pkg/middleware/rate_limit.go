package middleware

import (
	"net/netip"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/google/uuid"
)

func requestLimit(max int, key func(*fiber.Ctx) string) fiber.Handler {
	return limiter.New(limiter.Config{
		Max: max, Expiration: time.Minute, KeyGenerator: key,
		LimitReached: func(c *fiber.Ctx) error {
			c.Set("Cache-Control", "no-store")
			return fiber.NewError(fiber.StatusTooManyRequests, "Too many requests; try again later")
		},
	})
}

// IPRequestLimit runs before auth/DB work. IPv6 clients share a /64 budget.
func IPRequestLimit(max int) fiber.Handler {
	return requestLimit(max, func(c *fiber.Ctx) string {
		ip := ClientIP(c)
		if address, err := netip.ParseAddr(ip); err == nil && address.Is6() && !address.Is4In6() {
			ip = netip.PrefixFrom(address, 64).Masked().String()
		}
		return "ip:" + ip
	})
}

// UserRequestLimit must be mounted after Session; never use a caller's user ID.
func UserRequestLimit(max int) fiber.Handler {
	limit := requestLimit(max, func(c *fiber.Ctx) string { return "user:" + c.Locals("user_id").(uuid.UUID).String() })
	return func(c *fiber.Ctx) error {
		id, ok := c.Locals("user_id").(uuid.UUID)
		if !ok || id == uuid.Nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Authentication required")
		}
		return limit(c)
	}
}
