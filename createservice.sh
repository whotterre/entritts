#!/bin/bash

# This script creates a new Go/Fiber microservice with a standard structure.
# Usage: ./create-service.sh <service_name>

# If no name is passed, use 'newservice' as the default
service_name=${1:-newservice}

echo "Creating new service: $service_name..."
if [ -d "services" ]; then 
   cd services
else 
   mkdir services
   cd services
fi

# Check if a directory already exists with that name
if [ -d "$service_name" ]; then
  echo "Error: A directory named '$service_name' already exists. Aborting."
  exit 1
fi

# Create the root service directory
mkdir "$service_name"
cd "$service_name" || exit 1

# Create the standard Go/Fiber microservice structure
echo "Creating directory structure..."
mkdir -p cmd
mkdir -p internal/config
mkdir -p internal/handlers
mkdir -p internal/models
mkdir -p internal/repository
mkdir -p internal/services
mkdir -p internal/routes
mkdir -p internal/middleware
mkdir -p migrations

echo "Creating core files..."

# Create main.go
cat > cmd/main.go << 'EOF'
package main

import (
	"log"
	"{{.ServiceName}}/internal/config"
	"{{.ServiceName}}/internal/routes"

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
	log.Printf("Starting %s server on port %s", cfg.AppName, cfg.ServerPort)
	log.Fatal(app.Listen(":" + cfg.ServerPort))
}
EOF

# Create config file
cat > internal/config/config.go << 'EOF'
package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName              string
	ServerPort           string
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	RabbitMQURL          string
}

func LoadConfig() *Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	return &Config{
		AppName:              getEnv("APP_NAME", "{{.ServiceName}}"),
		ServerPort:           getEnv("SERVER_PORT", "8080"),
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPassword:           getEnv("DB_PASSWORD", "password"),
		DBName:               getEnv("DB_NAME", "{{.ServiceName}}_db"),
		RabbitMQURL:          getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
EOF

# Create routes file
cat > internal/routes/routes.go << 'EOF'
package routes

import (
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Health check endpoint
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "OK",
			"service": "{{.ServiceName}}",
		})
	})

	// Add your service-specific routes here
	// Example:
	// api.Get("/users", handlers.GetAllUsers)
}
EOF

# Create a sample model
cat > internal/models/sample.go << 'EOF'
package models

import (
	"time"

	"gorm.io/gorm"
)

type Sample struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
EOF

# Create Dockerfile
cat > Dockerfile << 'EOF'
# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main ./cmd/main.go

# Run stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/.env .  # Optional, for local config
EXPOSE 8080
CMD ["./main"]
EOF

# Create .gitignore
cat > .gitignore << 'EOF'
# Binaries
main
*.exe
*.out

# Dependencies
vendor/

# Environment files
.env
.env.local

# IDE
.vscode/
.idea/

# Logs
*.log
logs/

# Coverage
coverage.txt
EOF

# Create a sample .env file
cat > .env.example << 'EOF'
APP_NAME={{.ServiceName}}
SERVER_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME={{.ServiceName}}_db
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
EOF

# Create go.mod file
go mod init "$service_name"

# Add essential dependencies
echo "Adding essential dependencies..."
go get github.com/gofiber/fiber/v2
go get github.com/joho/godotenv
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/streadway/amqp

# Replace placeholder with actual service name in all files
echo "Customizing files for service: $service_name"

# Use sed to replace the placeholder {{.ServiceName}}
if [[ "$OSTYPE" == "darwin"* ]]; then
  # MacOS (BSD sed)
  sed -i '' "s/{{\.ServiceName}}/$service_name/g" cmd/main.go
  sed -i '' "s/{{\.ServiceName}}/$service_name/g" internal/config/config.go
  sed -i '' "s/{{\.ServiceName}}/$service_name/g" internal/routes/routes.go
  sed -i '' "s/{{\.ServiceName}}/$service_name/g" .env.example
else
  # Linux (GNU sed)
  sed -i "s/{{\.ServiceName}}/$service_name/g" cmd/main.go
  sed -i "s/{{\.ServiceName}}/$service_name/g" internal/config/config.go
  sed -i "s/{{\.ServiceName}}/$service_name/g" internal/routes/routes.go
  sed -i "s/{{\.ServiceName}}/$service_name/g" .env.example
fi

echo "Service '$service_name' created successfully!"
echo "Next steps:"
echo "1. cd $service_name"
echo "2. Copy .env.example to .env and configure your settings"
echo "3. Start implementing your domain logic in internal/"
echo "4. Run: go run ./cmd/main.go"