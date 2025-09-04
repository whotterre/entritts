package main

import (
	"log"
	"user-service/internal/config"
	"user-service/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Create new Fiber instance
	app := fiber.New()

	// Middleware
	app.Use(logger.New()) // Add logger middleware

	// Setup routes
	routes.SetupRoutes(app)

	// Start server
	log.Printf("Starting %s  on port %s", cfg.ServiceName, cfg.ServerPort)
	log.Fatal(app.Listen(":" + cfg.ServerPort))
}
