package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"keepnode/app/models"
)

const (
	defaultBaseURL = "https://eodhd.com/api"
	defaultTimeout = 30 * time.Second
)

// EODHDClient handles HTTP requests to EODHD API
type EODHDClient struct {
	BaseURL    string
	APIToken   string
	HTTPClient *http.Client
}

// NewEODHDClient creates a new EODHD API client
func NewEODHDClient() *EODHDClient {
	apiToken := "695125bfe5ad16.63343807"
	if apiToken == "" {
		apiToken = " 695125bfe5ad16.63343807" // Demo token for testing (limited symbols)
	}

	return &EODHDClient{
		BaseURL:  defaultBaseURL,
		APIToken: apiToken,
		HTTPClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// GetEODData fetches End-of-Day historical data for a symbol
func (c *EODHDClient) GetEODData(symbol, exchange, from, to, period, order string) ([]models.EODData, error) {
	// Build symbol with exchange
	fullSymbol := symbol
	if exchange != "" {
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
		return nil, fmt.Errorf("failed to fetch EOD data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("EODHD API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var data []models.EODData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse EOD data: %w", err)
	}

	return data, nil
}

// GetLivePrice fetches real-time (delayed) price for a symbol
func (c *EODHDClient) GetLivePrice(symbol, exchange string) (*models.LivePrice, error) {
	// Build symbol with exchange
	fullSymbol := symbol
	if exchange != "" {
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
		return nil, fmt.Errorf("failed to fetch live price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("EODHD API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var data models.LivePrice
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse live price: %w", err)
	}

	return &data, nil
}

// GetMultipleLivePrices fetches real-time prices for multiple symbols
func (c *EODHDClient) GetMultipleLivePrices(symbols []string) ([]models.LivePrice, error) {
	if len(symbols) == 0 {
		return []models.LivePrice{}, nil
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
		return nil, fmt.Errorf("failed to fetch multiple live prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("EODHD API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response - can be single object or array
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Try parsing as array first
	var dataArray []models.LivePrice
	if err := json.Unmarshal(body, &dataArray); err == nil {
		return dataArray, nil
	}

	// Try parsing as single object
	var singleData models.LivePrice
	if err := json.Unmarshal(body, &singleData); err != nil {
		return nil, fmt.Errorf("failed to parse live prices: %w", err)
	}

	return []models.LivePrice{singleData}, nil
}

// SearchSymbol searches for stocks by query
func (c *EODHDClient) SearchSymbol(query string, limit int) ([]models.SearchResult, error) {
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
		return nil, fmt.Errorf("failed to search symbols: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("EODHD API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var results []models.SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	return results, nil
}
