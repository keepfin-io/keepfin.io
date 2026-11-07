package activity

import (
	"encoding/csv"
	"strings"
	"testing"
)

func TestJSONConversionRetainsNotesRatioAndRejectsInvalidDocuments(t *testing.T) {
	data, err := JSONToCSV(`{"version":1,"activities":[{"reference":"split1","date":"2026-01-02","type":"split","symbol":"ABC.US","quantity":0,"unit_price":0,"currency":"USD","notes":"=a,\nquote","split_numerator":3,"split_denominator":2}]}`)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(strings.NewReader(data)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if rows[1][10] != "3" || rows[1][11] != "2" || csvRestore(rows[1][9]) != "=a,\nquote" {
		t.Fatalf("lost fields %+v", rows)
	}
	for _, invalid := range []string{`{"version":2,"activities":[]}`, `{"version":1,"activities":[]} {}`, `{"version":1,"unknown":true}`} {
		if _, err = JSONToCSV(invalid); err == nil {
			t.Fatal("invalid document accepted")
		}
	}
}
