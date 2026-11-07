package middleware

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type testIndexFlusher struct {
	called int
	user   uuid.UUID
	err    error
}

func (f *testIndexFlusher) FlushUser(_ context.Context, user uuid.UUID) error {
	f.called++
	f.user = user
	return f.err
}

func TestHoldingSearchFailurePreservesCommittedResponse(t *testing.T) {
	user := uuid.New()
	index := &testIndexFlusher{err: errors.New("index unavailable")}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error { c.Locals("user_id", user); return c.Next() }, HoldingSearchSync(index))
	app.Post("/holdings", func(c *fiber.Ctx) error { return c.Status(201).SendString("committed") })
	app.Get("/holdings", func(c *fiber.Ctx) error { return c.SendStatus(200) })
	app.Delete("/holdings", func(c *fiber.Ctx) error { return c.SendStatus(400) })
	for _, method := range []string{"POST", "GET", "DELETE"} {
		response, err := app.Test(httptest.NewRequest(method, "/holdings", nil))
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(response.Body)
		response.Body.Close()
		if method == "POST" && (response.StatusCode != 201 || string(body) != "committed" || response.Header.Get("X-Holdings-Search-Sync") != "pending") {
			t.Fatalf("committed response changed: %d %s", response.StatusCode, body)
		}
	}
	if index.called != 1 || index.user != user {
		t.Fatal("flush must only run for the authenticated successful mutation")
	}
}
