package holding_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	holdingcontroller "keepnode/app/controllers/holding"
	accountmodel "keepnode/app/models/account"
	holdingmodel "keepnode/app/models/holding"
	usermodel "keepnode/app/models/user"
	accountservice "keepnode/app/services/account"
	holdingservice "keepnode/app/services/holding"
	"keepnode/pkg/authhandler"
	"keepnode/pkg/typesense"
	"keepnode/platform/database"
)

func searchDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("PORTFOLIO_TEST_DATABASE_URL")
	if dsn == "" || os.Getenv("HOLDINGS_SEARCH_TEST_URL") == "" {
		t.Skip("PostgreSQL and Typesense test URLs are required")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	schema := "holding_search_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err = admin.Exec(fmt.Sprintf("CREATE SCHEMA %q", schema)).Error; err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	db, err := gorm.Open(postgres.Open(u.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool, _ := db.DB()
		_ = pool.Close()
		if e := admin.Exec(fmt.Sprintf("DROP SCHEMA %q CASCADE", schema)).Error; e != nil {
			t.Error(e)
		}
		pool, _ = admin.DB()
		_ = pool.Close()
	})
	if err = database.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func searchAccount(t *testing.T, db *gorm.DB) (uuid.UUID, accountmodel.Account) {
	t.Helper()
	user := usermodel.User{Email: uuid.NewString() + "@test.invalid", BaseCurrency: "USD"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	account := accountmodel.Account{UserID: user.ID, AccountName: "Search test", Currency: "USD", InitialBalance: 100000, LedgerReady: true, IsActive: true}
	if err := db.Create(&account).Error; err != nil {
		t.Fatal(err)
	}
	return user.ID, account
}

func searchClient(t *testing.T, address string) *typesense.Client {
	t.Helper()
	client, err := typesense.New(address, os.Getenv("TYPESENSE_API_KEY"))
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func searchCollection(t *testing.T, client *typesense.Client) string {
	t.Helper()
	name := "holding_search_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	t.Cleanup(func() {
		_, err := client.Request(context.Background(), http.MethodDelete, "/collections/"+name, nil, nil)
		if err != nil && !typesense.IsStatus(err, 404) {
			t.Error(err)
		}
	})
	return name
}

func searchParams() holdingmodel.SearchParams {
	return holdingmodel.SearchParams{Page: 1, PerPage: 10, Sort: "symbol", Direction: "asc"}
}

func searchState(t *testing.T, db *gorm.DB, account uuid.UUID) holdingmodel.SearchState {
	t.Helper()
	var state holdingmodel.SearchState
	if err := db.First(&state, "account_id = ?", account).Error; err != nil {
		t.Fatal(err)
	}
	return state
}

