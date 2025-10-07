package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"user-service/internal/config"
	"user-service/internal/models"
	"user-service/internal/rabbitmq"
	"user-service/internal/repositories"
	"user-service/internal/routes"
	"user-service/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/whotterre/entritts/pkg/database"
	rabbit "github.com/whotterre/entritts/pkg/rabbitmq"
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

	// Initialize repositories and services
	userRepo := repositories.NewUserRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	userService := services.NewUserService(userRepo, sessionRepo)

	// Setup RabbitMQ consumer
	logger.Info("Initializing RabbitMQ consumer...")

	rabbitConfig := rabbit.Config{URL: cfg.RabbitMQURL}
	rabbitClient, err := rabbit.NewClient(rabbitConfig, log.New(os.Stdout, "rabbitmq: ", log.LstdFlags))
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ", zap.Error(err))
		log.Fatal("RabbitMQ connection failed")
	}
	defer rabbitClient.Close(log.New(os.Stdout, "rabbitmq: ", log.LstdFlags))

	// Create consumer and event handler
	consumer := rabbit.NewConsumer(rabbitClient)
	eventConsumer := rabbitmq.NewUserEventConsumer(userService, logger)

	// Setup queue and bindings
	queueName := "user_events_queue"

	// Ensure the events exchange exists
	err = rabbitClient.EnsureExchange("events", "topic")
	if err != nil {
		logger.Error("Failed to declare exchange", zap.Error(err))
		log.Fatal("Exchange declaration failed")
	}

	_, err = rabbitClient.EnsureQueue(queueName)
	if err != nil {
		logger.Error("Failed to declare queue", zap.Error(err))
		log.Fatal("Queue declaration failed")
	}

	// Bind queue to events exchange with routing keys
	routingKeys := []string{"event.created", "event.updated", "event.deleted"}
	for _, routingKey := range routingKeys {
		err = rabbitClient.BindQueue(queueName, routingKey, "events")
		if err != nil {
			logger.Error("Failed to bind queue",
				zap.String("routing_key", routingKey),
				zap.Error(err))
			log.Fatal("Queue binding failed")
		}
	}

	// Start consuming events
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumeOpts := rabbit.ConsumeOptions{
		QueueName:     queueName,
		ConsumerTag:   "user-service-consumer",
		AutoAck:       false, // Manual acknowledgment for reliability
		Exclusive:     false,
		NoLocal:       false,
		NoWait:        false,
		PrefetchCount: 10,
		PrefetchSize:  0,
	}

	err = consumer.Consume(ctx, consumeOpts, eventConsumer.HandleEventMessage)
	if err != nil {
		logger.Error("Failed to start consumer", zap.Error(err))
		log.Fatal("Consumer startup failed")
	}

	logger.Info("RabbitMQ consumer started successfully")

	// Setup routes
	routes.SetupRoutes(app, db, cfg.PasetoSecret, logger)

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("Shutting down gracefully...")
		cancel() // Cancel consumer context
		app.Shutdown()
	}()

	// Start server
	log.Printf("Starting %s on port %s", cfg.ServiceName, cfg.ServerPort)
	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		logger.Error("Server failed to start", zap.Error(err))
	}
}
