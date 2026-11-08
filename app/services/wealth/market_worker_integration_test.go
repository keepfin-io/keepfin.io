package wealth

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/gorm"

	model "keepnode/app/models/wealth"
	marketservice "keepnode/app/services/market"
	"keepnode/pkg/worker"
)

func TestMarketProviderCacheAndInvalidation(t *testing.T) {
	db := wealthDB(t)
	s := New(db, marketservice.NewMarketService(), nil, nil)
	var queries atomic.Int64
	if err := db.Callback().Query().Before("gorm:query").Register("count_config_queries", func(tx *gorm.DB) {
		if tx.Statement.Table == "market_configs" {
			queries.Add(1)
		}
	}); err != nil {
		t.Fatal(err)
	}
	for range 3 {
		if err := s.LoadProvider(); err != nil {
			t.Fatal(err)
		}
	}
	if queries.Load() != 1 {
		t.Fatal("missing configuration must be cached too")
	}
	if err := db.Create(&model.MarketConfig{ID: 1, Provider: "alpha-vantage"}).Error; err != nil {
		t.Fatal(err)
	}
	s.InvalidateProvider()
	if err := s.LoadProvider(); err != nil || s.Market.Provider() != "alpha-vantage" || queries.Load() != 2 {
		t.Fatalf("configuration notification did not refresh provider: %v", err)
	}
	if err := db.Model(&model.MarketConfig{}).Where("id = 1").Update("provider", "eodhd").Error; err != nil {
		t.Fatal(err)
	}
	s.providerLoadedAt = time.Now().Add(-2 * worker.ReconcileInterval)
	if err := s.LoadProvider(); err != nil || s.Market.Provider() != "eodhd" || queries.Load() != 3 {
		t.Fatalf("cache expiry did not recover a missed notification: %v", err)
	}
}

func TestMarketEmptyQueueCommitsAbandonedJobRecovery(t *testing.T) {
	db := wealthDB(t)
	s := New(db, marketservice.NewMarketService(), nil, nil)
	user := fixtureUser(t, db)
	job := model.MarketJob{UserID: user.ID, Symbol: "ABC.US", From: "2026-01-01", To: "2026-01-02", Status: "running", Attempts: 3, UpdatedAt: time.Now().Add(-10 * time.Minute)}
	if err := db.Create(&job).Error; err != nil {
		t.Fatal(err)
	}
	if err := s.RunMarketJob(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&job, "id = ?", job.ID).Error; err != nil || job.Status != "failed" {
		t.Fatalf("empty claim rolled back recovery: status=%s error=%v", job.Status, err)
	}
	if processed, err := s.runMarketJob(context.Background()); err != nil || processed {
		t.Fatalf("empty queue is normal: %t %v", processed, err)
	}
}

func TestMarketWorkerWakesAndPreservesRetryDelay(t *testing.T) {
	t.Setenv("EODHD_API_TOKEN", "")
	t.Setenv("MARKET_DATA_PROVIDER", "eodhd")
	db := wealthDB(t)
	s := New(db, marketservice.NewMarketService(), nil, nil)
	user := fixtureUser(t, db)
	if err := db.Model(&user).Update("user_role", "admin").Error; err != nil {
		t.Fatal(err)
	}
	var queries atomic.Int64
	if err := db.Callback().Query().After("gorm:query").Register("count_worker_queries", func(tx *gorm.DB) {
		if tx.Statement.Table == "market_jobs" || tx.Statement.Table == "market_configs" {
			queries.Add(1)
		}
	}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); s.RunMarketWorker(ctx) }()
	t.Cleanup(func() { cancel(); <-done })
	waitMarket(t, func() bool { return queries.Load() >= 2 })
	before := queries.Load()
	time.Sleep(100 * time.Millisecond)
	if queries.Load() != before {
		t.Fatal("idle worker queried again without an event")
	}
	job, err := s.QueueJob(user.ID, "ABC.US", "2026-01-01", "2026-01-02")
	if err != nil {
		t.Fatal(err)
	}
	// A missing provider key fails locally, so no external request is needed to
	// verify retry scheduling.
	waitMarket(t, func() bool {
		var current model.MarketJob
		return db.First(&current, "id = ?", job.ID).Error == nil && current.Attempts == 1 && current.Status == "queued" && current.AvailableAt.After(time.Now())
	})
	s.WakeMarketWorker()
	time.Sleep(100 * time.Millisecond)
	if err := db.First(job, "id = ?", job.ID).Error; err != nil || job.Attempts != 1 {
		t.Fatalf("notification bypassed retry delay: attempts=%d err=%v", job.Attempts, err)
	}
}

func waitMarket(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("worker did not react before the fallback interval")
}
