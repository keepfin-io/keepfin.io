package account

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	accountmodel "keepnode/app/models/account"
)

// GetAccountForUpdate locks the account row for the duration of the transaction.
func GetAccountForUpdate(tx *gorm.DB, id uuid.UUID) (accountmodel.Account, error) {
	var a accountmodel.Account
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&a, "id = ?", id).Error
	return a, err
}
