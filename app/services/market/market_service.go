package market

import (
	"errors"
	"os"
	"sync"

	marketmodel "keepnode/app/models/market"
)

type MarketService struct {
	eodhd    *EODHDClient
	alpha    *AlphaVantageClient
	mu       sync.RWMutex
	provider string
}

func NewMarketService() *MarketService {
	return &MarketService{
		eodhd:    NewEODHDClient(),
		alpha:    NewAlphaVantageClient(),
		provider: os.Getenv("MARKET_DATA_PROVIDER"),
	}
}

func (s *MarketService) Provider() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.provider == "alpha-vantage" {
		return s.provider
	}
	return "eodhd"
}
func (s *MarketService) SetProvider(name string) error {
	if name != "eodhd" && name != "alpha-vantage" {
		return errors.New("invalid market provider")
	}
	s.mu.Lock()
	s.provider = name
	s.mu.Unlock()
	return nil
}
func (s *MarketService) Configured(name string) bool {
	if name == "eodhd" {
		return s.eodhd != nil && s.eodhd.APIToken != ""
	}
	return name == "alpha-vantage" && s.alpha != nil && s.alpha.APIKey != ""
}

func (s *MarketService) GetHistoricalData(req *marketmodel.EODRequest) ([]marketmodel.EODData, error) {
	if req.Symbol == "" {
		return nil, errors.New("symbol is required")
	}

	exchange := req.Exchange
	if exchange == "" {
		exchange = "US"
	}

	period := req.Period
	if period == "" {
		period = "d"
	}

	if period != "d" && period != "w" && period != "m" {
		return nil, errors.New("period must be 'd' (daily), 'w' (weekly), or 'm' (monthly)")
	}

	order := req.Order
	if order == "" {
		order = "a"
	}

	if order != "a" && order != "d" {
		return nil, errors.New("order must be 'a' (ascending) or 'd' (descending)")
	}

	if s.Provider() == "alpha-vantage" {
		return s.alpha.History(req)
	}
	return s.eodhd.GetEODData(req.Symbol, exchange, req.From, req.To, period, order)
}

func (s *MarketService) GetLivePrice(req *marketmodel.LivePriceRequest) (*marketmodel.LivePrice, error) {
	if req.Symbol == "" {
		return nil, errors.New("symbol is required")
	}

	exchange := req.Exchange
	if exchange == "" {
		exchange = "US"
	}

	if s.Provider() == "alpha-vantage" {
		return s.alpha.Quote(req)
	}
	return s.eodhd.GetLivePrice(req.Symbol, exchange)
}

func (s *MarketService) GetMultipleLivePrices(symbols []string) ([]marketmodel.LivePrice, error) {
	if len(symbols) == 0 {
		return nil, errors.New("at least one symbol is required")
	}

	if s.Provider() == "alpha-vantage" {
		if len(symbols) > 20 {
			return nil, errors.New("request at most 20 quotes")
		}
		out := []marketmodel.LivePrice{}
		for _, symbol := range symbols {
			v, e := s.alpha.Quote(&marketmodel.LivePriceRequest{Symbol: symbol})
			if e != nil {
				return nil, e
			}
			out = append(out, *v)
		}
		return out, nil
	}
	return s.eodhd.GetMultipleLivePrices(symbols)
}

func (s *MarketService) SearchAssets(req *marketmodel.SearchRequest) ([]marketmodel.SearchResult, error) {
	if req.Query == "" {
		return nil, errors.New("search query is required")
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	if s.Provider() == "alpha-vantage" {
		return s.alpha.Search(req)
	}
	return s.eodhd.SearchSymbol(req.Query, limit)
}
