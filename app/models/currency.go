package models

import "github.com/go-playground/validator/v10"

var SupportedCurrencies = []string{
	"USD", "EUR", "TRY",
	"GBP", "CHF", "JPY",
	"AUD", "CAD", "NZD",
	"SEK", "NOK", "DKK",
	"PLN", "CZK", "HUF",
	"SAR", "AED", "QAR",
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
