package models

import "github.com/go-playground/validator/v10"

var SupportedCurrencies = []string{
	// Major Currencies
	"USD", "EUR", "GBP", "JPY", "CHF",
	"CAD", "AUD", "NZD",
	// Asian Currencies
	"CNY", "JPY", "KRW", "SGD", "HKD", "INR", "THB", "MYR", "IDR", "PHP", "VND",
	// European Currencies
	"SEK", "NOK", "DKK", "PLN", "CZK", "HUF", "RON", "BGN", "HRK", "RUB", "UAH",
	// Middle Eastern Currencies
	"TRY", "SAR", "AED", "QAR", "KWD", "BHD", "OMR", "ILS", "EGP",
	// Americas
	"BRL", "MXN", "ARS", "CLP", "COP", "PEN",
	// African Currencies
	"ZAR", "NGN", "KES", "MAD",
	// Crypto
	"BTC", "ETH",
}

func ValidateCurrency(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	for _, c := range SupportedCurrencies {
		if c == value {
			return true
		}
	}
	return false
}
