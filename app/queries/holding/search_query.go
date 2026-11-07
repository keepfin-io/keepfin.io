package holding

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	holdingmodel "keepnode/app/models/holding"
)

// SearchSnapshot excludes soft-deleted accounts and positions, even if a stale
// index still contains them. Ownership is always checked against PostgreSQL.
func SearchSnapshot(db *gorm.DB, accountID, userID uuid.UUID) ([]holdingmodel.Holding, error) {
	holdings := []holdingmodel.Holding{}
	err := db.Model(&holdingmodel.Holding{}).Select("holdings.*").
		Joins("JOIN accounts ON accounts.id = holdings.account_id").
		Where("holdings.account_id = ? AND accounts.user_id = ? AND accounts.deleted_at IS NULL", accountID, userID).
		Where("holdings.quantity > 0").Order("holdings.symbol, holdings.id").Find(&holdings).Error
	return holdings, err
}
