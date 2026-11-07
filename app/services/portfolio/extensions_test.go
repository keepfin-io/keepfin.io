package portfolio

import (
	"math"
	"testing"
)

func TestXIRRKnownAndAmbiguousFlows(t *testing.T) {
	rate := xirr([]float64{-1000, 1100}, []float64{0, 365})
	if rate == nil || math.Abs(*rate-.1) > 1e-9 {
		t.Fatalf("known annual return %v", rate)
	}
	if rate = xirr([]float64{-100, 230, -132}, []float64{0, 365, 730}); rate != nil {
		t.Fatal("ambiguous roots must stay unknown")
	}
	if rate = xirr([]float64{0, 100}, []float64{0, 365}); rate != nil {
		t.Fatal("no invested capital")
	}
}
func TestReturnsAndFXAttribution(t *testing.T) {
	value := func(v float64) *float64 { return &v }
	for _, test := range []struct {
		method string
		want   float64
	}{{"twr", 10}, {"roai", 10. / 150 * 100}} {
		r := &PortfolioReport{Complete: true, History: []HistoryPoint{{Date: "2026-01-01", Value: value(100), ValueWithoutFX: value(100)}, {Date: "2026-01-02", Value: value(210), NetFlow: 100, ValueWithoutFX: value(205), FlowWithoutFX: 100}}}
		applyPerformance(r, test.method)
		if r.Performance.PeriodReturn == nil || math.Abs(*r.Performance.PeriodReturn-test.want) > 1e-8 {
			t.Fatalf("%s %+v", test.method, r.Performance)
		}
		if r.Performance.Profit == nil || *r.Performance.Profit != 10 || r.Performance.PriceEffect == nil || *r.Performance.PriceEffect != 5 || r.Performance.CurrencyEffect == nil || *r.Performance.CurrencyEffect != 5 {
			t.Fatalf("attribution %+v", r.Performance)
		}
	}
	r := &PortfolioReport{History: []HistoryPoint{{Date: "2026-01-01", Value: value(100)}, {Date: "2026-01-02"}, {Date: "2026-01-03", Value: value(150)}}}
	applyPerformance(r, "twr")
	if r.Performance.PeriodReturn != nil {
		t.Fatal("missing day treated as known performance")
	}
}
