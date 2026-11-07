package activity

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"io"
	"strconv"
	"strings"
)

// JSON uses the same preview, validation, ownership and idempotency pipeline as CSV.
type JSONActivity struct {
	Reference        string  `json:"reference"`
	Date             string  `json:"date"`
	Type             string  `json:"type"`
	Symbol           string  `json:"symbol"`
	Quantity         float64 `json:"quantity"`
	UnitPrice        float64 `json:"unit_price"`
	Fee              float64 `json:"fee"`
	Currency         string  `json:"currency"`
	ExchangeRate     float64 `json:"exchange_rate"`
	Notes            string  `json:"notes"`
	SplitNumerator   int64   `json:"split_numerator,omitempty"`
	SplitDenominator int64   `json:"split_denominator,omitempty"`
}
type JSONDocument struct {
	Version    int            `json:"version"`
	Activities []JSONActivity `json:"activities"`
}

func JSONToCSV(data string) (string, error) {
	if len(data) > 2*1024*1024 {
		return "", errors.New("JSON must be smaller than 2 MB")
	}
	var doc JSONDocument
	decoder := json.NewDecoder(strings.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&doc); err != nil {
		return "", errors.New("invalid activity JSON")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return "", errors.New("provide one JSON document")
	}
	if doc.Version != 1 || len(doc.Activities) == 0 || len(doc.Activities) > 1000 {
		return "", errors.New("version 1 and between 1 and 1000 activities are required")
	}
	var out bytes.Buffer
	w := csv.NewWriter(&out)
	if err := w.Write(csvHeaders); err != nil {
		return "", err
	}
	number := func(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }
	for _, a := range doc.Activities {
		if err := w.Write([]string{a.Reference, a.Date, a.Type, a.Symbol, number(a.Quantity), number(a.UnitPrice), number(a.Fee), a.Currency, number(a.ExchangeRate), csvSafe(a.Notes), strconv.FormatInt(a.SplitNumerator, 10), strconv.FormatInt(a.SplitDenominator, 10)}); err != nil {
			return "", err
		}
	}
	w.Flush()
	return out.String(), w.Error()
}
func (s *ActivityService) ImportJSON(account uuid.UUID, data string, commit bool) (*ImportResult, error) {
	csvData, err := JSONToCSV(data)
	if err != nil {
		return nil, err
	}
	return s.ImportCSV(account, csvData, commit)
}
func (s *ActivityService) ExportJSON(account uuid.UUID) ([]byte, error) {
	events, err := s.GetActivities(account)
	if err != nil {
		return nil, err
	}
	doc := JSONDocument{Version: 1, Activities: []JSONActivity{}}
	for _, a := range events {
		kind := a.Type
		if !a.Ledger {
			kind = "archived_" + kind
		}
		doc.Activities = append(doc.Activities, JSONActivity{a.ID.String(), a.Date.UTC().Format("2006-01-02"), kind, a.Symbol, a.Quantity, a.UnitPrice, a.Fee, a.Currency, a.ExchangeRate, a.Notes, a.SplitNumerator, a.SplitDenominator})
	}
	return json.MarshalIndent(doc, "", "  ")
}
