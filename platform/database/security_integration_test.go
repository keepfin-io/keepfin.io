package database

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	accountmodel "keepnode/app/models/account"
	usermodel "keepnode/app/models/user"
	userquery "keepnode/app/queries/user"
	userservice "keepnode/app/services/user"
	"keepnode/pkg/authsession"
)

func TestRestrictedRuntimePreservesUserAndAccountOperations(t *testing.T) {
	dsn := os.Getenv("SECURITY_RUNTIME_DATABASE_URL")
	if dsn == "" {
		t.Skip("run scripts/test-db-security.sh")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	ctx := context.Background()
	identity := &authsession.Identity{ID: uuid.NewString(), Email: uuid.NewString() + "@security.test", EmailVerified: true}
	service := userservice.NewUserService(db)
	user, err := service.ResolveIdentity(ctx, identity)
	if err != nil {
		t.Fatal(err)
	}
	user.Theme = "dark"
	if err := (&userquery.UserQueries{DB: db}).UpdateUser(ctx, user.ID, user); err != nil {
		t.Fatal(err)
	}
	var persisted usermodel.User
	if err := db.First(&persisted, "id = ?", user.ID).Error; err != nil {
		t.Fatal(err)
	}
	if persisted.Theme != "dark" || persisted.AuthUserID == nil || *persisted.AuthUserID != identity.ID {
		t.Fatal("profile changed identity or did not persist")
	}
	account := accountmodel.Account{UserID: user.ID, AccountName: "Security integration", Currency: "USD"}
	if err := db.Create(&account).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&account).Update("account_name", "Updated account").Error; err != nil {
		t.Fatal(err)
	}
	var dirty int64
	if err := db.Table("holding_search_states").Where("account_id = ?", account.ID).Count(&dirty).Error; err != nil || dirty != 1 {
		t.Fatalf("outbox trigger failed: %d %v", dirty, err)
	}
	if err := db.Delete(&account).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE TABLE public.runtime_ddl_forbidden (id int)").Error; err == nil {
		t.Fatal("runtime could perform DDL")
	}
}
