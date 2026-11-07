package market

import (
	model "keepnode/app/models/market"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAlphaVantageRawHistoryFXSearchAndSafeErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("apikey") != "test-key" {
			t.Error("missing API key")
		}
		switch r.URL.Query().Get("function") {
		case "TIME_SERIES_DAILY":
			w.Write([]byte(`{"Time Series (Daily)":{"2026-01-02":{"4. close":"12.5"},"2026-01-01":{"4. close":"10"}}}`))
		case "FX_DAILY":
			if r.URL.Query().Get("from_symbol") != "EUR" || r.URL.Query().Get("to_symbol") != "USD" {
				t.Error("FX mapping")
			}
			w.Write([]byte(`{"Time Series FX (Daily)":{"2026-01-02":{"4. close":"1.1"}}}`))
		case "SYMBOL_SEARCH":
			w.Write([]byte(`{"bestMatches":[{"1. symbol":"IBM","2. name":"IBM","3. type":"Equity","4. region":"United States","8. currency":"USD"},{"1. symbol":"IBM.DE","4. region":"Germany"}]}`))
		default:
			w.Write([]byte(`{"Information":"Quota or subscription test-key secret"}`))
		}
	}))
	defer server.Close()
	client := &AlphaVantageClient{BaseURL: server.URL, APIKey: "test-key", HTTPClient: server.Client()}
	rows, err := client.History(&model.EODRequest{Symbol: "IBM.US", From: "2026-01-02", To: "2026-01-02"})
	if err != nil || len(rows) != 1 || rows[0].Close != 12.5 {
		t.Fatalf("history %+v %v", rows, err)
	}
	rows, err = client.History(&model.EODRequest{Symbol: "EURUSD.FOREX"})
	if err != nil || len(rows) != 1 || rows[0].Close != 1.1 {
		t.Fatalf("FX %+v %v", rows, err)
	}
	found, err := client.Search(&model.SearchRequest{Query: "IBM"})
	if err != nil || len(found) != 1 || found[0].Exchange != "US" {
		t.Fatalf("search %+v %v", found, err)
	}
	_, err = client.Quote(&model.LivePriceRequest{Symbol: "IBM.US"})
	if err == nil || strings.Contains(err.Error(), "test-key") {
		t.Fatalf("unsafe provider error %v", err)
	}
}
