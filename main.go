package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"keepnode/pkg/routes"
	"keepnode/platform/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Initialize database connection (singleton)
	if err := database.Init(); err != nil {
		log.Fatalf("Database initialization error: %v", err)
	}
	log.Println("Database initialized successfully")

	// Run migrations
	db := database.GetDB()
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Migration error: %v", err)
	}
	log.Println("Migrations completed successfully")

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:        "KeepNode API",
		ReadBufferSize: 16384, // 16KB - prevents 431 error for large headers/cookies
	})

	// CORS middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "http://localhost:8080")
		c.Set("Access-Control-Allow-Credentials", "true")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	})

	// Recovery middleware
	app.Use(recover.New())

	// Register routes
	routes.Register(app)

	// Graceful shutdown
	go func() {
		if err := app.Listen(":3000"); err != nil {
			log.Printf("Fiber server stopped: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Graceful shutdown failed: %v", err)
	}

	// Close database connection
	if err := database.Close(); err != nil {
		log.Printf("Failed to close database connection: %v", err)
	}

	log.Println("Server gracefully stopped")
}
