package apiresponse

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
)

type APIError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string { return e.Message }

func NewError(status int, code, message string) *APIError {
	return &APIError{Status: status, Code: code, Message: message}
}

// ErrorHandler preserves intentional public errors and masks unexpected errors.
// Recovered panics and routing/middleware errors use this same envelope.
func ErrorHandler(c *fiber.Ctx, err error) error {
	public := NewError(fiber.StatusInternalServerError, "internal_error", "Internal server error")
	var apiErr *APIError
	var fiberErr *fiber.Error
	switch {
	case errors.As(err, &apiErr):
		public = apiErr
	case errors.As(err, &fiberErr):
		public = NewError(fiberErr.Code, statusCode(fiberErr.Code), fiberErr.Message)
	default:
		log.Printf("request_id=%s unhandled error: %v", RequestID(c), err)
	}
	return c.Status(public.Status).JSON(Envelope[any]{Data: nil, Meta: newMeta(c), Error: public})
}

func statusCode(status int) string {
	switch status {
	case fiber.StatusBadRequest:
		return "bad_request"
	case fiber.StatusUnauthorized:
		return "unauthorized"
	case fiber.StatusForbidden:
		return "forbidden"
	case fiber.StatusNotFound:
		return "not_found"
	case fiber.StatusMethodNotAllowed:
		return "method_not_allowed"
	case fiber.StatusConflict:
		return "conflict"
	case fiber.StatusRequestEntityTooLarge:
		return "payload_too_large"
	case fiber.StatusUnprocessableEntity:
		return "validation_failed"
	case fiber.StatusTooManyRequests:
		return "rate_limited"
	case fiber.StatusServiceUnavailable:
		return "service_unavailable"
	case fiber.StatusInternalServerError:
		return "internal_error"
	default:
		return "request_failed"
	}
}
