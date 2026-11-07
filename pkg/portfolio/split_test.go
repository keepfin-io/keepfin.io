package portfolio

import (
	accountmodel "keepnode/app/models/account"
	activitymodel "keepnode/app/models/activity"
	"testing"
	"time"
)

func TestSplitsPreserveCostAndCash(t *testing.T) {
	day := Day(time.Now()).AddDate(0, 0, -4)
	a := accountmodel.Account{Currency: "USD", InitialBalance: 1000}
	events := []activitymodel.Activity{
		{Type: "buy", Symbol: "TEST.US", Currency: "USD", Date: day, Quantity: 10, UnitPrice: 20, ExchangeRate: 1, Ledger: true},
		{Type: "split", Symbol: "TEST.US", Currency: "USD", Date: day.AddDate(0, 0, 1), SplitNumerator: 3, SplitDenominator: 2, ExchangeRate: 1, Ledger: true},
		{Type: "sell", Symbol: "TEST.US", Currency: "USD", Date: day.AddDate(0, 0, 2), Quantity: 3, UnitPrice: 16, ExchangeRate: 1, Ledger: true},
	}
	s, err := Replay(a, events, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if s.Positions["TEST.US"].Quantity != 12 || s.Positions["TEST.US"].Cost != 160 || s.Cash != 848 || s.Realized != 8 {
		t.Fatalf("unexpected split ledger: %+v %+v", s, s.Positions["TEST.US"])
	}
	events[1].SplitDenominator = 0
	if _, err := Replay(a, events, time.Now()); err == nil {
		t.Fatal("invalid ratio accepted")
	}
}
