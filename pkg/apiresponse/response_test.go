package apiresponse_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"

	"keepnode/pkg/apiresponse"
	"keepnode/pkg/middleware"
)

func TestResponseContract(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: apiresponse.ErrorHandler})
	app.Use(middleware.RequestContext(), recover.New())
	app.Get("/report", func(c *fiber.Ctx) error {
		return apiresponse.RespondOK(c, fiber.Map{"value": 1990, "warnings": []string{}})
	})
	app.Post("/accounts", func(c *fiber.Ctx) error {
		return apiresponse.RespondCreated(c, fiber.Map{"id": "account"})
	})
	app.Get("/empty", func(c *fiber.Ctx) error {
		return apiresponse.RespondOK(c, []string{}, apiresponse.WithPagination(1, 10, 0))
	})
	app.Get("/page", func(c *fiber.Ctx) error {
		return apiresponse.RespondOK(c, []string{"asset"}, apiresponse.WithPagination(3, 10, 21))
	})
	app.Get("/denied", func(c *fiber.Ctx) error {
		return fiber.NewError(401, "Authentication required")
	})
	app.Get("/coded", func(c *fiber.Ctx) error {
		return fmt.Errorf("wrapped: %w", apiresponse.NewError(404, "account_not_found", "Account not found"))
	})
	app.Get("/internal", func(c *fiber.Ctx) error { return errors.New("private database credential") })
	app.Get("/panic", func(c *fiber.Ctx) error { panic("private panic detail") })
	app.Get("/export", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		c.Set("Content-Disposition", "attachment; filename=activities.json")
		return c.SendString(`{"version":1,"activities":[]}`)
	})

	seen := map[string]bool{}
	for _, tc := range []struct {
		path, method, data, code string
		status                   int
		pages                    int
	}{
		{path: "/report", method: "GET", status: 200, data: `{"value":1990,"warnings":[]}`},
		{path: "/accounts", method: "POST", status: 201, data: `{"id":"account"}`},
		{path: "/empty", method: "GET", status: 200, data: `[]`, pages: 1},
		{path: "/page", method: "GET", status: 200, data: `["asset"]`, pages: 3},
		{path: "/denied", method: "GET", status: 401, data: `null`, code: "unauthorized"},
		{path: "/coded", method: "GET", status: 404, data: `null`, code: "account_not_found"},
		{path: "/missing", method: "GET", status: 404, data: `null`, code: "not_found"},
		{path: "/internal", method: "GET", status: 500, data: `null`, code: "internal_error"},
		{path: "/panic", method: "GET", status: 500, data: `null`, code: "internal_error"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, tc.path, nil)
			request.Header.Set("X-Request-ID", "untrusted-client-value")
			response, err := app.Test(request)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			var body struct {
				Data  json.RawMessage       `json:"data"`
				Meta  apiresponse.Meta      `json:"meta"`
				Error *apiresponse.APIError `json:"error"`
			}
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != tc.status || string(body.Data) != tc.data {
				t.Fatalf("status=%d body=%+v", response.StatusCode, body)
			}
			id := body.Meta.RequestID
			if _, err := uuid.Parse(id); err != nil || seen[id] || response.Header.Get("X-Request-ID") != id {
				t.Fatalf("invalid request identity: %q", id)
			}
			seen[id] = true
			if body.Meta.Timestamp.IsZero() || time.Since(body.Meta.Timestamp) > time.Minute {
				t.Fatalf("invalid timestamp: %v", body.Meta.Timestamp)
			}
			if tc.code == "" && body.Error != nil || tc.code != "" && (body.Error == nil || body.Error.Code != tc.code) {
				t.Fatalf("unexpected error: %+v", body.Error)
			}
			if tc.status == 500 && strings.Contains(body.Error.Message, "private") {
				t.Fatal("internal details leaked")
			}
			if tc.pages == 0 && body.Meta.Pagination != nil || tc.pages > 0 && (body.Meta.Pagination == nil || body.Meta.Pagination.TotalPages != tc.pages) {
				t.Fatalf("unexpected pagination: %+v", body.Meta.Pagination)
			}
		})
	}
	response, err := app.Test(httptest.NewRequest("GET", "/export", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"version":1,"activities":[]}` || !strings.Contains(response.Header.Get("Content-Disposition"), "attachment") {
		t.Fatalf("export was wrapped: %s", body)
	}
}
