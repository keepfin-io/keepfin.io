// Package apiresponse defines the JSON contract shared by Keepfin API endpoints.
package apiresponse

import "github.com/gofiber/fiber/v2"

type Envelope[T any] struct {
	Data  T         `json:"data"`
	Meta  Meta      `json:"meta"`
	Error *APIError `json:"error"`
}

type Option func(*Meta)

// RespondOK writes a JSON envelope with HTTP 200 OK.
func RespondOK[T any](c *fiber.Ctx, data T, options ...Option) error {
	return writeJSONResponse(
		c,
		fiber.StatusOK,
		data,
		options,
	)
}

// RespondCreated writes a JSON envelope with HTTP 201 Created.
func RespondCreated[T any](c *fiber.Ctx, data T, options ...Option) error {
	return writeJSONResponse(
		c,
		fiber.StatusCreated,
		data,
		options,
	)
}

func writeJSONResponse[T any](c *fiber.Ctx, status int, data T, options []Option) error {
	meta := newMeta(c)
	for _, option := range options {
		option(&meta)
	}
	return c.Status(status).JSON(Envelope[T]{Data: data, Meta: meta})
}
