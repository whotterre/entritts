package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GatewayPort             string
	JWTSecret               string
	UserServiceHost         string
	EventServiceHost        string
	OrderServiceHost        string
	TicketServiceHost       string
	NotificationServiceHost string
	PaymentServiceHost      string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		GatewayPort:             getEnv("GATEWAY_PORT", "3000"),
		JWTSecret:               getEnv("JWT_SECRET", ""),
		UserServiceHost:         getEnv("USER_SERVICE_HOST", "user-service"),
		EventServiceHost:        getEnv("EVENT_SERVICE_HOST", "event-service"),
		OrderServiceHost:        getEnv("ORDER_SERVICE_HOST", "order-service"),
		TicketServiceHost:       getEnv("TICKET_SERVICE_HOST", "ticket-service"),
		NotificationServiceHost: getEnv("NOTIF_SERVICE_HOST", "notif-service"),
		PaymentServiceHost: 	 getEnv("PAYMENT_SERVICE_HOST", "payment-service"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
