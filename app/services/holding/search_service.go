package holding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	holdingmodel "keepnode/app/models/holding"
	holdingquery "keepnode/app/queries/holding"
	"keepnode/pkg/typesense"
)

const HoldingsCollection = "keepfin_holdings_v1"

type SearchService struct {
	db         *gorm.DB
	client     *typesense.Client
	collection string
}

func NewSearchService(db *gorm.DB, client *typesense.Client, collection string) *SearchService {
	return &SearchService{db: db, client: client, collection: collection}
}

func (s *SearchService) ensureCollection(ctx context.Context) error {
	if s.client == nil {
		return errors.New("search is not configured")
	}
	_, err := s.client.Request(ctx, http.MethodGet, "/collections/"+s.collection, nil, nil)
	if err == nil {
		return nil
	}
	if !typesense.IsStatus(err, http.StatusNotFound) {
		return err
	}
	fields := []map[string]any{}
	for _, name := range []string{"user_id", "account_id", "exchange", "currency", "asset_class", "sector", "country"} {
		fields = append(fields, map[string]any{"name": name, "type": "string", "facet": true})
	}
	for _, name := range []string{"symbol", "name"} {
		fields = append(fields, map[string]any{"name": name, "type": "string", "sort": true, "infix": true})
	}
	for _, name := range []string{"quantity", "buy_price"} {
		fields = append(fields, map[string]any{"name": name, "type": "float"})
	}
	for _, name := range []string{"revision", "created_at"} {
		fields = append(fields, map[string]any{"name": name, "type": "int64"})
	}
	body, err := json.Marshal(map[string]any{"name": s.collection, "fields": fields})
	if err != nil {
		return err
	}
	_, err = s.client.Request(ctx, http.MethodPost, "/collections", nil, body)
	if err != nil && !typesense.IsStatus(err, http.StatusConflict) {
		return err
	}
	// Recreating a lost collection must also requeue previously acknowledged rows.
	return s.db.WithContext(ctx).Model(&holdingmodel.SearchState{}).
		Where("indexed_revision > 0").Update("indexed_revision", 0).Error
}

func documentsFromHoldings(rows []holdingmodel.Holding, state holdingmodel.SearchState) []holdingmodel.SearchDocument {
	documents := make([]holdingmodel.SearchDocument, 0, len(rows))
	for _, h := range rows {
		exchange := ""
		if at := strings.LastIndex(h.Symbol, "."); at >= 0 {
			exchange = h.Symbol[at+1:]
		}
		documents = append(documents, holdingmodel.SearchDocument{
			ID: h.ID.String(), AccountID: h.AccountID.String(), UserID: state.UserID.String(), Revision: state.Revision,
			Symbol: h.Symbol, Name: h.Name, Exchange: exchange, Currency: h.BuyCurrency,
			AssetClass: h.AssetClass, Sector: h.Sector, Country: h.Country,
			Quantity: h.Quantity, BuyPrice: h.BuyPrice, CreatedAt: h.CreatedAt.Unix(),
		})
	}
	return documents
}

func (s *SearchService) syncLocked(ctx context.Context, tx *gorm.DB, state *holdingmodel.SearchState) error {
	if state.IndexedRevision == state.Revision && state.IndexedAt != nil && time.Since(*state.IndexedAt) < 5*time.Minute {
		return nil
	}
	rows, err := holdingquery.SearchSnapshot(tx, state.AccountID, state.UserID)
	if err != nil {
		return err
	}
	documents := documentsFromHoldings(rows, *state)
	// Check every batch before acknowledging the outbox. Repeating an interrupted
	// import is safe because document IDs are the database holding IDs.
	for start := 0; start < len(documents); start += 100 {
		batch := []json.RawMessage{}
		for _, doc := range documents[start:min(start+100, len(documents))] {
			encoded, e := json.Marshal(doc)
			if e != nil {
				return e
			}
			batch = append(batch, encoded)
		}
		if err = s.client.Import(ctx, s.collection, batch); err != nil {
			return err
		}
	}
	filter := "account_id:=" + state.AccountID.String()
	if len(documents) > 0 {
		filter += " && revision:!=" + strconv.FormatInt(state.Revision, 10)
	}
	_, err = s.client.Request(ctx, http.MethodDelete, "/collections/"+s.collection+"/documents",
		url.Values{"filter_by": {filter}}, nil)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	return tx.Model(state).Updates(map[string]any{"indexed_revision": state.Revision, "indexed_at": now}).Error
}

// SyncAccount serializes foreground and background writers across API replicas.
// The row lock also prevents acknowledging a newer, not-yet-indexed revision.
func (s *SearchService) SyncAccount(ctx context.Context, accountID uuid.UUID) error {
	if err := s.ensureCollection(ctx); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var state holdingmodel.SearchState
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&state, "account_id = ?", accountID).Error; err != nil {
			return err
		}
		return s.syncLocked(ctx, tx, &state)
	})
}

func (s *SearchService) FlushUser(ctx context.Context, userID uuid.UUID) error {
	states := []holdingmodel.SearchState{}
	if err := s.db.WithContext(ctx).Where("user_id = ? AND revision > indexed_revision", userID).
		Order("account_id").Find(&states).Error; err != nil {
		return err
	}
	for _, state := range states {
		if err := s.SyncAccount(ctx, state.AccountID); err != nil {
			return err
		}
	}
	return nil
}

