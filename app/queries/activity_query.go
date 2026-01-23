package queries

import (
	"keepnode/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ActivityQueries struct {
	DB *gorm.DB
}

func (q *ActivityQueries) CreateActivity(activity *models.Activity) error {
	return q.DB.Create(activity).Error
}

func (q *ActivityQueries) GetActivitiesByAccountID(accountID uuid.UUID) ([]models.Activity, error) {
	var activities []models.Activity
	err := q.DB.Where("account_id = ?", accountID).
		Order("date DESC, created_at DESC").
		Find(&activities).Error
	return activities, err
}

func (q *ActivityQueries) GetActivitiesByHoldingID(holdingID uuid.UUID) ([]models.Activity, error) {
	var activities []models.Activity
	err := q.DB.Where("holding_id = ?", holdingID).
		Order("date DESC, created_at DESC").
		Find(&activities).Error
	return activities, err
}

func (q *ActivityQueries) GetActivityByID(id uuid.UUID) (models.Activity, error) {
	var activity models.Activity
	err := q.DB.Where("id = ?", id).First(&activity).Error
	return activity, err
}

func (q *ActivityQueries) GetActivityByIDAndAccount(id, accountID uuid.UUID) (models.Activity, error) {
	var activity models.Activity
	err := q.DB.Where("id = ? AND account_id = ?", id, accountID).First(&activity).Error
	return activity, err
}

func (q *ActivityQueries) DeleteActivity(id uuid.UUID) error {
	return q.DB.Delete(&models.Activity{}, id).Error
}
func (q *ActivityQueries) GetActivitiesBySymbol(accountID uuid.UUID, symbol string) ([]models.Activity, error) {
	var activities []models.Activity
	err := q.DB.Where("account_id = ? AND symbol = ?", accountID, symbol).
		Order("date DESC").
		Find(&activities).Error
	return activities, err
}

func (q *ActivityQueries) CountActivitiesByAccountID(accountID uuid.UUID) (int64, error) {
	var count int64
	err := q.DB.Model(&models.Activity{}).Where("account_id = ?", accountID).Count(&count).Error
	return count, err
}
