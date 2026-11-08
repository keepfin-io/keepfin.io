package middleware

import (
	"context"
	"errors"
	"net/url"

	"github.com/gofiber/fiber/v2"

	usermodel "keepnode/app/models/user"
	userservice "keepnode/app/services/user"
	"keepnode/pkg/authsession"
)

type sessionVerifier interface {
	Verify(context.Context, string) (*authsession.Identity, []string, error)
}

type identityResolver interface {
	ResolveIdentity(context.Context, *authsession.Identity) (*usermodel.User, error)
}

func Session(client sessionVerifier, users identityResolver) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-store")
		identity, cookies, err := client.Verify(authsession.WithClientIP(c.UserContext(), ClientIP(c)), c.Get("Cookie"))
		for _, cookie := range cookies {
			c.Response().Header.Add("Set-Cookie", cookie)
		}
		if errors.Is(err, authsession.ErrUnauthorized) {
			return fiber.NewError(401, "Authentication required")
		}
		var rateLimit *authsession.RateLimitError
		if errors.As(err, &rateLimit) {
			c.Set("Retry-After", rateLimit.RetryAfter)
			return fiber.NewError(fiber.StatusTooManyRequests, "Too many requests; try again later")
		}
		if err != nil {
			return fiber.NewError(503, "Authentication service unavailable")
		}
		user, err := users.ResolveIdentity(c.UserContext(), identity)
		if errors.Is(err, userservice.ErrIdentityConflict) {
			return fiber.NewError(409, "Existing account requires migration. Contact support.")
		}
		if errors.Is(err, userservice.ErrUserBlocked) {
			return fiber.NewError(403, "Account unavailable")
		}
		if err != nil {
			return fiber.NewError(503, "Unable to load account")
		}
		c.Locals("user_id", user.ID)
		return c.Next()
	}
}

// SameOrigin protects cookie-authenticated Go mutations independently of Better
// Auth's own CSRF checks. API clients must send the configured browser Origin.
func SameOrigin(origin string) (fiber.Handler, error) {
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") ||
		u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("APP_ORIGIN must be an explicit origin without a trailing slash")
	}
	return func(c *fiber.Ctx) error {
		switch c.Method() {
		case fiber.MethodGet, fiber.MethodHead, fiber.MethodOptions:
			return c.Next()
		}
		if c.Get("Origin") != origin || c.Get("Sec-Fetch-Site") == "cross-site" {
			return fiber.NewError(403, "Invalid request origin")
		}
		return c.Next()
	}, nil
}
