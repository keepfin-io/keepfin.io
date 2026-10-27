// Package authsession validates opaque session cookies with Better Auth.
package authsession

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var ErrUnauthorized = errors.New("invalid session")

type Identity struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
}

type Client struct {
	endpoint string
	http     *http.Client
}

func NewClient(serviceURL string) (*Client, error) {
	u, err := url.Parse(serviceURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") ||
		u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return nil, errors.New("AUTH_SERVICE_URL must be the internal auth service origin")
	}
	u.Path = "/api/auth/get-session"
	u.RawQuery = "disableCookieCache=true"
	return &Client{endpoint: u.String(), http: &http.Client{
		Timeout:       5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}, nil
}

// Verify forwards cookies only to the configured service; caller-supplied identity
// headers and bearer tokens never establish application identity.
func (c *Client) Verify(ctx context.Context, cookie string) (*Identity, []string, error) {
	if cookie == "" {
		return nil, nil, ErrUnauthorized
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Cookie", cookie)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("session service: %w", err)
	}
	defer res.Body.Close()
	cookies := res.Header.Values("Set-Cookie")
	if res.StatusCode == http.StatusUnauthorized {
		return nil, cookies, ErrUnauthorized
	}
	if res.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("session service status: %d", res.StatusCode)
	}
	var session *struct {
		User    Identity `json:"user"`
		Session struct {
			ID        string    `json:"id"`
			UserID    string    `json:"userId"`
			ExpiresAt time.Time `json:"expiresAt"`
		} `json:"session"`
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 65537))
	if err != nil || len(body) > 65536 {
		return nil, nil, errors.New("invalid session service response")
	}
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, nil, fmt.Errorf("decode session: %w", err)
	}
	if session == nil || session.Session.ID == "" || session.User.ID == "" || session.User.Email == "" ||
		!session.User.EmailVerified || session.Session.UserID != session.User.ID ||
		!session.Session.ExpiresAt.After(time.Now()) {
		return nil, cookies, ErrUnauthorized
	}
	return &session.User, cookies, nil
}
