package database

import (
	"time"

	"keepnode/pkg/configs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// OpenDBConnection func for opening database connection with GORM.
func OpenDBConnection() (*gorm.DB, error) {
	// Get database connection DSN.
	dsn := configs.GetDatabaseDSN()

	// Define database connection.
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, err
	}

	// Get generic database object sql.DB to set connection pool settings.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Set database connection settings.
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
