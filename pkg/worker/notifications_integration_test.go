package worker

import (
	"context"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestListenerReconnectsAndRechecksDurableState(t *testing.T) {
	dsn := os.Getenv("PORTFOLIO_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PORTFOLIO_TEST_DATABASE_URL is required")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(context.Background())
	var schema string
	if err := admin.QueryRow(ctx, "SELECT current_schema()").Scan(&schema); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	application := "worker_test_" + uuid.NewString()
	query := u.Query()
	query.Set("application_name", application)
	u.RawQuery = query.Encode()
	channel := "worker_test_" + uuid.NewString()
	events := make(chan struct{}, 10)
	done := make(chan struct{})
	go func() {
		defer close(done)
		Listen(ctx, u.String(), map[string]func(){channel: func() { events <- struct{}{} }})
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("listener did not stop on cancellation")
		}
	})
	expectWake := func(timeout time.Duration) {
		t.Helper()
		select {
		case <-events:
		case <-time.After(timeout):
			t.Fatal("listener did not wake workers")
		}
	}
	expectWake(3 * time.Second) // LISTEN first, initial reconciliation second.
	if _, err := admin.Exec(ctx, "SELECT pg_notify($1, $2)", channel, "other_schema"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-events:
		t.Fatal("another schema woke this worker")
	case <-time.After(100 * time.Millisecond):
	}
	if _, err := admin.Exec(ctx, "SELECT pg_notify($1, $2)", channel, schema); err != nil {
		t.Fatal(err)
	}
	expectWake(time.Second)
	var terminated bool
	if err := admin.QueryRow(ctx, "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE application_name = $1", application).Scan(&terminated); err != nil || !terminated {
		t.Fatalf("could not simulate a disconnected listener: %v", err)
	}
	expectWake(8 * time.Second) // Reconnect must rescan even without a notification.
	if _, err := admin.Exec(ctx, "SELECT pg_notify($1, $2)", channel, schema); err != nil {
		t.Fatal(err)
	}
	expectWake(time.Second)
}
