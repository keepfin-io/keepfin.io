package market

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	marketservice "keepnode/app/services/market"
	"keepnode/pkg/apiresponse"
)

func TestMarketErrorResponse(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"daily quota", &marketservice.ProviderError{StatusCode: 402}, 503, "market_daily_limit_reached"},
		{"burst limit", &marketservice.ProviderError{StatusCode: 429}, 503, "market_rate_limited"},
		{"provider credentials do not sign user out", &marketservice.ProviderError{StatusCode: 401}, 503, "market_provider_access_denied"},
		{"provider subscription", &marketservice.ProviderError{StatusCode: 403}, 503, "market_provider_access_denied"},
		{"transport failure", &marketservice.ProviderError{}, 503, "market_unavailable"},
		{"missing config", marketservice.ErrProviderNotConfigured, 503, "market_not_configured"},
		{"invalid input", errors.New("symbol is required"), 400, "invalid_market_request"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New(fiber.Config{ErrorHandler: apiresponse.ErrorHandler})
			app.Get("/", func(c *fiber.Ctx) error { return marketErrorResponse(c, tc.err) })
			response, err := app.Test(httptest.NewRequest("GET", "/", nil))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			var body apiresponse.Envelope[any]
			if err = json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != tc.status || body.Error == nil || body.Error.Code != tc.code || body.Data != nil {
				t.Fatalf("unexpected error: %d %+v", response.StatusCode, body)
			}
		})
	}
}
