package main

import (
	"log"
	"strconv"
	"ticket-service/internal/config"
	"ticket-service/internal/rabbitmq"
	"ticket-service/internal/repository"
	"ticket-service/internal/routes"
	"ticket-service/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/rabbitmq/amqp091-go"
	"github.com/whotterre/entritts/pkg/database"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	app := fiber.New()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Printf("Failed to initialize new Zap logger for ticket service %v", err)
		return
	}

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
	// Connect to RabbitMQ
	conn, err := amqp091.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Println("Failed to connect to RabbitMQ", err)
	}

	// Middleware
	app.Use(logger)

	// Create producer
	producer := rabbitmq.NewTicketProducer(conn, logger)
	// Create initial consumer with nil service to break circular dependency; service will be attached after it's created.
	consumer := rabbitmq.NewTicketConsumer(conn, logger, nil)
	// Create repository and service
	ticketRepo := repository.NewTicketRepository(db, logger)
	ticketService := services.NewTicketService(ticketRepo, producer, consumer, logger)
	ticketService.SetConsumer(consumer)

	// Setup routes with the constructed service and consumer
	routes.SetupRoutes(app, db, ticketService, consumer, producer, logger)

	// Start server
	log.Printf("Starting %s server on port %s", cfg.AppName, cfg.ServerPort)
	log.Fatal(app.Listen(":" + cfg.ServerPort))
}
