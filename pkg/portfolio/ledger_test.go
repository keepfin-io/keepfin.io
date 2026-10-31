package portfolio

import (
	"keepnode/app/models"
	"math"
	"testing"
	"time"
)

func TestReplay(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	account := models.Account{Currency: "USD", InitialBalance: 1000, OpeningDate: &now}
	events := []models.Activity{
		{Ledger: true, Type: "buy", Symbol: "ABC.US", Currency: "USD", Date: now, Quantity: 10, UnitPrice: 10, Fee: 2, ExchangeRate: 1},
		{Ledger: true, Type: "buy", Symbol: "ABC.US", Currency: "USD", Date: now.Add(time.Hour), Quantity: 10, UnitPrice: 20, Fee: 2, ExchangeRate: 1},
		{Ledger: true, Type: "sell", Symbol: "ABC.US", Currency: "USD", Date: now.Add(2 * time.Hour), Quantity: 5, UnitPrice: 30, Fee: 1, ExchangeRate: 1},
		{Ledger: true, Type: "dividend", Currency: "USD", Date: now.Add(3 * time.Hour), Quantity: 1, UnitPrice: 10, ExchangeRate: 1},
	}
	s, err := Replay(account, events, now.AddDate(0, 0, 1))
	if err != nil {
		t.Fatal(err)
	}
	for name, pair := range map[string][2]float64{"cash": {s.Cash, 855}, "quantity": {s.Positions["ABC.US"].Quantity, 15}, "cost": {s.Positions["ABC.US"].Cost, 228}, "realized": {s.Realized, 73}, "income": {s.Dividends, 10}} {
		if math.Abs(pair[0]-pair[1]) > 1e-8 {
			t.Errorf("%s got %v want %v", name, pair[0], pair[1])
		}
	}
	events[2].Quantity = 21
	if _, err := Replay(account, events, now.AddDate(0, 0, 1)); err == nil {
		t.Fatal("oversell allowed")
	}
}

func TestOpeningAndFX(t *testing.T) {
	now := time.Now().UTC()
	a := models.Account{Currency: "USD", InitialBalance: 100, OpeningDate: &now}
	events := []models.Activity{
		{Ledger: false, Type: "buy", Symbol: "ABC", Date: now, Quantity: 100, UnitPrice: 99},
		{Ledger: true, Type: "opening", Symbol: "ABC", Currency: "EUR", Date: now, Quantity: 10, UnitPrice: 10, ExchangeRate: 1.2},
		{Ledger: true, Type: "sell", Symbol: "ABC", Currency: "EUR", Date: now.Add(time.Second), Quantity: 5, UnitPrice: 12, ExchangeRate: 1.3},
	}
	s, err := Replay(a, events, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if s.Cash != 178 || s.Realized != 18 || s.Positions["ABC"].Quantity != 5 {
		t.Fatalf("unexpected state: %+v", s)
	}
}
