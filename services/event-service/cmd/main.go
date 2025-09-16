package main

import (
	"event-service/internal/config"
	"event-service/internal/models"
	"event-service/internal/rabbitmq"
	"event-service/internal/routes"
	"log"
	"strconv"

	"github.com/whotterre/entritts/pkg/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Create new Fiber instance
	app := fiber.New()

	// Middleware
	app.Use(logger.New())

	// Logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Println("Failed to initialize Zap logger", err)
	}

	cfg = config.LoadConfig()

	// Populate db config
	dbPort, _ := strconv.Atoi(cfg.DBPort)
	dbConfig := database.DBConfig{
		Host:     cfg.DBHost,
		Port:     dbPort,
		SSLMode:  cfg.SSLMode,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
	}

	// Connect to Postgres DB
	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		logger.Error("Failed to initialize PostgreSQL connection for event_db", zap.Error(err))
		return
	}

	if err := db.AutoMigrate(&models.EventCategory{}); err != nil {
		logger.Error("Failed to migrate EventCategory", zap.Error(err))
		return
	}

	if err := db.AutoMigrate(&models.EventVenue{}); err != nil {
		logger.Error("Failed to migrate EventVenue", zap.Error(err))
		return
	}

	if err := db.AutoMigrate(&models.Event{}); err != nil {
		logger.Error("Failed to migrate Event", zap.Error(err))
		return
	}

	if err := db.AutoMigrate(&models.EventParticipant{}); err != nil {
		logger.Error("Failed to migrate EventParticipant", zap.Error(err))
		return
	}

	logger.Info("Database migration completed successfully")
	// Initialize RabbitMQ producer in the service
	producer, err := rabbitmq.NewEventProducer(cfg.RabbitMQURL, logger)
	if err != nil {
		logger.Error("Failed to initialize RabbitMQ broker", zap.Error(err))
		return
	}
	// Setup routes
	routes.SetupRoutes(app, db, logger, producer)

	// Start server
	log.Printf("Starting %s server on port %s", cfg.AppName, cfg.ServerPort)
	log.Fatal(app.Listen(":" + cfg.ServerPort))
}
