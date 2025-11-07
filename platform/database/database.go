package database

import (
	"log"
	"sync"

	"gorm.io/gorm"
)

var (
	// DB is the global database instance
	DB   *gorm.DB
	once sync.Once
)

// Init initializes the database connection (singleton pattern)
func Init() error {
	var err error
	once.Do(func() {
		DB, err = OpenDBConnection()
		if err != nil {
			log.Printf("Failed to open database connection: %v", err)
			return
		}
		log.Println("Database connection established successfully")
	})
	return err
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	if DB == nil {
		log.Fatal("Database not initialized. Call Init() first")
	}
	return DB
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
