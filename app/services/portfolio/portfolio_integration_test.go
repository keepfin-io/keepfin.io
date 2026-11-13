package portfolio

import (
	"fmt"
	"math"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	accountmodel "keepnode/app/models/account"
	activitymodel "keepnode/app/models/activity"
	holdingmodel "keepnode/app/models/holding"
	marketmodel "keepnode/app/models/market"
	usermodel "keepnode/app/models/user"
	activityquery "keepnode/app/queries/activity"
	accountservice "keepnode/app/services/account"
	activityservice "keepnode/app/services/activity"
	holdingservice "keepnode/app/services/holding"
	"keepnode/pkg/portfolio"
	"keepnode/platform/database"
)

func portfolioDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("PORTFOLIO_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PORTFOLIO_TEST_DATABASE_URL is required")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	schema := "portfolio_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err = admin.Exec(fmt.Sprintf("CREATE SCHEMA %q", schema)).Error; err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(dsn)
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	db, err := gorm.Open(postgres.Open(u.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		admin.Exec(fmt.Sprintf("DROP SCHEMA %q CASCADE", schema))
		a, _ := admin.DB()
		a.Close()
	})
	if err = database.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	return db
}
func fixtureAccount(t *testing.T, db *gorm.DB, user uuid.UUID, currency string, balance float64, day time.Time) accountmodel.Account {
	t.Helper()
	a := accountmodel.Account{UserID: user, AccountName: currency + uuid.NewString()[:4], Currency: currency, InitialBalance: balance, OpeningDate: &day, LedgerReady: true, IsActive: true}
	if err := db.Create(&a).Error; err != nil {
		t.Fatal(err)
	}
	return a
}
func fixtureUser(t *testing.T, db *gorm.DB) usermodel.User {
	t.Helper()
	u := usermodel.User{Email: uuid.NewString() + "@test.invalid", BaseCurrency: "USD"}
	if err := db.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	return u
}
func closeTo(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("%s got %.10f want %.10f", name, got, want)
	}
}
func createEvent(t *testing.T, s *activityservice.ActivityService, a accountmodel.Account, day time.Time, kind, symbol string, q, p, fee float64) *activitymodel.Activity {
	t.Helper()
	e, err := s.CreateActivity(a.ID, nil, &activitymodel.CreateActivityRequest{Date: day.Format("2006-01-02"), Type: kind, Symbol: symbol, Quantity: q, UnitPrice: p, Fee: fee, Currency: a.Currency})
	if err != nil {
		t.Fatal(err)
	}
	return e
}
func TestPortfolioLedgerDatabase(t *testing.T) {
	db := portfolioDB(t)
	u := fixtureUser(t, db)
	day := portfolio.Day(time.Now()).AddDate(0, 0, -3)
	a := fixtureAccount(t, db, u.ID, "USD", 1000, day)
	svc := activityservice.NewActivityService(db, nil)
	buy := createEvent(t, svc, a, day, "buy", "ABC.US", 10, 10, 2)
	createEvent(t, svc, a, day.AddDate(0, 0, 1), "buy", "ABC.US", 10, 20, 2)
	sell := createEvent(t, svc, a, day.AddDate(0, 0, 2), "sell", "ABC.US", 5, 30, 1)
	var h holdingmodel.Holding
	db.First(&h, "account_id = ?", a.ID)
	closeTo(t, "quantity", h.Quantity, 15)
	closeTo(t, "average cost", h.BuyPrice, 15.2)
	cash, err := accountCashForTest(db, a)
	if err != nil {
		t.Fatal(err)
	}
	closeTo(t, "cash", cash, 845)
	_, err = svc.UpdateActivity(buy.ID, a.ID, &activitymodel.CreateActivityRequest{Date: day.Format("2006-01-02"), Type: "buy", Symbol: "ABC.US", Quantity: 1, UnitPrice: 10, Currency: "USD"})
	if err != nil {
		t.Fatal(err)
	}
	if err = svc.DeleteActivity(buy.ID, a.ID); err != nil {
		t.Fatal(err)
	}
	// Removing the last remaining buy must not leave a short position.
	var other activitymodel.Activity
	db.Where("account_id = ? AND type = ?", a.ID, "buy").First(&other)
	if err = svc.DeleteActivity(other.ID, a.ID); err == nil {
		t.Fatal("deleted cost basis needed by a sale")
	}
	var count int64
	db.Model(&activitymodel.Activity{}).Where("id = ?", other.ID).Count(&count)
	if count != 1 {
		t.Fatal("failed delete was not rolled back")
	}
	if err = svc.DeleteActivity(sell.ID, a.ID); err != nil {
		t.Fatal(err)
	}
	// Two concurrent sales serialize under the account lock; only one can succeed.
	successes := make(chan bool, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, e := svc.CreateActivity(a.ID, nil, &activitymodel.CreateActivityRequest{Date: day.AddDate(0, 0, 2).Format("2006-01-02"), Type: "sell", Symbol: "ABC.US", Quantity: 7, UnitPrice: 30, Currency: "USD"})
			successes <- e == nil
		}()
	}
	wg.Wait()
	close(successes)
	n := 0
	for ok := range successes {
		if ok {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("concurrent oversell successes: %d", n)
	}
}
func TestPortfolioTransfersAndScope(t *testing.T) {
	db := portfolioDB(t)
	u := fixtureUser(t, db)
	other := fixtureUser(t, db)
	day := portfolio.Day(time.Now()).AddDate(0, 0, -1)
	a := fixtureAccount(t, db, u.ID, "USD", 1000, day)
	b := fixtureAccount(t, db, u.ID, "EUR", 0, day)
	foreign := fixtureAccount(t, db, other.ID, "USD", 0, day)
	svc := activityservice.NewActivityService(db, nil)
	req := activityservice.TransferRequest{FromAccountID: a.ID, ToAccountID: b.ID, Date: day.Format("2006-01-02"), Amount: 100, ReceivedAmount: 90, Fee: 2}
	if err := svc.Transfer(u.ID, req); err != nil {
		t.Fatal(err)
	}
	cash, _ := accountCashForTest(db, a)
	closeTo(t, "source cash", cash, 898)
	cash, _ = accountCashForTest(db, b)
	closeTo(t, "destination cash", cash, 90)
	req.ToAccountID = foreign.ID
	if err := svc.Transfer(u.ID, req); err == nil {
		t.Fatal("cross-user transfer accepted")
	}
	var events []activitymodel.Activity
	db.Where("account_id = ?", a.ID).Find(&events)
	if len(events) != 1 {
		t.Fatal("unauthorized transfer changed source")
	}
	if err := svc.DeleteActivity(events[0].ID, a.ID); err != nil {
		t.Fatal(err)
	}
	cash, _ = accountCashForTest(db, a)
	closeTo(t, "reversed cash", cash, 1000)
	cash, _ = accountCashForTest(db, b)
	closeTo(t, "reversed destination", cash, 0)
}
func TestPortfolioCSV(t *testing.T) {
	db := portfolioDB(t)
	u := fixtureUser(t, db)
	day := portfolio.Day(time.Now()).AddDate(0, 0, -1)
	a := fixtureAccount(t, db, u.ID, "USD", 1000, day)
	svc := activityservice.NewActivityService(db, nil)
	csv := "reference,date,type,symbol,quantity,unit_price,fee,currency,exchange_rate,notes\ntrade-1," + day.Format("2006-01-02") + ",buy,ABC.US,10,10,1,USD,1,=SUM(A1)\n"
	preview, err := svc.ImportCSV(a.ID, csv, false)
	if err != nil || !preview.Valid {
		t.Fatalf("preview: %+v %v", preview, err)
	}
	var count int64
	db.Model(&activitymodel.Activity{}).Count(&count)
	if count != 0 {
		t.Fatal("preview wrote activities")
	}
	result, err := svc.ImportCSV(a.ID, csv, true)
	if err != nil || result.Imported != 1 {
		t.Fatalf("import: %+v %v", result, err)
	}
	result, err = svc.ImportCSV(a.ID, csv, true)
	if err != nil || result.Imported != 0 || result.Duplicates != 1 {
		t.Fatalf("repeat: %+v %v", result, err)
	}
	exported, err := svc.ExportCSV(a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(exported), "'=SUM(A1)") {
		t.Fatal("spreadsheet formula not escaped")
	}
	result, err = svc.ImportCSV(a.ID, string(exported), true)
	if err != nil || result.Imported != 0 {
		t.Fatalf("export replay: %+v %v", result, err)
	}
	bad := strings.Replace(csv, "trade-1", "trade-2", 1) + "trade-3," + day.Format("2006-01-02") + ",sell,ABC.US,999,10,0,USD,1,bad\n"
	result, err = svc.ImportCSV(a.ID, bad, true)
	if err != nil || result.Valid {
		t.Fatalf("oversell import accepted: %+v %v", result, err)
	}
	db.Model(&activitymodel.Activity{}).Count(&count)
	if count != 1 {
		t.Fatal("invalid batch partially imported")
	}
}
func TestPortfolioReportHistory(t *testing.T) {
	db := portfolioDB(t)
	u := fixtureUser(t, db)
	day := portfolio.Day(time.Now()).AddDate(0, 0, -2)
	a := fixtureAccount(t, db, u.ID, "USD", 1000, day)
	b := fixtureAccount(t, db, u.ID, "USD", 0, day)
	svc := activityservice.NewActivityService(db, nil)
	createEvent(t, svc, a, day, "buy", "ABC.US", 10, 10, 0)
	createEvent(t, svc, a, day.AddDate(0, 0, 1), "deposit", "", 1, 100, 0)
	createEvent(t, svc, a, day.AddDate(0, 0, 2), "dividend", "ABC.US", 1, 10, 0)
	if err := svc.Transfer(u.ID, activityservice.TransferRequest{FromAccountID: a.ID, ToAccountID: b.ID, Date: day.AddDate(0, 0, 2).Format("2006-01-02"), Amount: 50, ReceivedAmount: 50}); err != nil {
		t.Fatal(err)
	}
	for i, price := range []float64{10, 11, 12} {
		for _, symbol := range []string{"ABC.US", "SPY.US"} {
			if err := db.Create(&marketmodel.MarketPrice{Symbol: symbol, Date: day.AddDate(0, 0, i).Format("2006-01-02"), Close: price}).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	report, err := NewPortfolioService(db, nil).Report(u.ID, nil, "max", "SPY.US")
	if err != nil {
		t.Fatal(err)
	}
	if !report.Complete {
		t.Fatalf("incomplete %+v", report.Warnings)
	}
	closeTo(t, "value", report.Summary.Value, 1130)
	closeTo(t, "contributions", report.Summary.NetContributions, 1100)
	closeTo(t, "profit", report.Summary.ProfitLoss, 30)
	closeTo(t, "dividends", report.Summary.Dividends, 10)
	if len(report.History) != 3 || report.History[1].Return == nil {
		t.Fatalf("history %+v", report.History)
	}
	closeTo(t, "deposit adjusted return", *report.History[1].Return, 10.0/1050*100)
	closeTo(t, "internal transfer flow", report.History[2].NetFlow, 0)
	closeTo(t, "benchmark", *report.History[2].Benchmark, 20)
	// Missing a historical price must produce a gap, never a synthetic buy-price valuation.
	db.Where("symbol = ?", "ABC.US").Delete(&marketmodel.MarketPrice{})
	report, err = NewPortfolioService(db, nil).Report(u.ID, nil, "max", "")
	if err != nil {
		t.Fatal(err)
	}
	if report.Complete || report.History[0].Value != nil {
		t.Fatal("missing prices produced a complete valuation")
	}
}
func TestPortfolioMigrationPreservesSnapshots(t *testing.T) {
	db := portfolioDB(t)
	u := fixtureUser(t, db)
	a := accountmodel.Account{UserID: u.ID, AccountName: "Legacy", Currency: "USD", InitialBalance: 500}
	if err := db.Create(&a).Error; err != nil {
		t.Fatal(err)
	}
	h := holdingmodel.Holding{AccountID: a.ID, Symbol: "ABC.US", Quantity: 10, BuyPrice: 12, BuyCurrency: "USD"}
	if err := db.Create(&h).Error; err != nil {
		t.Fatal(err)
	}
	old := activitymodel.Activity{AccountID: a.ID, Symbol: "ABC.US", Type: "buy", Date: time.Now().AddDate(0, -1, 0), Quantity: 20, UnitPrice: 8, Currency: "USD"}
	if err := db.Create(&old).Error; err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := database.AutoMigrate(db); err != nil {
			t.Fatal(err)
		}
	}
	db.First(&a, "id = ?", a.ID)
	events, _ := activityquery.GetLedgerEvents(db, a.ID)
	if len(events) != 1 {
		t.Fatalf("migration duplicated snapshots: %d", len(events))
	}
	state, err := portfolio.Replay(a, events, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	closeTo(t, "legacy cash", state.Cash, 500)
	closeTo(t, "legacy quantity", state.Positions["ABC.US"].Quantity, 10)
	closeTo(t, "legacy cost", state.Positions["ABC.US"].Cost, 120)
	var retained activitymodel.Activity
	if err = db.First(&retained, "id = ?", old.ID).Error; err != nil || retained.Ledger {
		t.Fatal("archived activity was changed")
	}
	createEvent(t, activityservice.NewActivityService(db, nil), a, portfolio.Day(time.Now()), "sell", "ABC.US", 2, 15, 0)
}

func TestPortfolioHistoricalFXAndConsolidation(t *testing.T) {
	db := portfolioDB(t)
	u := fixtureUser(t, db)
	day := portfolio.Day(time.Now()).AddDate(0, 0, -2)
	a := fixtureAccount(t, db, u.ID, "EUR", 1000, day)
	b := fixtureAccount(t, db, u.ID, "EUR", 0, day)
	svc := activityservice.NewActivityService(db, nil)
	createEvent(t, svc, a, day, "buy", "ABC.XETRA", 10, 10, 0)
	createEvent(t, svc, a, day.AddDate(0, 0, 1), "sell", "ABC.XETRA", 5, 20, 0)
	for i, rate := range []float64{1.2, 1.3, 1.4} {
		for symbol, price := range map[string]float64{"EURUSD.FOREX": rate, "ABC.XETRA": 25, "SPY.US": 100} {
			if err := db.Create(&marketmodel.MarketPrice{Symbol: symbol, Date: day.AddDate(0, 0, i).Format("2006-01-02"), Close: price}).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	report, err := NewPortfolioService(db, nil).Report(u.ID, nil, "max", "")
	if err != nil {
		t.Fatal(err)
	}
	closeTo(t, "historical base cost", report.Summary.Cost, 60)
	closeTo(t, "historical realized gain", report.Summary.Realized, 70)
	closeTo(t, "current value", report.Summary.Value, 1575)
	closeTo(t, "historical contributions", report.Summary.NetContributions, 1200)
	accountReport, err := NewPortfolioService(db, nil).Report(u.ID, &a.ID, "max", "SPY.US")
	if err != nil {
		t.Fatal(err)
	}
	closeTo(t, "FX benchmark", *accountReport.History[2].Benchmark, (1.2/1.4-1)*100)
	createEvent(t, svc, b, day, "buy", "ABC.XETRA", 2, 10, 0)
	report, err = NewPortfolioService(db, nil).Report(u.ID, nil, "max", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Holdings) != 1 || len(report.Holdings[0].Accounts) != 2 {
		t.Fatal("positions did not consolidate")
	}
	closeTo(t, "combined quantity", report.Holdings[0].Quantity, 7)
	for kind, rows := range report.Allocations {
		sum := 0.0
		for _, row := range rows {
			sum += row.Weight
		}
		closeTo(t, kind+" allocation", sum, 100)
	}
}
func TestPortfolioPrecisionAndCSVOrder(t *testing.T) {
	db := portfolioDB(t)
	u := fixtureUser(t, db)
	day := portfolio.Day(time.Now())
	a := fixtureAccount(t, db, u.ID, "USD", 1000, day)
	b := fixtureAccount(t, db, u.ID, "USD", 1000, day)
	svc := activityservice.NewActivityService(db, nil)
	_, err := svc.CreateActivity(a.ID, nil, &activitymodel.CreateActivityRequest{Date: day.Format("2006-01-02"), Type: "buy", Symbol: "ABC.US", Quantity: 1, UnitPrice: 0.00001, Currency: "USD"})
	if err == nil {
		t.Fatal("unpersistable precision accepted")
	}
	createEvent(t, svc, a, day, "buy", "ABC.US", 3, 10, 1)
	createEvent(t, svc, a, day, "sell", "ABC.US", 1, 12, 0)
	data, err := svc.ExportCSV(a.ID)
	if err != nil {
		t.Fatal(err)
	}
	result, err := svc.ImportCSV(b.ID, string(data), true)
	if err != nil || !result.Valid || result.Imported != 2 {
		t.Fatalf("same-day round trip: %+v %v", result, err)
	}
	cashA, _ := accountCashForTest(db, a)
	cashB, _ := accountCashForTest(db, b)
	closeTo(t, "round trip cash", cashA, cashB)
}

func accountCashForTest(db *gorm.DB, account accountmodel.Account) (float64, error) {
	current, err := accountservice.NewAccountService(db).GetAccountByID(account.ID, account.UserID)
	if err != nil {
		return 0, err
	}
	return current.CashBalance, nil
}

func TestPortfolioConvertsQuoteCurrency(t *testing.T) {
	db := portfolioDB(t)
	u := fixtureUser(t, db)
	day := portfolio.Day(time.Now()).AddDate(0, 0, -1)
	lira := fixtureAccount(t, db, u.ID, "TRY", 1000, day)
	sterling := fixtureAccount(t, db, u.ID, "GBP", 100, day)
	svc := activityservice.NewActivityService(db, nil)
	// Recorded before the quote currency was known: a USD listing bought in TRY.
	createEvent(t, svc, lira, day, "buy", "ABC.US", 2, 400, 0)
	for _, row := range []marketmodel.Instrument{{Symbol: "ABC.US", Currency: "USD"}, {Symbol: "VOD.LSE", Currency: "GBX"}} {
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		date := day.AddDate(0, 0, i).Format("2006-01-02")
		for symbol, close := range map[string]float64{"ABC.US": 10, "VOD.LSE": 70, "USDTRY.FOREX": 40, "GBPUSD.FOREX": 1.25} {
			if err := db.Create(&marketmodel.MarketPrice{Symbol: symbol, Date: date, Close: close}).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	// A pence-quoted listing is traded in pounds.
	createEvent(t, svc, sterling, day, "buy", "VOD.LSE", 10, 0.7, 0)
	if _, err := svc.CreateActivity(lira.ID, nil, &activitymodel.CreateActivityRequest{Date: day.Format("2006-01-02"), Type: "buy", Symbol: "ABC.US", Quantity: 1, UnitPrice: 400, Currency: "TRY"}); err == nil || !strings.Contains(err.Error(), "trades in USD") {
		t.Fatalf("TRY trade of a USD listing was accepted: %v", err)
	}
	if _, err := svc.CreateActivity(sterling.ID, nil, &activitymodel.CreateActivityRequest{Date: day.Format("2006-01-02"), Type: "buy", Symbol: "VOD.LSE", Quantity: 1, UnitPrice: 70, Currency: "GBX"}); err == nil {
		t.Fatal("minor-unit trade currency was accepted")
	}

	report, err := NewPortfolioService(db, nil).Report(u.ID, nil, "max", "")
	if err != nil {
		t.Fatal(err)
	}
	if !report.Complete {
		t.Fatalf("incomplete %+v", report.Warnings)
	}
	// TRY: 200 cash / 40 + 2 x 10 USD. GBP: (93 cash + 10 x 70p) x 1.25.
	closeTo(t, "value", report.Summary.Value, 5+20+116.25+8.75)
	for _, h := range report.Holdings {
		switch h.Symbol {
		case "ABC.US":
			closeTo(t, "ABC price in TRY", h.Price, 400)
		case "VOD.LSE":
			closeTo(t, "VOD price in GBP", h.Price, 0.7)
		}
	}

	rows, summary, err := holdingservice.NewHoldingService(db, nil).GetHoldingsWithLiveData(lira.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].QuoteCurrency != "USD" || !rows[0].PriceAvailable {
		t.Fatalf("holdings %+v", rows)
	}
	closeTo(t, "row price", rows[0].CurrentPrice, 400)
	closeTo(t, "row value", rows[0].Value, 800)
	closeTo(t, "row profit", rows[0].ProfitLoss, 0)
	closeTo(t, "account equity", summary.TotalValue, 800)
}
