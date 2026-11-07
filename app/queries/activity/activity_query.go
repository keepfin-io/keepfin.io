package activity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	activitymodel "keepnode/app/models/activity"
)

type ActivityQueries struct {
	DB *gorm.DB
}

func (q *ActivityQueries) CreateActivity(activity *activitymodel.Activity) error {
	return q.DB.Create(activity).Error
}

func (q *ActivityQueries) GetActivitiesByAccountID(accountID uuid.UUID) ([]activitymodel.Activity, error) {
	var activities []activitymodel.Activity
	err := q.DB.Where("account_id = ?", accountID).
		Order("date DESC, created_at DESC").
		Find(&activities).Error
	return activities, err
}

func (q *ActivityQueries) GetActivitiesByHoldingID(holdingID uuid.UUID) ([]activitymodel.Activity, error) {
	var activities []activitymodel.Activity
	err := q.DB.Where("holding_id = ?", holdingID).
		Order("date DESC, created_at DESC").
		Find(&activities).Error
	return activities, err
}

func (q *ActivityQueries) GetActivityByID(id uuid.UUID) (activitymodel.Activity, error) {
	var activity activitymodel.Activity
	err := q.DB.Where("id = ?", id).First(&activity).Error
	return activity, err
}

func (q *ActivityQueries) GetActivityByIDAndAccount(id, accountID uuid.UUID) (activitymodel.Activity, error) {
	var activity activitymodel.Activity
	err := q.DB.Where("id = ? AND account_id = ?", id, accountID).First(&activity).Error
	return activity, err
}

func (q *ActivityQueries) DeleteActivity(id uuid.UUID) error {
	return q.DB.Delete(&activitymodel.Activity{}, id).Error
}
func (q *ActivityQueries) GetActivitiesBySymbol(accountID uuid.UUID, symbol string) ([]activitymodel.Activity, error) {
	var activities []activitymodel.Activity
	err := q.DB.Where("account_id = ? AND symbol = ?", accountID, symbol).
		Order("date DESC").
		Find(&activities).Error
	return activities, err
}

func (q *ActivityQueries) CountActivitiesByAccountID(accountID uuid.UUID) (int64, error) {
	var count int64
	err := q.DB.Model(&activitymodel.Activity{}).Where("account_id = ?", accountID).Count(&count).Error
	return count, err
}
