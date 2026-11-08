package holding_test

import (
	"context"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"

	holdingmodel "keepnode/app/models/holding"
	holdingservice "keepnode/app/services/holding"
	"keepnode/pkg/worker"
)

func TestSearchWorkerDrainsBacklogAndWakesOnCommittedChanges(t *testing.T) {
	db := searchDB(t)
	client := searchClient(t, os.Getenv("HOLDINGS_SEARCH_TEST_URL"))
	collection := searchCollection(t, client)
	index := holdingservice.NewSearchService(db, client, collection)
	user, account := searchAccount(t, db)
	// More than one 25-account batch must drain without a one-minute pause.
	for range 25 {
		searchAccount(t, db)
	}
	ctx, cancel := context.WithCancel(context.Background())
	workerDone, listenerDone := make(chan struct{}), make(chan struct{})
	go func() { defer close(workerDone); index.Run(ctx) }()
	go func() {
		defer close(listenerDone)
		worker.Listen(ctx, db.Dialector.(*postgres.Dialector).Config.DSN, map[string]func(){worker.HoldingsChanged: index.Wake})
	}()
	t.Cleanup(func() { cancel(); <-listenerDone; <-workerDone })
	waitIndexed := func(check func() bool) {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if check() {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatal("index worker waited for fallback instead of draining or reacting to a notification")
	}
	waitIndexed(func() bool {
		var pending int64
		return db.Model(&holdingmodel.SearchState{}).Where("revision > indexed_revision").Count(&pending).Error == nil && pending == 0
	})
	before := searchState(t, db, account.ID).Revision
	holding, err := holdingservice.NewHoldingService(db, nil).AddHolding(account.ID, &holdingmodel.CreateHoldingRequest{
		Symbol: "WAKE.US", Name: "Wake test", Quantity: 3, BuyPrice: 10, BuyCurrency: "USD",
	})
	if err != nil {
		t.Fatal(err)
	}
	// No FlushUser/Search call here: only the committed database notification can
	// wake this idle worker before its one-minute reconciliation deadline.
	waitIndexed(func() bool {
		state := searchState(t, db, account.ID)
		return state.Revision > before && state.IndexedRevision == state.Revision
	})
	page, err := index.Search(ctx, user, account.ID, searchParams())
	if err != nil || page.Source != "typesense" || page.Total != 1 || page.Documents[0].ID != holding.ID.String() {
		t.Fatalf("committed holding was not indexed: %+v %v", page, err)
	}
}
