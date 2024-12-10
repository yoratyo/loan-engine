package config

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	SendgridAPIKey     string
	EmailSenderName    string
	EmailSenderAddress string
	DatabaseURL        string
	AuthUsername       string
	AuthPassword       string
}

var (
	// singleton instance
	configInstance *Config
	// ensure thread-safe lazy initialization
	once sync.Once
)

// LoadConfig initializes the configuration once and returns the instance
func LoadConfig() *Config {
	once.Do(func() {
		// Load .env file
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using OS environment variables or defaults")
		}

		// Initialize the configuration
		// Default value is just for testing purpose
		configInstance = &Config{
			SendgridAPIKey:     getEnv("SENDGRID_API_KEY", ""),
			EmailSenderName:    getEnv("EMAIL_SENDER_NAME", "name"),
			EmailSenderAddress: getEnv("EMAIL_SENDER_ADDRESS", "mail@gmail.com"),
			DatabaseURL:        getEnv("DB_URL", "postgres://user:pass@localhost:5432/loan?sslmode=disable"),
			AuthUsername:       getEnv("AUTH_USERNAME", "user"),
			AuthPassword:       getEnv("AUTH_PASSWORD", "123456"),
		}
	})

	return configInstance
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
