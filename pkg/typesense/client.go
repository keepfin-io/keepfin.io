// Package typesense provides the small HTTP surface needed by the holdings index.
package typesense

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

type HTTPError struct{ Status int }

func (e *HTTPError) Error() string { return fmt.Sprintf("search service returned status %d", e.Status) }

func IsStatus(err error, status int) bool {
	var target *HTTPError
	return errors.As(err, &target) && target.Status == status
}

func New(baseURL, apiKey string) (*Client, error) {
	if baseURL == "" && apiKey == "" {
		return nil, nil
	}
	u, err := url.Parse(baseURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") ||
		u.User != nil || u.RawQuery != "" || u.Fragment != "" || apiKey == "" {
		return nil, errors.New("TYPESENSE_URL and TYPESENSE_API_KEY must configure an HTTP(S) search server")
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), apiKey: apiKey, http: &http.Client{
		Timeout:       2 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}}, nil
}

func (c *Client) Request(ctx context.Context, method, path string, query url.Values, body []byte) ([]byte, error) {
	address := c.baseURL + path
	if len(query) > 0 {
		address += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, address, bytes.NewReader(body))
	if err != nil {
		return nil, errors.New("invalid search request")
	}
	req.Header.Set("X-TYPESENSE-API-KEY", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	response, err := c.http.Do(req)
	if err != nil {
		return nil, errors.New("search service is unavailable")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 65536))
		return nil, &HTTPError{Status: response.StatusCode}
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return nil, errors.New("cannot read search response")
	}
	return data, nil
}

// Import checks each JSONL result: HTTP 200 does not mean every document succeeded.
func (c *Client) Import(ctx context.Context, collection string, rows []json.RawMessage) error {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	for _, row := range rows {
		if err := encoder.Encode(row); err != nil {
			return err
		}
	}
	data, err := c.Request(ctx, http.MethodPost, "/collections/"+collection+"/documents/import",
		url.Values{"action": {"upsert"}, "dirty_values": {"reject"}}, buffer.Bytes())
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	for range len(rows) {
		var result struct {
			Success bool `json:"success"`
		}
		if err := decoder.Decode(&result); err != nil || !result.Success {
			return errors.New("search index rejected a document")
		}
	}
	return nil
}
