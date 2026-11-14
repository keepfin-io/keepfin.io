package wealth

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	accountmodel "keepnode/app/models/account"
	activitymodel "keepnode/app/models/activity"
	holdingmodel "keepnode/app/models/holding"
	usermodel "keepnode/app/models/user"
	wealthmodel "keepnode/app/models/wealth"
	activityservice "keepnode/app/services/activity"
	marketservice "keepnode/app/services/market"
	portfolioservice "keepnode/app/services/portfolio"
	"keepnode/pkg/portfolio"
	"keepnode/platform/database"
	"math"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func wealthDB(t *testing.T) *gorm.DB {
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
func TestWealthOwnershipValuationAndHistory(t *testing.T) {
	db := wealthDB(t)
	owner := fixtureUser(t, db)
	other := fixtureUser(t, db)
	day := portfolio.Day(time.Now()).AddDate(0, 0, -30)
	date := day.Format("2006-01-02")
	account := fixtureAccount(t, db, owner.ID, "USD", 1000, day)
	activities := activityservice.NewActivityService(db, nil)
	service := New(db, marketservice.NewMarketService(), activities, portfolioservice.NewPortfolioService(db, nil))
	asset, err := service.SaveAsset(owner.ID, nil, wealthmodel.ManualAsset{Name: "Private asset", Currency: "USD", AssetClass: "Private equity"})
	if err != nil {
		t.Fatal(err)
	}
	if err = service.SaveValuation(owner.ID, asset.ID, date, 15); err != nil {
		t.Fatal(err)
	}
	if err = service.SaveValuation(other.ID, asset.ID, date, 99); err == nil {
		t.Fatal("foreign valuation accepted")
	}
	buy := createEvent(t, activities, account, day, "buy", asset.Symbol, 10, 10, 0)
	if err = service.DeleteAsset(owner.ID, asset.ID); err == nil {
		t.Fatal("deleted asset with history")
	}
	foreignAccount := fixtureAccount(t, db, other.ID, "USD", 1000, day)
	if _, err = activities.CreateActivity(foreignAccount.ID, nil, &activitymodel.CreateActivityRequest{Type: "buy", Date: date, Symbol: asset.Symbol, Quantity: 1, UnitPrice: 15, Currency: "USD"}); err == nil {
		t.Fatal("foreign manual asset purchase")
	}
	debt, err := service.CreateLiability(owner.ID, "Loan", "USD", date, 200)
	if err != nil {
		t.Fatal(err)
	}
	if err = service.SaveLiabilityBalance(other.ID, debt.ID, date, 0); err == nil {
		t.Fatal("foreign liability edit")
	}
	r, err := service.Portfolio.Report(owner.ID, nil, "max", "")
	if err != nil {
		t.Fatal(err)
	}
	if !r.Complete || r.Summary.NetWorth == nil {
		t.Fatalf("incomplete %+v", r.Warnings)
	}
	closeTo(t, "value", r.Summary.Value, 1050)
	closeTo(t, "debt", *r.Summary.Liabilities, 200)
	closeTo(t, "net worth", *r.Summary.NetWorth, 850)
	if len(r.History) != 31 {
		t.Fatalf("history length %d", len(r.History))
	}
	if err = service.SaveLiabilityBalance(owner.ID, debt.ID, time.Now().UTC().Format("2006-01-02"), 0); err != nil {
		t.Fatal(err)
	}
	debts, err := service.Liabilities(owner.ID)
	if err != nil || len(debts) != 1 || len(debts[0].Balances) != 2 {
		t.Fatalf("liability history: %+v %v", debts, err)
	}
	tag, err := service.SaveTag(owner.ID, nil, "Long term", "")
	if err != nil {
		t.Fatal(err)
	}
	if err = service.AssignTags(owner.ID, account.ID, "account", []uuid.UUID{tag.ID}); err != nil {
		t.Fatal(err)
	}
	if err = service.AssignTags(other.ID, buy.ID, "activity", []uuid.UUID{tag.ID}); err == nil {
		t.Fatal("foreign tag target")
	}
	if tag.Color != "gray" {
		t.Fatalf("default tag color %q", tag.Color)
	}
	if _, err = service.SaveTag(owner.ID, nil, "Styled", "#ff0000"); err == nil {
		t.Fatal("raw CSS tag color accepted")
	}
	if recolored, e := service.SaveTag(owner.ID, &tag.ID, "Long term", "green"); e != nil || recolored.Color != "green" {
		t.Fatalf("recolor: %+v %v", recolored, e)
	}
	// Renaming without a color keeps the chosen color.
	if renamed, e := service.SaveTag(owner.ID, &tag.ID, "Long hold", ""); e != nil || renamed.Color != "green" || renamed.Name != "Long hold" {
		t.Fatalf("rename: %+v %v", renamed, e)
	}
	byAccount, err := service.TagsByEntity(owner.ID, "account")
	if err != nil || len(byAccount[account.ID]) != 1 || byAccount[account.ID][0].Color != "green" {
		t.Fatalf("account tags: %+v %v", byAccount, err)
	}
	if leaked, e := service.TagsByEntity(other.ID, "account"); e != nil || len(leaked) != 0 {
		t.Fatalf("foreign account tags leaked %+v %v", leaked, e)
	}
	if _, err = service.TagsByEntity(owner.ID, "holding"); err == nil {
		t.Fatal("unsupported tag target accepted")
	}
	page, err := service.AllActivities(owner.ID, ActivityFilter{TagID: &tag.ID, PerPage: 1, Query: "Private"})
	if err != nil || page.Total != 1 {
		t.Fatalf("filter: %+v %v", page, err)
	}
	foreign, err := service.AllActivities(other.ID, ActivityFilter{TagID: &tag.ID})
	if err != nil || foreign.Total != 0 {
		t.Fatalf("foreign tags leaked %+v %v", foreign, err)
	}
	document, err := activities.ExportJSON(account.ID)
	if err != nil {
		t.Fatal(err)
	}
	imported, err := activities.ImportJSON(account.ID, string(document), true)
	if err != nil || !imported.Valid || imported.Imported != 0 {
		t.Fatalf("JSON duplicate import %+v %v", imported, err)
	}
	splitDate := day.AddDate(0, 0, 1).Format("2006-01-02")
	split, err := activities.CreateActivity(account.ID, nil, &activitymodel.CreateActivityRequest{Type: "split", Date: splitDate, Symbol: asset.Symbol, Currency: "USD", SplitNumerator: 2, SplitDenominator: 1})
	if err != nil {
		t.Fatal(err)
	}
	var position holdingmodel.Holding
	if err = db.First(&position, "account_id = ? AND symbol = ?", account.ID, asset.Symbol).Error; err != nil {
		t.Fatal(err)
	}
	closeTo(t, "split quantity", position.Quantity, 20)
	closeTo(t, "split cost", position.Quantity*position.BuyPrice, 100)
	_, err = activities.CreateActivity(account.ID, nil, &activitymodel.CreateActivityRequest{Type: "split", Date: splitDate, Symbol: asset.Symbol, Currency: "USD", SplitNumerator: 2, SplitDenominator: 1})
	if err == nil {
		t.Fatal("duplicate split accepted")
	}
	if err = activities.DeleteActivity(split.ID, account.ID); err != nil {
		t.Fatal(err)
	}
}

func TestSharesScopeRecipientExpiryAndRevocation(t *testing.T) {
	db := wealthDB(t)
	owner := fixtureUser(t, db)
	recipient := fixtureUser(t, db)
	stranger := fixtureUser(t, db)
	auth := "test-auth"
	if err := db.Model(&recipient).Update("auth_user_id", auth).Error; err != nil {
		t.Fatal(err)
	}
	day := portfolio.Day(time.Now()).AddDate(0, 0, -2)
	selected := fixtureAccount(t, db, owner.ID, "USD", 100, day)
	fixtureAccount(t, db, owner.ID, "USD", 900, day)
	service := New(db, marketservice.NewMarketService(), activityservice.NewActivityService(db, nil), portfolioservice.NewPortfolioService(db, nil))
	_, err := service.CreateLiability(owner.ID, "Private loan", "USD", day.Format("2006-01-02"), 50)
	if err != nil {
		t.Fatal(err)
	}
	share, token, err := service.CreateShare(owner.ID, ShareRequest{Name: "Selected account", AccountIDs: []uuid.UUID{selected.ID}, ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if token == share.TokenHash || len(token) != 64 {
		t.Fatal("share stores plaintext token")
	}
	r, err := service.SharedReport(token, nil, "max")
	if err != nil {
		t.Fatal(err)
	}
	closeTo(t, "scope", r.Summary.Value, 100)
	if r.Summary.Liabilities != nil || r.Summary.NetWorth != nil {
		t.Fatal("personal liabilities leaked")
	}
	if err = service.RevokeShare(stranger.ID, share.ID); err == nil {
		t.Fatal("foreign revocation")
	}
	if err = service.RevokeShare(owner.ID, share.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = service.SharedReport(token, nil, "max"); err == nil {
		t.Fatal("revoked link accepted")
	}
	share, token, err = service.CreateShare(owner.ID, ShareRequest{Name: "Private", AccountIDs: []uuid.UUID{selected.ID}, RecipientEmail: recipient.Email, ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	for _, viewer := range []*uuid.UUID{nil, &stranger.ID} {
		if _, err = service.SharedReport(token, viewer, "max"); err == nil {
			t.Fatal("private share leaked")
		}
	}
	if _, err = service.SharedReport(token, &recipient.ID, "max"); err != nil {
		t.Fatal(err)
	}
	if err = db.Model(share).Update("expires_at", time.Now().Add(-time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	if _, err = service.SharedReport(token, &recipient.ID, "max"); err == nil {
		t.Fatal("expired share accepted")
	}
}

func TestWatchlistPreferencesAndJobsIsolation(t *testing.T) {
	db := wealthDB(t)
	owner := fixtureUser(t, db)
	other := fixtureUser(t, db)
	service := New(db, marketservice.NewMarketService(), activityservice.NewActivityService(db, nil), portfolioservice.NewPortfolioService(db, nil))
	item, err := service.AddWatch(owner.ID, wealthmodel.WatchItem{Symbol: "ABC.US", Currency: "USD", Name: "ABC"})
	if err != nil {
		t.Fatal(err)
	}
	again, err := service.AddWatch(owner.ID, wealthmodel.WatchItem{Symbol: "ABC.US", Currency: "USD", Name: "ABC"})
	if err != nil || again.ID != item.ID {
		t.Fatal("watchlist duplicate", err)
	}
	if err = service.DeleteWatch(other.ID, item.ID); err == nil {
		t.Fatal("foreign watch deletion")
	}
	p := DefaultPreferences(owner.ID)
	p.Zen = true
	p.MonthlyContribution = 100
	if _, err = service.SavePreferences(owner.ID, p); err != nil {
		t.Fatal(err)
	}
	loaded, err := service.Preferences(owner.ID)
	if err != nil || !loaded.Zen {
		t.Fatal("preferences not persisted", err)
	}
	p.WithdrawalRate = 0
	if _, err = service.SavePreferences(owner.ID, p); err == nil {
		t.Fatal("invalid withdrawal rate")
	}
	date := portfolio.Day(time.Now()).AddDate(0, 0, -1).Format("2006-01-02")
	job, err := service.QueueJob(owner.ID, "ABC.US", date, date)
	if err != nil {
		t.Fatal(err)
	}
	againJob, err := service.QueueJob(owner.ID, "ABC.US", date, date)
	if err != nil || againJob.ID != job.ID {
		t.Fatal("duplicate job", err)
	}
	if _, err = service.QueueJob(other.ID, "ABC.US", date, date); err == nil {
		t.Fatal("unowned symbol job")
	}
	jobs, err := service.Jobs(other.ID)
	if err != nil || len(jobs) != 0 {
		t.Fatal("foreign jobs", err)
	}
	if err = service.CorrectPrice(owner.ID, wealthmodel.PriceOverride{Symbol: "ABC.US", Date: date, Close: 12, Reason: "Test"}); err == nil {
		t.Fatal("nonadmin correction")
	}
}
