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

type ActivityService struct {
	DB            *gorm.DB
	marketService *MarketService
}

func NewActivityService(db *gorm.DB, marketService *MarketService) *ActivityService {
	return &ActivityService{
		DB:            db,
		marketService: marketService,
	}
}

// CreateActivity creates a new activity record
func (s *ActivityService) CreateActivity(accountID uuid.UUID, holdingID *uuid.UUID, req *models.CreateActivityRequest) (*models.Activity, error) {
	if req.Symbol == "" {
		return nil, errors.New("symbol is required")
	}
	if req.Type == "" {
		return nil, errors.New("activity type is required")
	}
	if req.Date == "" {
		return nil, errors.New("date is required")
	}

	// Validate activity type
	validTypes := models.ValidActivityTypes()
	isValid := false
	for _, t := range validTypes {
		if t == req.Type {
			isValid = true
			break
		}
	}
	if !isValid {
		return nil, errors.New("invalid activity type")
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		// Try alternative format
		date, err = time.Parse(time.RFC3339, req.Date)
		if err != nil {
			return nil, errors.New("invalid date format, use YYYY-MM-DD")
		}
	}

	// Normalize symbol
	symbol := strings.ToUpper(req.Symbol)

	// Try to get asset name from market service
	var name string
	if s.marketService != nil {
		searchReq := &models.SearchRequest{Query: strings.Split(symbol, ".")[0], Limit: 1}
		results, err := s.marketService.SearchAssets(searchReq)
		if err == nil && len(results) > 0 {
			name = results[0].Name
		}
	}

	// Set default currency
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	activity := &models.Activity{
		AccountID: accountID,
		HoldingID: holdingID,
		Symbol:    symbol,
		Name:      name,
		Type:      req.Type,
		Date:      date,
		Quantity:  req.Quantity,
		UnitPrice: req.UnitPrice,
		Fee:       req.Fee,
		Currency:  currency,
		Notes:     req.Notes,
	}

	activityQueries := &queries.ActivityQueries{DB: s.DB}
	if err := activityQueries.CreateActivity(activity); err != nil {
		return nil, err
	}

	// Calculate total after creation
	activity.Total = (activity.Quantity * activity.UnitPrice) + activity.Fee

	return activity, nil
}

// CreateActivityFromHolding creates a buy activity when a holding is added
func (s *ActivityService) CreateActivityFromHolding(accountID uuid.UUID, holding *models.Holding, fee float64) (*models.Activity, error) {
	req := &models.CreateActivityRequest{
		Symbol:    holding.Symbol,
		Type:      models.ActivityTypeBuy,
		Date:      time.Now().Format("2006-01-02"),
		Quantity:  holding.Quantity,
		UnitPrice: holding.BuyPrice,
		Fee:       fee,
		Currency:  holding.BuyCurrency,
		Notes:     holding.Notes,
	}

	return s.CreateActivity(accountID, &holding.ID, req)
}

// GetActivities returns all activities for an account
func (s *ActivityService) GetActivities(accountID uuid.UUID) ([]models.Activity, error) {
	activityQueries := &queries.ActivityQueries{DB: s.DB}
	return activityQueries.GetActivitiesByAccountID(accountID)
}

// GetActivitiesByHolding returns all activities for a specific holding
func (s *ActivityService) GetActivitiesByHolding(holdingID uuid.UUID) ([]models.Activity, error) {
	activityQueries := &queries.ActivityQueries{DB: s.DB}
	return activityQueries.GetActivitiesByHoldingID(holdingID)
}

// DeleteActivity deletes an activity
func (s *ActivityService) DeleteActivity(activityID, accountID uuid.UUID) error {
	activityQueries := &queries.ActivityQueries{DB: s.DB}

	// Verify activity belongs to account
	_, err := activityQueries.GetActivityByIDAndAccount(activityID, accountID)
	if err != nil {
		return errors.New("activity not found")
	}

	return activityQueries.DeleteActivity(activityID)
}

// GetActivityCount returns the number of activities for an account
func (s *ActivityService) GetActivityCount(accountID uuid.UUID) (int64, error) {
	activityQueries := &queries.ActivityQueries{DB: s.DB}
	return activityQueries.CountActivitiesByAccountID(accountID)
}
