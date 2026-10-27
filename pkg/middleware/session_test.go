package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"keepnode/app/models"
	"keepnode/app/services"
	"keepnode/pkg/authsession"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

func (f fakeResolver) ResolveIdentity(context.Context, *authsession.Identity) (*models.User, error) {
	return &models.User{ID: f.id}, f.err
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
		{"blocked", "signed=value", nil, services.ErrUserBlocked, 403},
		{"legacy collision", "signed=value", nil, services.ErrIdentityConflict, 409},
	} {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", Session(fakeVerifier{tt.upstream}, fakeResolver{id, tt.resolve}), func(c *fiber.Ctx) error {
				if c.Locals("user_id") != id {
					t.Error("did not use mapped application UUID")
				}
				return c.SendStatus(200)
			})
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