func TestHoldingsSearchLifecycle(t *testing.T) {
	db := searchDB(t)
	client := searchClient(t, os.Getenv("HOLDINGS_SEARCH_TEST_URL"))
	collection := searchCollection(t, client)
	index := holdingservice.NewSearchService(db, client, collection)
	writer := holdingservice.NewHoldingService(db, nil)
	user, account := searchAccount(t, db)
	otherUser, otherAccount := searchAccount(t, db)
	ctx := context.Background()
	var first *holdingmodel.Holding
	for i := 0; i < 23; i++ {
		h, err := writer.AddHolding(account.ID, &holdingmodel.CreateHoldingRequest{
			Symbol: fmt.Sprintf("ASSET%02d.US", i), Name: fmt.Sprintf("Aurora Fund %02d", i),
			AssetClass: "ETF", Country: "United States", Quantity: float64(i + 1), BuyPrice: 10, BuyCurrency: "USD",
		})
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first = h
		}
	}
	_, err := writer.AddHolding(otherAccount.ID, &holdingmodel.CreateHoldingRequest{Symbol: "PRIVATE.US", Name: "Private Fund", Quantity: 1, BuyPrice: 5, BuyCurrency: "USD"})
	if err != nil {
		t.Fatal(err)
	}
	if err = index.FlushUser(ctx, user); err != nil {
		t.Fatal(err)
	}
	if err = index.FlushUser(ctx, otherUser); err != nil {
		t.Fatal(err)
	}
	params := searchParams()
	page, err := index.Search(ctx, user, account.ID, params)
	if err != nil || page.Source != "typesense" || page.Total != 23 || len(page.Documents) != 10 || page.Pages != 3 {
		t.Fatalf("initial index page: %+v, error %v", page, err)
	}
	if page.Documents[0].Name != "Aurora Fund 00" || page.Documents[0].AssetClass != "ETF" {
		t.Fatal("asset metadata not persisted")
	}
	params.Page = 3
	page, err = index.Search(ctx, user, account.ID, params)
	if err != nil || len(page.Documents) != 3 || page.Documents[0].Symbol != "ASSET20.US" {
		t.Fatalf("last page: %+v %v", page, err)
	}
	params = searchParams()
	params.Query = "Aurorra"
	params.Country = "United States"
	params.AssetClass = "ETF"
	page, err = index.Search(ctx, user, account.ID, params)
	if err != nil || page.Source != "typesense" || page.Total != 23 {
		t.Fatalf("typo and facets: %+v %v", page, err)
	}
	params.Country = "Canada"
	page, err = index.Search(ctx, user, account.ID, params)
	if err != nil || page.Total != 0 {
		t.Fatal("exact facet did not filter")
	}
	page, err = index.Search(ctx, otherUser, account.ID, searchParams())
	if err != nil || page.Total != 0 {
		t.Fatal("cross-user documents leaked")
	}
	quantity := 42.0
	if _, err = writer.UpdateHolding(first.ID, account.ID, &holdingmodel.UpdateHoldingRequest{Quantity: &quantity}); err != nil {
		t.Fatal(err)
	}
	// A read repairs a pending revision even before the retry worker runs.
	page, err = index.Search(ctx, user, account.ID, searchParams())
	if err != nil || page.Source != "typesense" || page.Documents[0].Quantity != 42 {
		t.Fatalf("updated quantity: %+v %v", page, err)
	}
	before := searchState(t, db, account.ID)
	rollback := errors.New("rollback")
	err = db.Transaction(func(tx *gorm.DB) error {
		if e := tx.Model(first).Update("quantity", 999).Error; e != nil {
			return e
		}
		return rollback
	})
	if !errors.Is(err, rollback) || searchState(t, db, account.ID).Revision != before.Revision {
		t.Fatal("rolled back write dirtied index")
	}
	if err = db.Model(first).Update("quantity", 0).Error; err != nil {
		t.Fatal(err)
	}
	page, err = index.Search(ctx, user, account.ID, searchParams())
	if err != nil || page.Total != 22 || page.Documents[0].Symbol == first.Symbol {
		t.Fatal("closed position remains searchable")
	}
	// Losing the collection requeues other accounts as well as the current read.
	if _, err = client.Request(ctx, http.MethodDelete, "/collections/"+collection, nil, nil); err != nil {
		t.Fatal(err)
	}
	page, err = index.Search(ctx, user, account.ID, searchParams())
	if err != nil || page.Source != "typesense" || page.Total != 22 {
		t.Fatalf("collection recovery: %+v %v", page, err)
	}
	if state := searchState(t, db, otherAccount.ID); state.IndexedRevision == state.Revision {
		t.Fatal("other account was not requeued after index loss")
	}
	if err = db.Delete(&account).Error; err != nil {
		t.Fatal(err)
	}
	if err = index.FlushUser(ctx, user); err != nil {
		t.Fatal(err)
	}
	page, err = index.Search(ctx, user, account.ID, searchParams())
	if err != nil || page.Total != 0 {
		t.Fatal("deleted account remains searchable")
	}
}

func TestHoldingsSearchOutageAndRestartRecovery(t *testing.T) {
	db := searchDB(t)
	address := os.Getenv("HOLDINGS_SEARCH_TEST_URL")
	direct := searchClient(t, address)
	collection := searchCollection(t, direct)
	target, _ := url.Parse(address)
	proxy := httputil.NewSingleHostReverseProxy(target)
	var failure atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if failure.Load() == 1 {
			w.WriteHeader(503)
			return
		}
		if failure.Load() == 2 && strings.HasSuffix(r.URL.Path, "/import") {
			_, _ = w.Write([]byte(`{"success":false}`))
			return
		}
		proxy.ServeHTTP(w, r)
	}))
	defer server.Close()
	client := searchClient(t, server.URL)
	index := holdingservice.NewSearchService(db, client, collection)
	user, account := searchAccount(t, db)
	writer := holdingservice.NewHoldingService(db, nil)
	ctx := context.Background()
	if err := index.FlushUser(ctx, user); err != nil {
		t.Fatal(err)
	}
	failure.Store(1)
	h, err := writer.AddHolding(account.ID, &holdingmodel.CreateHoldingRequest{Symbol: "NEW.US", Name: "New Fund", Quantity: 7, BuyPrice: 10, BuyCurrency: "USD"})
	if err != nil {
		t.Fatal(err)
	}
	if err = index.FlushUser(ctx, user); err == nil {
		t.Fatal("outage should leave durable work pending")
	}
	page, err := index.Search(ctx, user, account.ID, searchParams())
	if err != nil || page.Source != "database" || page.Total != 1 || page.Documents[0].ID != h.ID.String() {
		t.Fatalf("fresh fallback missing: %+v %v", page, err)
	}
	failure.Store(2)
	if err = index.FlushUser(ctx, user); err == nil {
		t.Fatal("partial import cannot be acknowledged")
	}
	if state := searchState(t, db, account.ID); state.IndexedRevision >= state.Revision {
		t.Fatal("failed import advanced outbox")
	}
	failure.Store(0)
	workerCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() { defer close(done); holdingservice.NewSearchService(db, client, collection).Run(workerCtx) }()
	defer func() { cancel(); <-done }()
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		state := searchState(t, db, account.ID)
		if state.IndexedRevision == state.Revision {
			break
		}
		select {
		case <-deadline.C:
			t.Fatal("restart did not recover pending index")
		case <-ticker.C:
		}
	}
	page, err = index.Search(ctx, user, account.ID, searchParams())
	if err != nil || page.Source != "typesense" || page.Total != 1 || page.Documents[0].Quantity != 7 {
		t.Fatalf("recovered search: %+v %v", page, err)
	}
}

