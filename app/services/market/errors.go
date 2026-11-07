package market

import (
	"errors"
	"net/http"
)

var ErrProviderNotConfigured = errors.New("market data provider is not configured")

// ProviderError describes the upstream service, not the signed-in user's
// credentials or request. Do not forward provider bodies or credential URLs.
type ProviderError struct {
	StatusCode int
}

func (e *ProviderError) Code() string {
	switch e.StatusCode {
	case http.StatusPaymentRequired:
		return "market_daily_limit_reached"
	case http.StatusTooManyRequests:
		return "market_rate_limited"
	case http.StatusUnauthorized, http.StatusForbidden:
		return "market_provider_access_denied"
	default:
		return "market_unavailable"
	}
}

func (e *ProviderError) Error() string {
	switch e.Code() {
	case "market_daily_limit_reached":
		return "The market data provider's daily quota has been exhausted. Try again after the quota resets."
	case "market_rate_limited":
		return "The market data provider is receiving too many requests. Please try again later."
	case "market_provider_access_denied":
		return "Market data access is unavailable. The service administrator must check the provider credentials and subscription."
	default:
		return "Market data is temporarily unavailable. Please try again later."
	}
}
