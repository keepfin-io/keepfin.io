package market

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	marketservice "keepnode/app/services/market"
)

func marketErrorResponse(c *fiber.Ctx, err error) error {
	status, code := fiber.StatusBadRequest, "invalid_market_request"
	var provider *marketservice.ProviderError
	if errors.As(err, &provider) {
		status, code = fiber.StatusServiceUnavailable, provider.Code()
	} else if errors.Is(err, marketservice.ErrProviderNotConfigured) {
		status, code = fiber.StatusServiceUnavailable, "market_not_configured"
	}
	return c.Status(status).JSON(fiber.Map{"error": true, "code": code, "message": err.Error()})
}