func TestHoldingsSearchConcurrentWriteCannotBeOverwritten(t *testing.T) {
	db := searchDB(t)
	address := os.Getenv("HOLDINGS_SEARCH_TEST_URL")
	direct := searchClient(t, address)
	collection := searchCollection(t, direct)
	target, _ := url.Parse(address)
	proxy := httputil.NewSingleHostReverseProxy(target)
	entered, release := make(chan struct{}), make(chan struct{})
	var blocked atomic.Bool
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/import") && blocked.CompareAndSwap(true, false) {
			close(entered)
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
		}
		proxy.ServeHTTP(w, r)
	}))
	defer server.Close()
	defer unblock()
	index := holdingservice.NewSearchService(db, searchClient(t, server.URL), collection)
	user, account := searchAccount(t, db)
	h, err := holdingservice.NewHoldingService(db, nil).AddHolding(account.ID, &holdingmodel.CreateHoldingRequest{Symbol: "RACE.US", Quantity: 1, BuyPrice: 5, BuyCurrency: "USD"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = index.FlushUser(ctx, user); err != nil {
		t.Fatal(err)
	}
	if err = db.Model(h).Update("quantity", 5).Error; err != nil {
		t.Fatal(err)
	}
	blocked.Store(true)
	synced, written := make(chan error, 1), make(chan error, 1)
	go func() { synced <- index.SyncAccount(ctx, account.ID) }()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("index import did not start")
	}
	// The second committed revision must never be acknowledged by the older import.
	go func() {
		written <- db.WithContext(ctx).Model(&holdingmodel.Holding{}).Where("id = ?", h.ID).Update("quantity", 9).Error
	}()
	unblock()
	if err = <-synced; err != nil {
		t.Fatal(err)
	}
	if err = <-written; err != nil {
		t.Fatal(err)
	}
	if state := searchState(t, db, account.ID); state.IndexedRevision >= state.Revision {
		t.Fatal("old import acknowledged newer write")
	}
	if err = index.FlushUser(ctx, user); err != nil {
		t.Fatal(err)
	}
	page, err := index.Search(ctx, user, account.ID, searchParams())
	if err != nil || page.Source != "typesense" || len(page.Documents) != 1 || page.Documents[0].Quantity != 9 {
		t.Fatalf("concurrent update lost: %+v %v", page, err)
	}
}

func TestHoldingsSearchHTTPAccountBoundary(t *testing.T) {
	db := searchDB(t)
	client := searchClient(t, os.Getenv("HOLDINGS_SEARCH_TEST_URL"))
	index := holdingservice.NewSearchService(db, client, searchCollection(t, client))
	writer := holdingservice.NewHoldingService(db, nil)
	user, account := searchAccount(t, db)
	other, _ := searchAccount(t, db)
	if _, err := writer.AddHolding(account.ID, &holdingmodel.CreateHoldingRequest{Symbol: "HTTP.US", Name: "HTTP Fund", Quantity: 1, BuyPrice: 10, BuyCurrency: "USD"}); err != nil {
		t.Fatal(err)
	}
	controller := holdingcontroller.NewHoldingController(writer, accountservice.NewAccountService(db), index)
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		// Simulate the identity already verified by the session middleware.
		id := user
		if c.Get("X-Test-User") == "other" {
			id = other
		}
		c.Locals("user_id", id)
		return c.Next()
	})
	app.Get("/accounts/:id/holdings/search", authhandler.WithUser(controller.SearchHoldings))
	for _, tc := range []struct {
		name, query, identity string
		status                int
	}{
		{"own account", "", "self", 200},
		{"another user's account", "", "other", 404},
		{"invalid page size", "?per_page=7", "self", 400},
		{"invalid filter", "?currency=%60%20%7C%7C%20user_id:*", "self", 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/accounts/"+account.ID.String()+"/holdings/search"+tc.query, nil)
			req.Header.Set("X-Test-User", tc.identity)
			response, err := app.Test(req, 5000)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != tc.status {
				t.Fatalf("status %d, want %d", response.StatusCode, tc.status)
			}
			if tc.status == 200 && response.Header.Get("Cache-Control") != "no-store" {
				t.Fatal("financial results may not be cached")
			}
		})
	}
}
