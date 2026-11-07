package market

import (
	"encoding/json"
	"errors"
	"io"
	model "keepnode/app/models/market"
	"keepnode/pkg/portfolio"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Alpha Vantage currently supports US instruments and daily FX in Keepfin.
// Its default quote is a closing quote; no real-time entitlement is assumed.
type AlphaVantageClient struct {
	BaseURL, APIKey string
	HTTPClient      *http.Client
}

func NewAlphaVantageClient() *AlphaVantageClient {
	return &AlphaVantageClient{"https://www.alphavantage.co/query", os.Getenv("ALPHA_VANTAGE_API_KEY"), &http.Client{Timeout: defaultTimeout}}
}
func (c *AlphaVantageClient) request(params url.Values, target any) error {
	if c.APIKey == "" {
		return ErrProviderNotConfigured
	}
	params.Set("apikey", c.APIKey)
	response, err := c.HTTPClient.Get(c.BaseURL + "?" + params.Encode())
	if err != nil {
		return &ProviderError{}
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return &ProviderError{StatusCode: response.StatusCode}
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 16*1024*1024))
	if err != nil {
		return &ProviderError{}
	}
	var meta map[string]json.RawMessage
	if json.Unmarshal(body, &meta) != nil {
		return &ProviderError{StatusCode: 502}
	}
	if _, ok := meta["Note"]; ok {
		return &ProviderError{StatusCode: 429}
	}
	if _, ok := meta["Information"]; ok {
		return &ProviderError{StatusCode: 403}
	}
	if _, ok := meta["Error Message"]; ok {
		return &ProviderError{StatusCode: 502}
	}
	if json.Unmarshal(body, target) != nil {
		return &ProviderError{StatusCode: 502}
	}
	return nil
}
func alphaSymbol(symbol, exchange string) (string, string, error) {
	if i := strings.LastIndex(symbol, "."); i >= 0 {
		exchange = symbol[i+1:]
		symbol = symbol[:i]
	}
	if exchange == "" {
		exchange = "US"
	}
	if exchange != "US" && exchange != "FOREX" {
		return "", "", errors.New("Alpha Vantage supports US instruments and FX in Keepfin")
	}
	return symbol, exchange, nil
}
func (c *AlphaVantageClient) History(req *model.EODRequest) ([]model.EODData, error) {
	symbol, exchange, err := alphaSymbol(req.Symbol, req.Exchange)
	if err != nil {
		return nil, err
	}
	if req.Period != "" && req.Period != "d" {
		return nil, errors.New("Alpha Vantage history supports daily prices")
	}
	params := url.Values{"function": {"TIME_SERIES_DAILY"}, "symbol": {symbol}, "outputsize": {"full"}}
	key := "Time Series (Daily)"
	if exchange == "FOREX" {
		if len(symbol) != 6 {
			return nil, errors.New("invalid FX pair")
		}
		params.Del("symbol")
		params.Set("function", "FX_DAILY")
		params.Set("from_symbol", symbol[:3])
		params.Set("to_symbol", symbol[3:])
		key = "Time Series FX (Daily)"
	}
	var payload map[string]json.RawMessage
	if err = c.request(params, &payload); err != nil {
		return nil, err
	}
	var series map[string]map[string]string
	if json.Unmarshal(payload[key], &series) != nil {
		return nil, &ProviderError{StatusCode: 502}
	}
	result := []model.EODData{}
	for date, row := range series {
		if _, e := time.Parse("2006-01-02", date); e != nil {
			return nil, &ProviderError{StatusCode: 502}
		}
		if req.From != "" && date < req.From || req.To != "" && date > req.To {
			continue
		}
		close, e := strconv.ParseFloat(row["4. close"], 64)
		if e != nil || !portfolio.ValidNumber(close) || close <= 0 {
			return nil, &ProviderError{StatusCode: 502}
		}
		open, _ := strconv.ParseFloat(row["1. open"], 64)
		high, _ := strconv.ParseFloat(row["2. high"], 64)
		low, _ := strconv.ParseFloat(row["3. low"], 64)
		volume, _ := strconv.ParseInt(row["5. volume"], 10, 64)
		result = append(result, model.EODData{Date: date, Open: open, High: high, Low: low, Close: close, Volume: volume})
	}
	sort.Slice(result, func(i, j int) bool {
		if req.Order == "d" {
			return result[i].Date > result[j].Date
		}
		return result[i].Date < result[j].Date
	})
	return result, nil
}
func (c *AlphaVantageClient) Quote(req *model.LivePriceRequest) (*model.LivePrice, error) {
	symbol, exchange, err := alphaSymbol(req.Symbol, req.Exchange)
	if err != nil {
		return nil, err
	}
	if exchange != "US" {
		return nil, errors.New("use daily history for FX rates")
	}
	var payload struct {
		Quote map[string]string `json:"Global Quote"`
	}
	if err = c.request(url.Values{"function": {"GLOBAL_QUOTE"}, "symbol": {symbol}}, &payload); err != nil {
		return nil, err
	}
	row := payload.Quote
	close, e := strconv.ParseFloat(row["05. price"], 64)
	if e != nil || !portfolio.ValidNumber(close) || close <= 0 {
		return nil, &ProviderError{StatusCode: 502}
	}
	day, e := time.Parse("2006-01-02", row["07. latest trading day"])
	if e != nil {
		return nil, &ProviderError{StatusCode: 502}
	}
	prev, _ := strconv.ParseFloat(row["08. previous close"], 64)
	change, _ := strconv.ParseFloat(row["09. change"], 64)
	percent, _ := strconv.ParseFloat(strings.TrimSuffix(row["10. change percent"], "%"), 64)
	return &model.LivePrice{Code: symbol + ".US", Close: close, PreviousClose: prev, Change: change, ChangeP: percent, Timestamp: day.Unix()}, nil
}
func (c *AlphaVantageClient) Search(req *model.SearchRequest) ([]model.SearchResult, error) {
	var payload struct {
		Matches []map[string]string `json:"bestMatches"`
	}
	if err := c.request(url.Values{"function": {"SYMBOL_SEARCH"}, "keywords": {req.Query}}, &payload); err != nil {
		return nil, err
	}
	out := []model.SearchResult{}
	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	for _, m := range payload.Matches {
		if m["4. region"] != "United States" {
			continue
		}
		out = append(out, model.SearchResult{Code: m["1. symbol"], Exchange: "US", Name: m["2. name"], Type: m["3. type"], Country: "USA", Currency: m["8. currency"]})
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}
