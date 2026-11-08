package authhandler_test

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"keepnode/pkg/apiresponse"
	"keepnode/pkg/authhandler"
)

func TestWithUser(t *testing.T) {
	userID := uuid.New()
	forgedID := uuid.NewString()
	for _, tt := range []struct {
		name       string
		identity   any
		handlerErr error
		status     int
		code       string
	}{
		{name: "missing identity", status: 401, code: "unauthorized"},
		{name: "UUID string is not a verified identity", identity: userID.String(), status: 401, code: "unauthorized"},
		{name: "wrong identity type", identity: 42, status: 401, code: "unauthorized"},
		{name: "zero UUID", identity: uuid.Nil, status: 401, code: "unauthorized"},
		{name: "verified application identity", identity: userID, status: 200},
		{
			name: "controller errors reach the envelope handler", identity: userID,
			handlerErr: fiber.NewError(404, "Account not found"), status: 404, code: "not_found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New(fiber.Config{ErrorHandler: apiresponse.ErrorHandler})
			app.Use(func(c *fiber.Ctx) error {
				if tt.identity != nil {
					c.Locals("user_id", tt.identity)
				}
				return c.Next()
			})
			calls := 0
			app.Post("/", authhandler.WithUser(func(c *fiber.Ctx, id uuid.UUID) error {
				calls++
				if id != userID {
					t.Errorf("got identity %s, want %s", id, userID)
				}
				if tt.handlerErr != nil {
					return tt.handlerErr
				}
				return apiresponse.RespondOK(c, id)
			}))
			req := httptest.NewRequest("POST", "/?user_id="+forgedID, strings.NewReader(`{"user_id":"`+forgedID+`"}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-User-ID", forgedID)
			req.Header.Set("Authorization", "Bearer "+forgedID)
			response, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != tt.status {
				t.Fatalf("got status %d, want %d", response.StatusCode, tt.status)
			}
			var envelope apiresponse.Envelope[*uuid.UUID]
			if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
				t.Fatal(err)
			}
			if envelope.Meta.RequestID == "" || envelope.Meta.RequestID != response.Header.Get("X-Request-ID") {
				t.Fatal("missing or inconsistent request identity")
			}
			if tt.code != "" {
				if envelope.Data != nil || envelope.Error == nil || envelope.Error.Code != tt.code {
					t.Fatalf("invalid error envelope: %+v", envelope)
				}
			} else if envelope.Error != nil || envelope.Data == nil || *envelope.Data != userID {
				t.Fatalf("invalid success envelope: %+v", envelope)
			}
			wantCalls := 1
			if tt.status == 401 {
				wantCalls = 0
			}
			if calls != wantCalls {
				t.Fatalf("controller called %d times, want %d", calls, wantCalls)
			}
		})
	}
}

func TestWithUserDoesNotRetainIdentityBetweenRequests(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: apiresponse.ErrorHandler})
	firstID, secondID := uuid.New(), uuid.New()
	identities := []uuid.UUID{firstID, secondID}
	app.Use(func(c *fiber.Ctx) error {
		if len(identities) > 0 {
			c.Locals("user_id", identities[0])
			identities = identities[1:]
		}
		return c.Next()
	})
	app.Get("/", authhandler.WithUser(func(c *fiber.Ctx, id uuid.UUID) error {
		return apiresponse.RespondOK(c, id)
	}))
	for _, wantID := range []uuid.UUID{firstID, secondID, uuid.Nil} {
		response, err := app.Test(httptest.NewRequest("GET", "/", nil))
		if err != nil {
			t.Fatal(err)
		}
		var envelope apiresponse.Envelope[*uuid.UUID]
		err = json.NewDecoder(response.Body).Decode(&envelope)
		response.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if wantID == uuid.Nil {
			if response.StatusCode != 401 || envelope.Data != nil {
				t.Fatal("previous request's identity leaked into a guest request")
			}
			continue
		}
		if response.StatusCode != 200 || envelope.Data == nil || *envelope.Data != wantID {
			t.Fatalf("response used another request's identity: %+v", envelope)
		}
	}
}
