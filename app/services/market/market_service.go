package market

import (
	"errors"

	marketmodel "keepnode/app/models/market"
)

type MarketService struct {
	eodhd *EODHDClient
}

func NewMarketService() *MarketService {
	return &MarketService{
		eodhd: NewEODHDClient(),
	}
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

	return s.eodhd.GetLivePrice(req.Symbol, exchange)
}

func (s *MarketService) GetMultipleLivePrices(symbols []string) ([]marketmodel.LivePrice, error) {
	if len(symbols) == 0 {
		return nil, errors.New("at least one symbol is required")
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

	return s.eodhd.SearchSymbol(req.Query, limit)
}
