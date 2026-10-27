package authsession

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestVerify(t *testing.T) {
	valid := fmt.Sprintf(`{"session":{"id":"s1","userId":"u1","expiresAt":%q},"user":{"id":"u1","email":"user@example.com","emailVerified":true}}`, time.Now().Add(time.Hour).Format(time.RFC3339))
	tests := []struct {
		name, body                string
		status                    int
		unauthorized, unavailable bool
	}{
		{"valid", valid, 200, false, false},
		{"revoked", "null", 200, true, false},
		{"expired", strings.Replace(valid, time.Now().Add(time.Hour).Format(time.RFC3339), "2000-01-01T00:00:00Z", 1), 200, true, false},
		{"unverified", strings.Replace(valid, "true", "false", 1), 200, true, false},
		{"mismatched identity", strings.Replace(valid, `"userId":"u1"`, `"userId":"u2"`, 1), 200, true, false},
		{"malformed", "<html>Bad Gateway</html>", 200, false, true},
		{"oversized", strings.Repeat(" ", 65537), 200, false, true},
		{"upstream down", "", 503, false, true},
		{"redirect rejected", "", 302, false, true},
		{"invalid", "", 401, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Cookie") != "keepfin.session_token=signed" {
					t.Error("cookie missing")
				}
				if r.URL.Path != "/api/auth/get-session" || r.URL.Query().Get("disableCookieCache") != "true" {
					t.Error("incorrect session endpoint")
				}
				w.Header().Add("Set-Cookie", "keepfin.session_token=renewed; HttpOnly; Path=/")
				w.Header().Add("Set-Cookie", "keepfin.session_data=; Max-Age=0; Path=/")
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.body)
			}))
			defer server.Close()
			client, err := NewClient(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			identity, cookies, err := client.Verify(context.Background(), "keepfin.session_token=signed")
			if errors.Is(err, ErrUnauthorized) != tt.unauthorized {
				t.Fatalf("unauthorized: %v", err)
			}
			if tt.unavailable && (err == nil || errors.Is(err, ErrUnauthorized)) {
				t.Fatalf("expected upstream error: %v", err)
			}
			if !tt.unavailable && !tt.unauthorized {
				if err != nil || identity.ID != "u1" || len(cookies) != 2 {
					t.Fatalf("identity/cookies: %v %v %v", identity, cookies, err)
				}
			}
		})
	}
}

func TestRejectEmptyCookieAndInvalidConfiguration(t *testing.T) {
	for _, input := range []string{"", "file:///etc/passwd", "http://user:secret@auth", "http://auth/path", "http://auth?url=evil"} {
		if _, err := NewClient(input); err == nil {
			t.Errorf("accepted %q", input)
		}
	}
	client, _ := NewClient("http://127.0.0.1:1")
	if _, _, err := client.Verify(context.Background(), ""); !errors.Is(err, ErrUnauthorized) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := client.Verify(ctx, "cookie=value"); err == nil || errors.Is(err, ErrUnauthorized) {
		t.Fatal("outage must fail closed")
	}
}
