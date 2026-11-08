package market

import (
	"gorm.io/gorm"
	"keepnode/pkg/apiresponse"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	marketmodel "keepnode/app/models/market"
	marketservice "keepnode/app/services/market"
)

type MarketController struct {
	marketService *marketservice.MarketService
	db            *gorm.DB
}

func NewMarketController(marketService *marketservice.MarketService, databases ...*gorm.DB) *MarketController {
	c := &MarketController{marketService: marketService}
	if len(databases) > 0 {
		c.db = databases[0]
	}
	return c
}

// GetEODData retrieves End-of-Day historical data
func (ctrl *MarketController) GetEODData(c *fiber.Ctx, userID uuid.UUID) error {
	symbolParam := c.Params("symbol")
	if strings.HasPrefix(symbolParam, "MANUAL-") {
		return ctrl.manual(c, userID, symbolParam)
	}
	if symbolParam == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Symbol is required")
	}

	symbol, exchange := parseSymbol(symbolParam)

	req := &marketmodel.EODRequest{
		Symbol:   symbol,
		Exchange: exchange,
		From:     c.Query("from"),
		To:       c.Query("to"),
		Period:   c.Query("period"),
		Order:    c.Query("order"),
	}

	data, err := ctrl.marketService.GetHistoricalData(req)
	if err != nil {
		return marketErrorResponse(c, err)
	}

	return apiresponse.RespondOK(c, data)
}

// GetLivePrice retrieves real-time (delayed) price
func (ctrl *MarketController) GetLivePrice(c *fiber.Ctx, userID uuid.UUID) error {
	symbolParam := c.Params("symbol")
	if strings.HasPrefix(symbolParam, "MANUAL-") {
		return ctrl.manual(c, userID, symbolParam)
	}
	if symbolParam == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Symbol is required")
	}

	symbol, exchange := parseSymbol(symbolParam)

	req := &marketmodel.LivePriceRequest{
		Symbol:   symbol,
		Exchange: exchange,
	}

	data, err := ctrl.marketService.GetLivePrice(req)
	if err != nil {
		return marketErrorResponse(c, err)
	}

	return apiresponse.RespondOK(c, data)
}

// GetMultipleLivePrices retrieves real-time prices for multiple symbols
func (ctrl *MarketController) GetMultipleLivePrices(c *fiber.Ctx, _ uuid.UUID) error {
	symbolsParam := c.Query("symbols")
	if symbolsParam == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Symbols parameter is required")
	}

	symbols := strings.Split(symbolsParam, ",")
	if len(symbols) > 50 {
		return fiber.NewError(400, "Request at most 50 quotes")
	}
	for _, symbol := range symbols {
		if strings.HasPrefix(strings.TrimSpace(symbol), "MANUAL-") {
			return fiber.NewError(400, "Request manual valuations individually")
		}
	}
	for i := range symbols {
		symbols[i] = strings.TrimSpace(symbols[i])
	}

	data, err := ctrl.marketService.GetMultipleLivePrices(symbols)
	if err != nil {
		return marketErrorResponse(c, err)
	}

	return apiresponse.RespondOK(c, data)
}

// SearchAssets searches for assets by query
func (ctrl *MarketController) SearchAssets(c *fiber.Ctx, _ uuid.UUID) error {
	query := c.Query("q")
	if query == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Search query 'q' is required")
	}

	limit := c.QueryInt("limit", 10)

	req := &marketmodel.SearchRequest{
		Query: query,
		Limit: limit,
	}

	results, err := ctrl.marketService.SearchAssets(req)
	if err != nil {
		return marketErrorResponse(c, err)
	}

	return apiresponse.RespondOK(c, results)
}

func parseSymbol(symbolParam string) (symbol, exchange string) {
	parts := strings.SplitN(symbolParam, ".", 2)
	symbol = parts[0]
	if len(parts) > 1 {
		exchange = parts[1]
	}
	return
}
