package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"io"
	"keepnode/app/models"
	"keepnode/pkg/portfolio"
	"sort"
	"strconv"
	"strings"
	"time"
)

var csvHeaders = []string{"reference", "date", "type", "symbol", "quantity", "unit_price", "fee", "currency", "exchange_rate", "notes"}

type ImportRow struct {
	Line     int                          `json:"line"`
	Activity models.CreateActivityRequest `json:"activity"`
	Status   string                       `json:"status"`
	Error    string                       `json:"error,omitempty"`
}
type ImportResult struct {
	Rows       []ImportRow `json:"rows"`
	Valid      bool        `json:"valid"`
	Imported   int         `json:"imported"`
	Duplicates int         `json:"duplicates"`
	Error      string      `json:"error,omitempty"`
}

func activityKey(a models.Activity) string {
	return fmt.Sprintf("%s|%s|%s|%.8f|%.4f|%.4f|%s|%.10f|%s", a.Date.UTC().Format("2006-01-02"), fmt.Sprintf("%t:%s", a.Ledger, a.Type), a.Symbol, a.Quantity, a.UnitPrice, a.Fee, a.Currency, a.ExchangeRate, a.Notes)
}
func digest(s string) string { v := sha256.Sum256([]byte(s)); return hex.EncodeToString(v[:]) }
func csvSafe(s string) string {
	if len(s) > 0 && strings.ContainsAny(s[:1], "=+-@\t\r") {
		return "'" + s
	}
	return s
}
func csvRestore(s string) string {
	if len(s) > 1 && s[0] == '\'' && strings.ContainsAny(s[1:2], "=+-@\t\r") {
		return s[1:]
	}
	return s
}
func (s *ActivityService) ExportCSV(accountID uuid.UUID) ([]byte, error) {
	events, err := s.GetActivities(accountID)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].Date.Equal(events[j].Date) {
			return events[i].CreatedAt.Before(events[j].CreatedAt)
		}
		return events[i].Date.Before(events[j].Date)
	})
	var out bytes.Buffer
	w := csv.NewWriter(&out)
	if err = w.Write(csvHeaders); err != nil {
		return nil, err
	}
	for _, a := range events {
		kind := a.Type
		if !a.Ledger {
			kind = "archived_" + kind
		}
		values := []string{a.ID.String(), a.Date.UTC().Format("2006-01-02"), kind, a.Symbol, strconv.FormatFloat(a.Quantity, 'f', 8, 64), strconv.FormatFloat(a.UnitPrice, 'f', 4, 64), strconv.FormatFloat(a.Fee, 'f', 4, 64), a.Currency, strconv.FormatFloat(a.ExchangeRate, 'f', 10, 64), csvSafe(a.Notes)}
		if err = w.Write(values); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return out.Bytes(), w.Error()
}
func (s *ActivityService) ImportCSV(accountID uuid.UUID, data string, commit bool) (*ImportResult, error) {
	if len(data) > 2*1024*1024 {
		return nil, errors.New("CSV must be smaller than 2 MB")
	}
	reader := csv.NewReader(strings.NewReader(strings.TrimPrefix(data, "\ufeff")))
	header, err := reader.Read()
	if err != nil {
		return nil, errors.New("CSV header is missing")
	}
	columns := map[string]int{}
	for i, h := range header {
		h = strings.TrimSpace(h)
		if _, ok := columns[h]; ok {
			return nil, errors.New("duplicate CSV header")
		}
		columns[h] = i
	}
	for _, h := range []string{"date", "type", "quantity", "unit_price", "currency"} {
		if _, ok := columns[h]; !ok {
			return nil, fmt.Errorf("missing CSV column: %s", h)
		}
	}
	result := &ImportResult{Rows: []ImportRow{}, Valid: true}
	records := [][]string{}
	for {
		row, e := reader.Read()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, fmt.Errorf("CSV row %d: %w", len(records)+2, e)
		}
		records = append(records, row)
		if len(records) > 1000 {
			return nil, errors.New("import up to 1000 rows at a time")
		}
	}
	if len(records) == 0 {
		return nil, errors.New("CSV contains no activities")
	}
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		account, e := lockAccount(tx, accountID)
		if e != nil {
			return e
		}
		if !account.IsActive {
			return errors.New("account is inactive")
		}
		existing, e := ledgerEvents(tx, accountID)
		if e != nil {
			return e
		}
		allEvents := []models.Activity{}
		if e = tx.Unscoped().Where("account_id = ?", accountID).Find(&allEvents).Error; e != nil {
			return e
		}
		keys := map[string]bool{}
		ids := map[string]bool{}
		for _, a := range allEvents {
			keys["content:"+digest(activityKey(a))] = true
			ids[a.ID.String()] = true
			if a.ImportKey != nil {
				keys[*a.ImportKey] = true
			}
		}
		pending := []models.Activity{}
		importedAt := time.Now().UTC()
		for i, row := range records {
			get := func(name string) string {
				j, ok := columns[name]
				if !ok {
					return ""
				}
				return csvRestore(strings.TrimSpace(row[j]))
			}
			req := models.CreateActivityRequest{Date: get("date"), Type: get("type"), Symbol: get("symbol"), Currency: get("currency"), Notes: get("notes")}
			item := ImportRow{Line: i + 2, Status: "ready"}
			reference := get("reference")
			if ids[reference] {
				item.Status = "duplicate"
				item.Activity = req
				result.Duplicates++
				result.Rows = append(result.Rows, item)
				continue
			}
			for _, n := range []struct {
				name     string
				target   *float64
				optional bool
			}{{"quantity", &req.Quantity, false}, {"unit_price", &req.UnitPrice, false}, {"fee", &req.Fee, true}, {"exchange_rate", &req.ExchangeRate, true}} {
				value := get(n.name)
				if value == "" && n.optional {
					continue
				}
				v, e := strconv.ParseFloat(value, 64)
				if e != nil || !portfolio.ValidNumber(v) {
					item.Error = "invalid " + n.name
					break
				}
				*n.target = v
			}
			item.Activity = req
			archived := strings.HasPrefix(req.Type, "archived_")
			prepareAccount := account
			if archived {
				req.Type = strings.TrimPrefix(req.Type, "archived_")
				prepareAccount.OpeningDate = nil
				prepareAccount.Currency = req.Currency
			}
			if req.Type == "transfer_in" {
				req.Type = "deposit"
				req.Notes = strings.TrimSpace(req.Notes + " [Imported transfer credit]")
			}
			if req.Type == "transfer_out" {
				req.Type = "withdrawal"
				req.Notes = strings.TrimSpace(req.Notes + " [Imported transfer debit]")
			}
			a, e := s.prepare(prepareAccount, &req, archived || req.Type == "opening")
			if archived {
				a.Ledger = false
				a.ExchangeRate = req.ExchangeRate
			}
			a.CreatedAt = importedAt.Add(time.Duration(i) * time.Microsecond)
			item.Activity = req
			if e != nil && item.Error == "" {
				item.Error = e.Error()
			}
			if item.Error != "" {
				item.Status = "error"
				result.Valid = false
				result.Rows = append(result.Rows, item)
				continue
			}
			key := "content:" + digest(activityKey(a))
			if reference != "" {
				key = "reference:" + digest(reference)
			}
			if keys[key] {
				item.Status = "duplicate"
				result.Duplicates++
			} else {
				keys[key] = true
				a.ImportKey = &key
				pending = append(pending, a)
			}
			result.Rows = append(result.Rows, item)
		}
		combined := append(append([]models.Activity(nil), existing...), pending...)
		if e := resolveOpeningRates(tx, s.marketService, account, combined); e != nil {
			result.Valid = false
			result.Error = e.Error()
		}
		if result.Valid {
			if _, e := portfolio.Replay(account, combined, time.Now().UTC()); e != nil {
				result.Valid = false
				result.Error = e.Error()
			}
		}
		if commit && result.Valid {
			for i := range pending {
				if e := tx.Create(&pending[i]).Error; e != nil {
					return e
				}
			}
			if e := s.rebuild(tx, account); e != nil {
				return e
			}
			result.Imported = len(pending)
		}
		return nil
	})
	return result, err
}
