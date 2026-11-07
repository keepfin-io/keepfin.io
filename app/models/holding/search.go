package holding

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SearchState is a durable, coalescing outbox. Database triggers advance Revision
// in the same transaction as a holding change, including ledger rebuilds.
type SearchState struct {
	AccountID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID `gorm:"type:uuid;not null;index"`
	Revision        int64     `gorm:"not null;default:1"`
	IndexedRevision int64     `gorm:"not null;default:0"`
	IndexedAt       *time.Time
	AttemptedAt     *time.Time
	UpdatedAt       time.Time
}

func (SearchState) TableName() string { return "holding_search_states" }

type SearchDocument struct {
	ID         string  `json:"id"`
	AccountID  string  `json:"account_id"`
	UserID     string  `json:"user_id"`
	Revision   int64   `json:"revision"`
	Symbol     string  `json:"symbol"`
	Name       string  `json:"name"`
	Exchange   string  `json:"exchange"`
	Currency   string  `json:"currency"`
	AssetClass string  `json:"asset_class"`
	Sector     string  `json:"sector"`
	Country    string  `json:"country"`
	Quantity   float64 `json:"quantity"`
	BuyPrice   float64 `json:"buy_price"`
	CreatedAt  int64   `json:"created_at"`
}

type SearchParams struct {
	Query, Currency, Exchange, AssetClass, Sector, Country string
	Sort, Direction                                        string
	Page, PerPage                                          int
}

func (p *SearchParams) Validate() error {
	p.Query = strings.TrimSpace(p.Query)
	if len(p.Query) > 200 {
		return errors.New("search query is too long")
	}
	for _, value := range []string{p.Currency, p.Exchange, p.AssetClass, p.Sector, p.Country} {
		if len(value) > 100 || strings.ContainsAny(value, "`\\\n\r\x00") {
			return errors.New("invalid search filter")
		}
	}
	if p.Page < 1 || p.Page > 100000 {
		return errors.New("page must be between 1 and 100000")
	}
	if p.PerPage != 10 && p.PerPage != 25 && p.PerPage != 50 && p.PerPage != 100 {
		return errors.New("per_page must be 10, 25, 50 or 100")
	}
	switch p.Sort {
	case "relevance", "symbol", "name", "quantity", "buy_price", "created_at":
	default:
		return errors.New("invalid sort field")
	}
	if p.Direction != "asc" && p.Direction != "desc" {
		return errors.New("invalid sort direction")
	}
	return nil
}

type SearchPage struct {
	Documents []SearchDocument      `json:"-"`
	Holdings  []HoldingWithLiveData `json:"holdings"`
	Total     int                   `json:"total"`
	Page      int                   `json:"page"`
	PerPage   int                   `json:"per_page"`
	Pages     int                   `json:"pages"`
	Source    string                `json:"source"`
	Facets    map[string][]string   `json:"facets"`
}
