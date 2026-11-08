package wealth

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"keepnode/pkg/apiresponse"
)

func TestRegisteredRoutesRequireUser(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: apiresponse.ErrorHandler})
	// Deliberately omit Session: WithUser must stop requests before any service call.
	New(nil).Register(app.Group("/api/v1"))
	for _, route := range app.GetRoutes(true) {
		if route.Method == fiber.MethodHead {
			continue
		}
		t.Run(route.Method+" "+route.Path, func(t *testing.T) {
			path := strings.ReplaceAll(route.Path, ":id", uuid.NewString())
			path = strings.ReplaceAll(path, ":kind", "activity")
			response, err := app.Test(httptest.NewRequest(route.Method, path, nil))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			var envelope apiresponse.Envelope[any]
			if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != 401 || envelope.Error == nil || envelope.Error.Code != "unauthorized" {
				t.Fatalf("route did not reject missing identity: %d %+v", response.StatusCode, envelope)
			}
		})
	}
}

func TestPublicReportRemainsAccessibleWithoutSession(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: apiresponse.ErrorHandler})
	app.Post("/api/public/portfolio", New(nil).PublicReport)
	// A malformed request must reach public request validation without requiring a session.
	req := httptest.NewRequest("POST", "/api/public/portfolio", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	response, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 400 {
		t.Fatalf("public report returned %d instead of request validation error", response.StatusCode)
	}
}
