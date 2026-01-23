package services

import (
	"errors"
	"strings"
	"time"

	"keepnode/app/models"
	"keepnode/app/queries"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HoldingService struct {
	DB              *gorm.DB
	marketService   *MarketService
	activityQueries *queries.ActivityQueries
}

func NewHoldingService(db *gorm.DB, marketService *MarketService) *HoldingService {
	return &HoldingService{
		DB:              db,
		marketService:   marketService,
		activityQueries: &queries.ActivityQueries{DB: db},
	}
}

func (s *HoldingService) AddHolding(accountID uuid.UUID, req *models.CreateHoldingRequest) (*models.Holding, error) {
	if req.Symbol == "" {
		return nil, errors.New("symbol is required")
	}
	if req.Quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	symbol := strings.ToUpper(req.Symbol)

	holdingQueries := &queries.HoldingQueries{DB: s.DB}
	existingHolding, err := holdingQueries.GetHoldingBySymbol(accountID, symbol)
	if err == nil && existingHolding.ID != uuid.Nil {
		return nil, errors.New("holding for this symbol already exists in account")
	}

	var name string
	searchReq := &models.SearchRequest{Query: strings.Split(symbol, ".")[0], Limit: 1}
	results, err := s.marketService.SearchAssets(searchReq)
	if err == nil && len(results) > 0 {
		name = results[0].Name
	}

	buyCurrency := req.BuyCurrency
	if buyCurrency == "" {
		buyCurrency = "USD"
	}

	holding := &models.Holding{
		AccountID:   accountID,
		Symbol:      symbol,
		Name:        name,
		Quantity:    req.Quantity,
		BuyPrice:    req.BuyPrice,
		BuyCurrency: buyCurrency,
		Notes:       req.Notes,
	}

	if err := holdingQueries.CreateHolding(holding); err != nil {
		return nil, err
	}

	activity := &models.Activity{
		AccountID: accountID,
		HoldingID: &holding.ID,
		Symbol:    symbol,
		Name:      name,
		Type:      models.ActivityTypeBuy,
		Date:      time.Now(),
		Quantity:  req.Quantity,
		UnitPrice: req.BuyPrice,
		Fee:       req.Fee,
		Currency:  buyCurrency,
		Notes:     req.Notes,
	}
	_ = s.activityQueries.CreateActivity(activity)

	return holding, nil
}

func (s *HoldingService) GetHoldingsWithLiveData(accountID uuid.UUID) ([]models.HoldingWithLiveData, *models.HoldingsSummary, error) {
	holdingQueries := &queries.HoldingQueries{DB: s.DB}

	holdings, err := holdingQueries.GetHoldingsByAccountID(accountID)
	if err != nil {
		return nil, nil, err
	}

	if len(holdings) == 0 {
		return []models.HoldingWithLiveData{}, &models.HoldingsSummary{}, nil
	}

	symbols := make([]string, len(holdings))
	for i, h := range holdings {
		symbols[i] = h.Symbol
	}

	livePrices, err := s.marketService.GetMultipleLivePrices(symbols)
	if err != nil {
		livePrices = []models.LivePrice{}
	}

	priceMap := make(map[string]models.LivePrice)
	for _, lp := range livePrices {
		priceMap[lp.Code] = lp
	}
	result := make([]models.HoldingWithLiveData, len(holdings))
	summary := &models.HoldingsSummary{}

	for i, h := range holdings {
		currentPrice := h.BuyPrice // default to buy price if no live data
		change := 0.0
		changePercent := 0.0

		if lp, ok := priceMap[h.Symbol]; ok {
			currentPrice = lp.Close
			change = lp.Change
			changePercent = lp.ChangeP
		}

		value := h.Quantity * currentPrice
		cost := h.Quantity * h.BuyPrice
		profitLoss := value - cost
		performance := 0.0
		if cost > 0 {
			performance = (profitLoss / cost) * 100
		}

		result[i] = models.HoldingWithLiveData{
			ID:            h.ID,
			AccountID:     h.AccountID,
			Symbol:        h.Symbol,
			Name:          h.Name,
			Quantity:      h.Quantity,
			BuyPrice:      h.BuyPrice,
			BuyCurrency:   h.BuyCurrency,
			CurrentPrice:  currentPrice,
			Value:         value,
			Change:        change,
			ChangePercent: changePercent,
			Performance:   performance,
			ProfitLoss:    profitLoss,
			FirstActivity: h.CreatedAt,
		}

		// Update summary
		summary.TotalValue += value
		summary.TotalCost += cost
		summary.TotalProfitLoss += profitLoss
		summary.DailyChange += change * h.Quantity
	}

	if summary.TotalCost > 0 {
		summary.TotalPerformance = (summary.TotalProfitLoss / summary.TotalCost) * 100
	}

	return result, summary, nil
}

func (s *HoldingService) GetHoldingByID(holdingID, accountID uuid.UUID) (*models.Holding, error) {
	holdingQueries := &queries.HoldingQueries{DB: s.DB}
	holding, err := holdingQueries.GetHoldingByID(holdingID, accountID)
	if err != nil {
		return nil, errors.New("holding not found")
	}
	return &holding, nil
}

func (s *HoldingService) UpdateHolding(holdingID, accountID uuid.UUID, req *models.UpdateHoldingRequest) (*models.Holding, error) {
	holdingQueries := &queries.HoldingQueries{DB: s.DB}

	holding, err := holdingQueries.GetHoldingByID(holdingID, accountID)
	if err != nil {
		return nil, errors.New("holding not found")
	}
	updates := make(map[string]interface{})
	if req.Quantity != nil && *req.Quantity > 0 {
		updates["quantity"] = *req.Quantity
	}
	if req.BuyPrice != nil && *req.BuyPrice >= 0 {
		updates["buy_price"] = *req.BuyPrice
	}
	if req.BuyCurrency != nil {
		updates["buy_currency"] = *req.BuyCurrency
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	if len(updates) == 0 {
		return &holding, nil
	}

	if err := holdingQueries.UpdateHolding(holdingID, updates); err != nil {
		return nil, err
	}

	updatedHolding, _ := holdingQueries.GetHoldingByID(holdingID, accountID)
	return &updatedHolding, nil
}

func (s *HoldingService) DeleteHolding(holdingID, accountID uuid.UUID) error {
	holdingQueries := &queries.HoldingQueries{DB: s.DB}

	_, err := holdingQueries.GetHoldingByID(holdingID, accountID)
	if err != nil {
		return errors.New("holding not found")
	}

	return holdingQueries.DeleteHolding(holdingID)
}
