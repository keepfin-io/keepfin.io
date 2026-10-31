package platform

import (
	"github.com/go-playground/validator/v10"
)

var SupportedPlatforms = []string{
	// Banks - International
	"Bank of America", "Chase", "Wells Fargo", "Citibank",
	"HSBC", "Barclays", "Deutsche Bank", "ING", "N26", "Revolut",
	// Brokerages
	"Interactive Brokers", "Fidelity", "Charles Schwab", "TD Ameritrade",
	"E*TRADE", "Robinhood", "Webull", "Trading 212", "eToro", "Degiro",
	// Crypto
	"Coinbase", "Binance", "Kraken", "Crypto.com", "Gemini", "Bybit", "OKX",
	// Payment Services
	"PayPal", "Wise", "Stripe", "Venmo", "Cash App",
	// Turkish Banks
	"Garanti BBVA", "İş Bankası", "Yapı Kredi", "Akbank",
	"QNB Finansbank", "Ziraat Bankası", "Midas",
	// Empty is allowed (optional field)
	"",
}

func ValidatePlatform(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	// Empty is valid (platform is optional)
	if value == "" {
		return true
	}

	for _, p := range SupportedPlatforms {
		if p == value {
			return true
		}
	}
	return false
}
