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
	app := fiber.New()

	// Recovery middleware
	app.Use(recover.New())

	db, err := database.OpenDBConnection()
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	err = database.AutoMigrate(db)
	if err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	routes.Register(app)

	// Graceful shutdown
	go func() {
		if err := app.Listen(":3000"); err != nil {
			log.Printf("Fiber server stopped: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Graceful shutdown failed: %v", err)
	}
	log.Println("Server gracefully stopped")
}
