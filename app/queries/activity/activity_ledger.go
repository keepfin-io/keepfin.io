package activity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	activitymodel "keepnode/app/models/activity"
)

// GetLedgerEvents returns ledger entries in their deterministic replay order.
func GetLedgerEvents(db *gorm.DB, id uuid.UUID) ([]activitymodel.Activity, error) {
	a := []activitymodel.Activity{}
	err := db.Where("account_id = ? AND ledger = true", id).Order("date, created_at, id").Find(&a).Error
	return a, err
}
