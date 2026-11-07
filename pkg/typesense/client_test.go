package typesense

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestImportRequiresEveryDocumentAcknowledgement(t *testing.T) {
	for _, tc := range []struct {
		name, response string
		status         int
		wantError      bool
	}{
		{"success", "{\"success\":true}\n{\"success\":true}", 200, false},
		{"partial rejection", "{\"success\":true}\n{\"success\":false}", 200, true},
		{"truncated acknowledgement", "{\"success\":true}", 200, true},
		{"unavailable", "private-secret", 503, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-TYPESENSE-API-KEY") != "private-secret" || r.URL.Query().Get("action") != "upsert" {
					t.Error("missing authentication or idempotent upsert")
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.response))
			}))
			defer server.Close()
			client, err := New(server.URL, "private-secret")
			if err != nil {
				t.Fatal(err)
			}
			err = client.Import(context.Background(), "test", []json.RawMessage{json.RawMessage(`{"id":"1"}`), json.RawMessage(`{"id":"2"}`)})
			if (err != nil) != tc.wantError {
				t.Fatalf("error = %v", err)
			}
			if err != nil && strings.Contains(err.Error(), "private-secret") {
				t.Fatal("error exposed credentials")
			}
		})
	}
}
