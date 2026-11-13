package market

import (
	"math"
	"testing"
	"time"

	marketmodel "keepnode/app/models/market"
)

func TestPriceBookConvertsQuoteCurrency(t *testing.T) {
	day := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	date := day.Format("2006-01-02")
	b := NewPriceBook(nil, nil, day, day)
	for symbol, close := range map[string]float64{"VOD.LSE": 70, "QQQ.US": 10, "MYSTERY.US": 3, "GBPUSD.FOREX": 1.25, "USDTRY.FOREX": 40} {
		b.series[symbol] = []marketmodel.MarketPrice{{Symbol: symbol, Date: date, Close: close}}
	}
	// Only the inverse pair exists, so Rate must fall back to it without a database.
	b.series["USDGBP.FOREX"] = []marketmodel.MarketPrice{}
	b.quotes["VOD.LSE"], b.quotes["QQQ.US"], b.quotes["MYSTERY.US"] = "GBX", "USD", ""
	for _, c := range []struct {
		symbol, currency string
		want             float64
	}{
		{"VOD.LSE", "GBP", 0.7},
		{"VOD.LSE", "USD", 0.875},
		{"QQQ.US", "USD", 10},
		{"QQQ.US", "TRY", 400},
		// Unknown quote currencies keep the previous same-currency assumption.
		{"MYSTERY.US", "TRY", 3},
	} {
		got, _, err := b.PriceIn(c.symbol, c.currency, day)
		if err != nil || math.Abs(got-c.want) > 1e-9 {
			t.Fatalf("%s in %s = %v, %v; want %v", c.symbol, c.currency, got, err, c.want)
		}
	}
	if rate, err := b.Rate("USD", "GBX", day); err != nil || math.Abs(rate-80) > 1e-9 {
		t.Fatalf("USD/GBX = %v, %v", rate, err)
	}
}

func TestMajorCurrency(t *testing.T) {
	for in, want := range map[string]string{"GBX": "GBP", "ZAC": "ZAR", "ILA": "ILS", "USD": "USD", "": ""} {
		if got := MajorCurrency(in); got != want {
			t.Fatalf("MajorCurrency(%q) = %q", in, got)
		}
	}
	if got := normalizeQuoteCurrency(" GBp "); got != "GBX" {
		t.Fatalf("normalize GBp = %q", got)
	}
}
