package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	accountmodel "keepnode/app/models/account"
	model "keepnode/app/models/wealth"
	userservice "keepnode/app/services/user"
	"keepnode/pkg/authsession"
	"keepnode/pkg/worker"
)

func TestWorkerTriggersWithRestrictedRuntime(t *testing.T) {
	dsn := os.Getenv("SECURITY_RUNTIME_DATABASE_URL")
	if dsn == "" {
		t.Skip("run scripts/test-db-security.sh")
	}
	ctx := context.Background()
	listener, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close(ctx)
	for _, channel := range []string{worker.HoldingsChanged, worker.MarketJobsChanged, worker.MarketConfigChanged} {
		if _, err := listener.Exec(ctx, "LISTEN "+pgx.Identifier{channel}.Sanitize()); err != nil {
			t.Fatal(err)
		}
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	pool, _ := db.DB()
	defer pool.Close()
	user, err := userservice.NewUserService(db).ResolveIdentity(ctx, &authsession.Identity{ID: uuid.NewString(), Email: uuid.NewString() + "@worker.test", EmailVerified: true})
	if err != nil {
		t.Fatal(err)
	}
	wait := func(channel string) {
		t.Helper()
		waitCtx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		n, err := listener.WaitForNotification(waitCtx)
		if err != nil || n.Channel != channel || n.Payload != "public" {
			t.Fatalf("unexpected notification: %+v %v", n, err)
		}
	}
	quiet := func() {
		t.Helper()
		waitCtx, cancel := context.WithTimeout(ctx, 80*time.Millisecond)
		defer cancel()
		if n, err := listener.WaitForNotification(waitCtx); err == nil || n != nil || waitCtx.Err() == nil {
			t.Fatalf("expected no notification: %+v %v", n, err)
		}
	}
	account := accountmodel.Account{UserID: user.ID, AccountName: "Worker test", Currency: "USD"}
	tx := db.Begin()
	if err := tx.Create(&account).Error; err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	quiet() // Uncommitted work must not wake workers.
	if err := tx.Commit().Error; err != nil {
		t.Fatal(err)
	}
	wait(worker.HoldingsChanged)
	if err := db.Table("holding_search_states").Where("account_id = ?", account.ID).Update("indexed_revision", gorm.Expr("revision")).Error; err != nil {
		t.Fatal(err)
	}
	quiet() // Acknowledging work must not create a notification loop.
	tx = db.Begin()
	if err := tx.Table("holding_search_states").Where("account_id = ?", account.ID).Update("revision", gorm.Expr("revision + 1")).Error; err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Rollback().Error; err != nil {
		t.Fatal(err)
	}
	quiet()
	if err := db.Table("holding_search_states").Where("account_id = ?", account.ID).Update("revision", gorm.Expr("revision + 1")).Error; err != nil {
		t.Fatal(err)
	}
	wait(worker.HoldingsChanged)
	job := model.MarketJob{UserID: user.ID, Symbol: "ABC.US", Status: "queued", AvailableAt: time.Now()}
	if err := db.Create(&job).Error; err != nil {
		t.Fatal(err)
	}
	wait(worker.MarketJobsChanged)
	if err := db.Model(&job).Update("status", "running").Error; err != nil {
		t.Fatal(err)
	}
	quiet()
	if err := db.Model(&job).Updates(map[string]any{"status": "queued", "available_at": time.Now().Add(time.Minute)}).Error; err != nil {
		t.Fatal(err)
	}
	wait(worker.MarketJobsChanged)
	if err := db.Save(&model.MarketConfig{ID: 1, Provider: "eodhd"}).Error; err != nil {
		t.Fatal(err)
	}
	wait(worker.MarketConfigChanged)
}
