package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	usermodel "keepnode/app/models/user"
	userservice "keepnode/app/services/user"
	"keepnode/pkg/apiresponse"
	"keepnode/pkg/authhandler"
	"keepnode/pkg/authsession"
)

type fakeVerifier struct{ err error }

func (f fakeVerifier) Verify(_ context.Context, cookie string) (*authsession.Identity, []string, error) {
	if cookie == "" {
		return nil, nil, authsession.ErrUnauthorized
	}
	return &authsession.Identity{ID: "auth-id"}, []string{"one=1; HttpOnly", "two=2; HttpOnly"}, f.err
}

type fakeResolver struct {
	id  uuid.UUID
	err error
}

func (f fakeResolver) ResolveIdentity(context.Context, *authsession.Identity) (*usermodel.User, error) {
	return &usermodel.User{ID: f.id}, f.err
}

func TestSessionBoundary(t *testing.T) {
	id := uuid.New()
	for _, tt := range []struct {
		name, cookie      string
		upstream, resolve error
		status            int
	}{
		{"valid", "signed=value", nil, nil, 200},
		{"bearer is not identity", "", nil, nil, 401},
		{"revoked", "signed=value", authsession.ErrUnauthorized, nil, 401},
		{"outage", "signed=value", errors.New("offline"), nil, 503},
		{"rate limited", "signed=value", &authsession.RateLimitError{RetryAfter: "30"}, nil, 429},
		{"blocked", "signed=value", nil, userservice.ErrUserBlocked, 403},
		{"legacy collision", "signed=value", nil, userservice.ErrIdentityConflict, 409},
	} {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New(fiber.Config{ErrorHandler: apiresponse.ErrorHandler})
			called := false
			handler := authhandler.WithUser(func(c *fiber.Ctx, userID uuid.UUID) error {
				called = true
				if userID != id {
					t.Error("did not use mapped application UUID")
				}
				return apiresponse.RespondOK(c, userID)
			})
			app.Get("/", Session(fakeVerifier{err: tt.upstream}, fakeResolver{id: id, err: tt.resolve}), handler)
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Cookie", tt.cookie)
			req.Header.Set("Authorization", "Bearer legacy-forged-token")
			req.Header.Set("X-User-ID", uuid.NewString())
			res, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != tt.status {
				t.Fatalf("got %d want %d", res.StatusCode, tt.status)
			}
			if tt.status == 429 && res.Header.Get("Retry-After") != "30" {
				t.Fatal("lost upstream retry delay")
			}
			if called != (tt.status == 200) {
				t.Fatalf("controller called=%t for status %d", called, tt.status)
			}
			var envelope apiresponse.Envelope[*uuid.UUID]
			if err := json.NewDecoder(res.Body).Decode(&envelope); err != nil {
				t.Fatal(err)
			}
			if tt.status == 200 {
				if envelope.Error != nil || envelope.Data == nil || *envelope.Data != id {
					t.Fatalf("invalid authenticated response: %+v", envelope)
				}
			} else if envelope.Data != nil || envelope.Error == nil {
				t.Fatalf("invalid error envelope: %+v", envelope)
			}
			if tt.status == 200 && len(res.Header.Values("Set-Cookie")) != 2 {
				t.Fatal("lost renewed cookies")
			}
		})
	}
}

func TestSameOrigin(t *testing.T) {
	handler, err := SameOrigin("http://localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	app.Use(handler)
	app.All("/", func(c *fiber.Ctx) error { return c.SendStatus(200) })
	for _, tt := range []struct {
		method, origin, site string
		status               int
	}{
		{"GET", "", "", 200},
		{"POST", "http://localhost:8080", "same-origin", 200},
		{"PUT", "", "", 403},
		{"PATCH", "https://evil.test", "cross-site", 403},
		{"DELETE", "http://localhost:8080", "cross-site", 403},
		{"POST", "null", "", 403},
	} {
		req := httptest.NewRequest(tt.method, "/", nil)
		req.Header.Set("Origin", tt.origin)
		req.Header.Set("Sec-Fetch-Site", tt.site)
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != tt.status {
			t.Errorf("%s %s: got %d want %d", tt.method, tt.origin, res.StatusCode, tt.status)
		}
	}
}
