package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"keepnode/pkg/apiresponse"
	"keepnode/pkg/middleware"
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

	// DDL credentials belong to the one-shot migration job, never the HTTP server.
	if len(os.Args) > 1 {
		if len(os.Args) != 2 || os.Args[1] != "migrate" {
			log.Fatal("Usage: keepnode [migrate]")
		}
		if err := database.AutoMigrate(database.GetDB()); err != nil {
			log.Fatalf("Migration error: %v", err)
		}
		if err := database.Close(); err != nil {
			log.Fatal(err)
		}
		log.Println("Migrations completed successfully")
		return
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:        "KeepNode API",
		ReadBufferSize: 16384, // 16KB - prevents 431 error for large headers/cookies
		ErrorHandler:   apiresponse.ErrorHandler,
	})

	// Recovery middleware
	app.Use(middleware.RequestContext(), recover.New())

	// Register routes
	if err := routes.Register(app); err != nil {
		log.Fatalf("Auth configuration error: %v", err)
	}

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
