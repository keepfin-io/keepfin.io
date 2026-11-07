package market

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type Dividend struct {
	Date            string   `json:"date"`
	PaymentDate     string   `json:"paymentDate"`
	Currency        string   `json:"currency"`
	UnadjustedValue *float64 `json:"unadjustedValue"`
}

// EODHD's adjusted value is deliberately not used with historical share counts.
func (s *MarketService) Dividends(symbol, from, to string) ([]Dividend, error) {
	c := s.eodhd
	if c == nil || c.APIToken == "" {
		return nil, ErrProviderNotConfigured
	}
	params := url.Values{"api_token": {c.APIToken}, "fmt": {"json"}, "from": {from}, "to": {to}}
	response, err := c.HTTPClient.Get(c.BaseURL + "/div/" + url.PathEscape(symbol) + "?" + params.Encode())
	if err != nil {
		return nil, &ProviderError{}
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, &ProviderError{StatusCode: response.StatusCode}
	}
	rows := []Dividend{}
	if json.NewDecoder(io.LimitReader(response.Body, 2*1024*1024)).Decode(&rows) != nil {
		return nil, &ProviderError{StatusCode: 502}
	}
	return rows, nil
}