// Run replays durable work after crashes and periodically reconciles lost index
// documents. Foreground writes and reads do not wait for the worker's interval.
func (s *SearchService) Run(ctx context.Context) {
	if s.client == nil {
		return
	}
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		states := []holdingmodel.SearchState{}
		err := s.db.WithContext(ctx).Where("revision > indexed_revision OR indexed_at < ?", time.Now().Add(-5*time.Minute)).
			Order("attempted_at NULLS FIRST, account_id").Limit(25).Find(&states).Error
		if err == nil {
			for _, state := range states {
				attempt, cancel := context.WithTimeout(ctx, 3*time.Second)
				err = s.SyncAccount(attempt, state.AccountID)
				cancel()
				// Rotate failing accounts too, so one bad document cannot starve others.
				if e := s.db.WithContext(ctx).Model(&state).UpdateColumn("attempted_at", time.Now().UTC()).Error; e != nil {
					err = e
				}
				if ctx.Err() != nil {
					return
				}
			}
		}
		if err != nil && ctx.Err() == nil {
			log.Printf("Holdings index retry pending: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

var facetFields = []string{"currency", "exchange", "asset_class", "sector", "country"}

func searchFilters(userID, accountID uuid.UUID, revision int64, p holdingmodel.SearchParams) string {
	parts := []string{"user_id:=" + userID.String(), "account_id:=" + accountID.String(),
		"revision:=" + strconv.FormatInt(revision, 10)}
	for key, value := range map[string]string{
		"currency": p.Currency, "exchange": p.Exchange, "asset_class": p.AssetClass, "sector": p.Sector, "country": p.Country,
	} {
		if value != "" {
			parts = append(parts, key+":=`"+value+"`")
		}
	}
	return strings.Join(parts, " && ")
}

func (s *SearchService) searchIndex(ctx context.Context, userID, accountID uuid.UUID, p holdingmodel.SearchParams) (*holdingmodel.SearchPage, error) {
	if err := s.ensureCollection(ctx); err != nil {
		return nil, err
	}
	var page *holdingmodel.SearchPage
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var state holdingmodel.SearchState
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&state, "account_id = ? AND user_id = ?", accountID, userID).Error; err != nil {
			return err
		}
		if err := s.syncLocked(ctx, tx, &state); err != nil {
			return err
		}
		q := p.Query
		if q == "" {
			q = "*"
		}
		order := p.Sort + ":" + p.Direction + ",symbol:asc"
		if p.Sort == "symbol" {
			order = "symbol:" + p.Direction
		}
		if p.Sort == "relevance" {
			order = "_text_match:desc,symbol:asc"
		}
		params := url.Values{
			"q": {q}, "query_by": {"symbol,name,asset_class,sector,country,currency,exchange"},
			"query_by_weights": {"5,4,2,2,1,1,1"}, "infix": {"fallback,fallback,off,off,off,off,off"},
			"filter_by": {searchFilters(userID, accountID, state.Revision, p)}, "sort_by": {order},
			"page": {strconv.Itoa(p.Page)}, "per_page": {strconv.Itoa(p.PerPage)},
			"facet_by": {strings.Join(facetFields, ",")}, "max_facet_values": {"100"},
			"drop_tokens_threshold": {"0"}, "use_cache": {"false"},
		}
		data, err := s.client.Request(ctx, http.MethodGet, "/collections/"+s.collection+"/documents/search", params, nil)
		if err != nil {
			return err
		}
		var result struct {
			Found int `json:"found"`
			Hits  []struct {
				Document holdingmodel.SearchDocument `json:"document"`
			} `json:"hits"`
			Facets []struct {
				Field  string `json:"field_name"`
				Counts []struct {
					Value string `json:"value"`
				} `json:"counts"`
			} `json:"facet_counts"`
		}
		if err = json.Unmarshal(data, &result); err != nil {
			return errors.New("invalid search response")
		}
		page = newSearchPage(p, result.Found, "typesense")
		for _, hit := range result.Hits {
			if hit.Document.UserID != userID.String() || hit.Document.AccountID != accountID.String() ||
				hit.Document.Revision != state.Revision {
				return errors.New("search scope mismatch")
			}
			page.Documents = append(page.Documents, hit.Document)
		}
		for _, facet := range result.Facets {
			for _, count := range facet.Counts {
				if count.Value != "" {
					page.Facets[facet.Field] = append(page.Facets[facet.Field], count.Value)
				}
			}
			sort.Strings(page.Facets[facet.Field])
		}
		return nil
	})
	return page, err
}

func newSearchPage(p holdingmodel.SearchParams, total int, source string) *holdingmodel.SearchPage {
	return &holdingmodel.SearchPage{
		Documents: []holdingmodel.SearchDocument{}, Holdings: []holdingmodel.HoldingWithLiveData{},
		Total: total, Page: p.Page, PerPage: p.PerPage, Pages: max(1, (total+p.PerPage-1)/p.PerPage),
		Source: source, Facets: map[string][]string{},
	}
}

func (s *SearchService) Search(ctx context.Context, userID, accountID uuid.UUID, p holdingmodel.SearchParams) (*holdingmodel.SearchPage, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	attempt, cancel := context.WithTimeout(ctx, 2500*time.Millisecond)
	page, err := s.searchIndex(attempt, userID, accountID, p)
	cancel()
	if err == nil {
		return page, nil
	}
	// Never return a stale index after a successful financial write. The fallback
	// preserves exact filters/pagination while typo tolerance awaits recovery.
	rows, err := holdingquery.SearchSnapshot(s.db.WithContext(ctx), accountID, userID)
	if err != nil {
		return nil, fmt.Errorf("read holdings: %w", err)
	}
	return searchDatabase(documentsFromHoldings(rows, holdingmodel.SearchState{UserID: userID}), p), nil
}
