package worker

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	HoldingsChanged     = "keepfin_holdings_changed"
	MarketJobsChanged   = "keepfin_market_jobs_changed"
	MarketConfigChanged = "keepfin_market_config_changed"
)

// Listen uses one dedicated connection for all workers. Notifications contain
// only a schema name; workers always read authoritative state from the database.
// Missed notifications are covered by startup scans and periodic reconciliation.
func Listen(ctx context.Context, dsn string, handlers map[string]func()) {
	delay := 5 * time.Second
	for ctx.Err() == nil {
		started := time.Now()
		err := listen(ctx, dsn, handlers)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			// Connection errors can contain DSN details. Do not log credentials.
			log.Printf("Worker notifications disconnected; reconnecting in %s (periodic reconciliation remains active)", delay)
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		if time.Since(started) > ReconcileInterval {
			delay = 5 * time.Second
		} else {
			delay = min(delay*2, ReconcileInterval)
		}
	}
}

func listen(ctx context.Context, dsn string, handlers map[string]func()) error {
	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	conn, err := pgx.Connect(connectCtx, dsn)
	if err != nil {
		return err
	}
	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), time.Second)
		defer closeCancel()
		_ = conn.Close(closeCtx)
	}()
	var schema string
	if err := conn.QueryRow(connectCtx, "SELECT current_schema()").Scan(&schema); err != nil {
		return err
	}
	for channel := range handlers {
		if _, err := conn.Exec(connectCtx, "LISTEN "+pgx.Identifier{channel}.Sanitize()); err != nil {
			return err
		}
	}
	// Subscribe first, then inspect state: closes the startup/reconnect race.
	for _, notify := range handlers {
		notify()
	}
	for {
		notification, err := conn.WaitForNotification(ctx)
		if err != nil {
			return err
		}
		if notify := handlers[notification.Channel]; notify != nil && notification.Payload == schema {
			notify()
		}
	}
}
