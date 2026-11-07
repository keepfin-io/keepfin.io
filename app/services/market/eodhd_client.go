package market

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	marketmodel "keepnode/app/models/market"
)

const (
	defaultBaseURL = "https://eodhd.com/api"
	defaultTimeout = 8 * time.Second
)

// EODHDClient handles HTTP requests to EODHD API
type EODHDClient struct {
	BaseURL    string
	APIToken   string
	HTTPClient *http.Client
}

// NewEODHDClient creates a new EODHD API client
func NewEODHDClient() *EODHDClient {
	apiToken := os.Getenv("EODHD_API_TOKEN")

	return &EODHDClient{
		BaseURL:  defaultBaseURL,
		APIToken: apiToken,
		HTTPClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// GetEODData fetches End-of-Day historical data for a symbol
func (c *EODHDClient) GetEODData(symbol, exchange, from, to, period, order string) ([]marketmodel.EODData, error) {
	if c.APIToken == "" {
		return nil, ErrProviderNotConfigured
	}
	// Build symbol with exchange
	fullSymbol := symbol
	if exchange != "" && !strings.Contains(symbol, ".") {
		fullSymbol = fmt.Sprintf("%s.%s", symbol, exchange)
	}

	// Build URL with query parameters
	endpoint := fmt.Sprintf("%s/eod/%s", c.BaseURL, fullSymbol)
	params := url.Values{}
	params.Set("api_token", c.APIToken)
	params.Set("fmt", "json")

	if from != "" {
		params.Set("from", from)
	}
	if to != "" {
		params.Set("to", to)
	}
	if period != "" {
		params.Set("period", period)
	}
	if order != "" {
		params.Set("order", order)
	}

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	// Make request
	resp, err := c.HTTPClient.Get(fullURL)
	if err != nil {
		return nil, &ProviderError{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &ProviderError{StatusCode: resp.StatusCode}
	}

	// Parse response
	var data []marketmodel.EODData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, &ProviderError{StatusCode: http.StatusBadGateway}
	}

	return data, nil
}

// GetLivePrice fetches real-time (delayed) price for a symbol
func (c *EODHDClient) GetLivePrice(symbol, exchange string) (*marketmodel.LivePrice, error) {
	if c.APIToken == "" {
		return nil, ErrProviderNotConfigured
	}
	// Build symbol with exchange
	fullSymbol := symbol
	if exchange != "" && !strings.Contains(symbol, ".") {
		fullSymbol = fmt.Sprintf("%s.%s", symbol, exchange)
	}

	// Build URL
	endpoint := fmt.Sprintf("%s/real-time/%s", c.BaseURL, fullSymbol)
	params := url.Values{}
	params.Set("api_token", c.APIToken)
	params.Set("fmt", "json")

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	// Make request
	resp, err := c.HTTPClient.Get(fullURL)
	if err != nil {
		return nil, &ProviderError{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &ProviderError{StatusCode: resp.StatusCode}
	}

	// Parse response
	var data marketmodel.LivePrice
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, &ProviderError{StatusCode: http.StatusBadGateway}
	}

	return &data, nil
}

// GetMultipleLivePrices fetches real-time prices for multiple symbols
func (c *EODHDClient) GetMultipleLivePrices(symbols []string) ([]marketmodel.LivePrice, error) {
	if c.APIToken == "" {
		return nil, ErrProviderNotConfigured
	}
	if len(symbols) == 0 {
		return []marketmodel.LivePrice{}, nil
	}

	// First symbol goes in path, others in query param
	endpoint := fmt.Sprintf("%s/real-time/%s", c.BaseURL, symbols[0])
	params := url.Values{}
	params.Set("api_token", c.APIToken)
	params.Set("fmt", "json")

	if len(symbols) > 1 {
		// Additional symbols as comma-separated
		additionalSymbols := ""
		for i := 1; i < len(symbols); i++ {
			if i > 1 {
				additionalSymbols += ","
			}
			additionalSymbols += symbols[i]
		}
		params.Set("s", additionalSymbols)
	}

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	// Make request
	resp, err := c.HTTPClient.Get(fullURL)
	if err != nil {
		return nil, &ProviderError{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &ProviderError{StatusCode: resp.StatusCode}
	}

	// Parse response - can be single object or array
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ProviderError{StatusCode: http.StatusBadGateway}
	}

	// Try parsing as array first
	var dataArray []marketmodel.LivePrice
	if err := json.Unmarshal(body, &dataArray); err == nil {
		return dataArray, nil
	}

	// Try parsing as single object
	var singleData marketmodel.LivePrice
	if err := json.Unmarshal(body, &singleData); err != nil {
		return nil, &ProviderError{StatusCode: http.StatusBadGateway}
	}

	return []marketmodel.LivePrice{singleData}, nil
}

// SearchSymbol searches for stocks by query
func (c *EODHDClient) SearchSymbol(query string, limit int) ([]marketmodel.SearchResult, error) {
	if c.APIToken == "" {
		return nil, ErrProviderNotConfigured
	}
	// Build URL
	endpoint := fmt.Sprintf("%s/search/%s", c.BaseURL, url.PathEscape(query))
	params := url.Values{}
	params.Set("api_token", c.APIToken)

	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	// Make request
	resp, err := c.HTTPClient.Get(fullURL)
	if err != nil {
		return nil, &ProviderError{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &ProviderError{StatusCode: resp.StatusCode}
	}

	// Parse response
	var results []marketmodel.SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, &ProviderError{StatusCode: http.StatusBadGateway}
	}

	return results, nil
}
