package queries

import (
	"keepnode/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HoldingQueries struct {
	DB *gorm.DB
}

func (q *HoldingQueries) CreateHolding(holding *models.Holding) error {
	return q.DB.Create(holding).Error
}

func (q *HoldingQueries) GetHoldingByID(id, accountID uuid.UUID) (models.Holding, error) {
	var holding models.Holding
	err := q.DB.Where("id = ? AND account_id = ?", id, accountID).First(&holding).Error
	return holding, err
}

func (q *HoldingQueries) GetHoldingsByAccountID(accountID uuid.UUID) ([]models.Holding, error) {
	var holdings []models.Holding
	err := q.DB.Where("account_id = ?", accountID).Order("created_at DESC").Find(&holdings).Error
	return holdings, err
}

func (q *HoldingQueries) GetHoldingBySymbol(accountID uuid.UUID, symbol string) (models.Holding, error) {
	var holding models.Holding
	err := q.DB.Where("account_id = ? AND symbol = ?", accountID, symbol).First(&holding).Error
	return holding, err
}

func (q *HoldingQueries) UpdateHolding(id uuid.UUID, updates map[string]interface{}) error {
	return q.DB.Model(&models.Holding{}).Where("id = ?", id).Updates(updates).Error
}

func (q *HoldingQueries) DeleteHolding(id uuid.UUID) error {
	return q.DB.Delete(&models.Holding{}, id).Error
}

func (q *HoldingQueries) GetAllHoldingSymbolsByAccountID(accountID uuid.UUID) ([]string, error) {
	var symbols []string
	err := q.DB.Model(&models.Holding{}).
		Where("account_id = ?", accountID).
		Distinct().
		Pluck("symbol", &symbols).Error
	return symbols, err
}
