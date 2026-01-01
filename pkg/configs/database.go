package configs

import (
	"fmt"
	"os"
	"strings"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
	RawURL   string
}

func GetDatabaseConfig() DatabaseConfig {
	rawURL := getEnv("DATABASE_URL", "")

	return DatabaseConfig{
		RawURL:   rawURL,
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		Username: getEnvMultiple([]string{"DB_USER", "DB_USERNAME"}, "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "keepnode"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}
}

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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvMultiple(keys []string, defaultValue string) string {
	for _, key := range keys {
		if value, exists := os.LookupEnv(key); exists && value != "" {
			return strings.TrimSpace(value)
		}
	}
	return defaultValue
}
