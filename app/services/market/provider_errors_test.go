package market

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEODHDProviderErrors(t *testing.T) {
	for _, status := range []int{401, 402, 403, 429, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("api_token") != "private-key" {
					t.Error("missing provider authentication")
				}
				w.WriteHeader(status)
				_, _ = w.Write([]byte("private-key must not reach the user"))
			}))
			defer server.Close()
			client := &EODHDClient{BaseURL: server.URL, APIToken: "private-key", HTTPClient: server.Client()}
			_, err := client.SearchSymbol("QQ", 20)
			var provider *ProviderError
			if !errors.As(err, &provider) || provider.StatusCode != status {
				t.Fatalf("provider status lost: %v", err)
			}
			if strings.Contains(err.Error(), "private-key") {
				t.Fatal("upstream content leaked")
			}
		})
	}
	_, err := (&EODHDClient{}).SearchSymbol("QQ", 20)
	if !errors.Is(err, ErrProviderNotConfigured) {
		t.Fatalf("missing config: %v", err)
	}
}
