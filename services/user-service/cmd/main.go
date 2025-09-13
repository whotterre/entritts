package main

import (
	"log"
	"user-service/internal/config"
	"user-service/internal/models"
	"user-service/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/whotterre/entritts/pkg/database"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Populate db config
	dbConfig := database.DBConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		SSLMode:  cfg.SSLMode,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
	}

	// Create new Fiber instance
	app := fiber.New()
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Initialize the database
	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Println("Failed to create a new PostgreSQL database connection", err)
	}
	// Auto-migrate database models
	if err := db.AutoMigrate(&models.User{}, &models.UserRole{}, &models.UserSession{}); err != nil {
		log.Printf("Failed to migrate database: %v", err)
		return
	}
	log.Println("Database migration completed successfully")

	// Setup routes
	routes.SetupRoutes(app, db, cfg.JwtSecret, logger)

	// Start server
	log.Printf("Starting %s  on port %s", cfg.ServiceName, cfg.ServerPort)
	log.Fatal(app.Listen(":" + cfg.ServerPort))
}
