package middleware

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"keepnode/pkg/apiresponse"
)

func TestClientIPTrustBoundary(t *testing.T) {
	resolver, err := NewClientIPResolver("192.0.2.10/32,2001:db8:1::/64")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ peer, header, want string }{
		{"192.0.2.10", "198.51.100.1", "198.51.100.1"},
		{"192.0.2.11", "198.51.100.1", "192.0.2.11"},
		{"192.0.2.10", "198.51.100.1, 198.51.100.2", "192.0.2.10"},
		{"::ffff:192.0.2.10", "::ffff:198.51.100.1", "198.51.100.1"},
		{"2001:db8:1::1", "2001:db8:2::4", "2001:db8:2::4"},
	} {
		if got := resolver.Resolve(tc.peer, tc.header); got != tc.want {
			t.Errorf("%+v: got %s", tc, got)
		}
	}
	if _, err := NewClientIPResolver("trust-everyone"); err == nil {
		t.Fatal("invalid proxy configuration accepted")
	}
}

func TestIPLimitStopsRequestsBeforeSessionVerification(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: apiresponse.ErrorHandler})
	calls := 0
	app.Use(RequestContext(), IPRequestLimit(2))
	app.Get("/", func(c *fiber.Ctx) error { calls++; return fiber.NewError(401, "Authentication required") })
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Real-IP", "198.51.100."+string(rune('1'+i)))
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		if i == 2 && (res.StatusCode != 429 || res.Header.Get("Retry-After") == "" || !strings.Contains(string(body), `"meta"`)) {
			t.Fatalf("missing limit/envelope: %d %s", res.StatusCode, body)
		}
	}
	if calls != 2 {
		t.Fatalf("verification called %d times", calls)
	}
}

func TestUserLimitSeparatesUsersAcrossIPs(t *testing.T) {
	app := fiber.New()
	users := []uuid.UUID{uuid.New(), uuid.New()}
	app.Use(func(c *fiber.Ctx) error {
		id := users[0]
		if c.Get("Test-User") == "second" {
			id = users[1]
		}
		c.Locals("user_id", id)
		return c.Next()
	}, UserRequestLimit(1))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(200) })
	for _, tc := range []struct {
		user string
		want int
	}{{"first", 200}, {"first", 429}, {"second", 200}} {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Test-User", tc.user)
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != tc.want {
			t.Fatalf("%s: %d", tc.user, res.StatusCode)
		}
	}
}
