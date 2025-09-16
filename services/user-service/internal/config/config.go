package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServiceName string
	ServerPort  string
	DBHost      string
	DBPort      int
	DBUser      string
	DBPassword  string
	DBName      string
	PasetoSecret   string
	SSLMode     string
	RabbitMQURL string
}

func LoadConfig() *Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	return &Config{
		ServiceName: getEnv("SERVICE_NAME", "user-service"),
		ServerPort:  getEnv("SERVER_PORT", "3000"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnvAsInt("DB_PORT", 5432),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "password"),
		DBName:      getEnv("DB_NAME", "user_db"),
		PasetoSecret:   getEnv("PASETO_SECRET", "0nb4W4x--rC60r9bDPiIAbcyXHTsRQ_"),
		SSLMode:     getEnv("SSL_MODE", "disable"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}
}
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
