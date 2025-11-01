package configs

import (
	"fmt"
	"os"
)

// DatabaseConfig func to get database configuration.
type DatabaseConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

// GetDatabaseConfig func to get database configuration from environment variables.
func GetDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		Username: getEnv("DB_USERNAME", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "keepnode"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}
}

// GetDatabaseDSN func to get database connection DSN for GORM.
func GetDatabaseDSN() string {
	config := GetDatabaseConfig()
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		config.DBName,
		config.SSLMode,
	)
}

// getEnv func to get environment variable or default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
