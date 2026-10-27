package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"
	"testing"

	"keepnode/app/models"
	"keepnode/pkg/authsession"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestResolveIdentityPostgres(t *testing.T) {
	dsn := os.Getenv("AUTH_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set AUTH_TEST_DATABASE_URL to a disposable PostgreSQL database")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	schema := "identity_test_" + uuid.New().String()[:8]
	if err := admin.Exec(fmt.Sprintf("CREATE SCHEMA %q", schema)).Error; err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := u.Query()
	query.Set("search_path", schema)
	u.RawQuery = query.Encode()
	db, err := gorm.Open(postgres.Open(u.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		admin.Exec(fmt.Sprintf("DROP SCHEMA %q CASCADE", schema))
		adminDB, _ := admin.DB()
		adminDB.Close()
	})
	if err := db.AutoMigrate(&models.User{}, &models.Account{}); err != nil {
		t.Fatal(err)
	}
	svc := NewUserService(db)
	ctx := context.Background()
	identity := &authsession.Identity{ID: uuid.NewString(), Email: "new@example.com", EmailVerified: true}
	var wg sync.WaitGroup
	ids := make(chan uuid.UUID, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			user, err := svc.ResolveIdentity(ctx, identity)
			if err != nil {
				t.Error(err)
				return
			}
			ids <- user.ID
		}()
	}
	wg.Wait()
	close(ids)
	var first uuid.UUID
	for id := range ids {
		if first == uuid.Nil {
			first = id
		}
		if id != first {
			t.Fatal("duplicate application identities")
		}
	}
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count != 1 {
		t.Fatalf("created %d users", count)
	}

	legacy := models.User{Email: "legacy@example.com", Theme: models.ThemeDark}
	if err := db.Create(&legacy).Error; err != nil {
		t.Fatal(err)
	}
	account := models.Account{UserID: legacy.ID, AccountName: "Existing savings", Currency: "USD", InitialBalance: 42}
	if err := db.Create(&account).Error; err != nil {
		t.Fatal(err)
	}
	legacyIdentity := &authsession.Identity{ID: uuid.NewString(), Email: "LEGACY@example.com", EmailVerified: true}
	if _, err := svc.ResolveIdentity(ctx, legacyIdentity); !errors.Is(err, ErrIdentityConflict) {
		t.Fatalf("email must not auto-link: %v", err)
	}
	if err := db.Model(&legacy).Update("auth_user_id", legacyIdentity.ID).Error; err != nil {
		t.Fatal(err)
	}
	mapped, err := svc.ResolveIdentity(ctx, legacyIdentity)
	if err != nil || mapped.ID != legacy.ID || mapped.Theme != models.ThemeDark {
		t.Fatalf("legacy mapping: %v %v", mapped, err)
	}
	var preserved models.Account
	db.First(&preserved, "id = ?", account.ID)
	if preserved.UserID != mapped.ID || preserved.InitialBalance != 42 {
		t.Fatal("financial record changed")
	}
	if err := db.Model(&legacy).Update("user_status", 0).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ResolveIdentity(ctx, legacyIdentity); !errors.Is(err, ErrUserBlocked) {
		t.Fatalf("blocked user: %v", err)
	}
	if err := db.Delete(&legacy).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ResolveIdentity(ctx, legacyIdentity); !errors.Is(err, ErrUserBlocked) {
		t.Fatalf("deleted user: %v", err)
	}
	if _, err := svc.ResolveIdentity(ctx, &authsession.Identity{ID: uuid.NewString(), Email: "new@example.com"}); !errors.Is(err, authsession.ErrUnauthorized) {
		t.Fatal("unverified identity accepted")
	}
}
