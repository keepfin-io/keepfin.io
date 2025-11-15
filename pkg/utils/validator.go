package utils

import (
	"encoding/json"
	"fmt"
	"keepnode/app/models"

	"github.com/go-playground/validator/v10"
)

// NewValidator func for create a new validator for model fields.
func NewValidator() *validator.Validate {
	// Create a new validator for a User model.
	validate := validator.New()

	validate.RegisterValidation("currency", models.ValidateCurrency)

	return validate
}

// ValidatorErrors func for show validation errors for each invalid fields.
func ValidatorErrors(err error) map[string]string {
	// Define fields map.
	fields := map[string]string{}

	// Make error message for each invalid field.
	for _, err := range err.(validator.ValidationErrors) {
		fields[err.Field()] = err.Error()
	}

	return fields
}

// ValidatorErrorsString converts validation errors to a JSON string
func ValidatorErrorsString(err error) string {
	fields := ValidatorErrors(err)

	// Convert map to JSON string
	jsonBytes, jsonErr := json.Marshal(fields)
	if jsonErr != nil {
		return fmt.Sprintf("validation error: %v", err)
	}

	return string(jsonBytes)
}
