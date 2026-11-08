package market

import (
	"errors"
	"keepnode/pkg/apiresponse"

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
	return apiresponse.NewError(status, code, err.Error())
}
