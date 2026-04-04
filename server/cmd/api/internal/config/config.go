package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret      string
	JWTExpiration  time.Duration
	ServerPort     string
	DatabaseURL    string
	AnthropicAPIKey string
	GeminiAPIKey    string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	}

	return &Config{
		JWTSecret:       getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
		JWTExpiration:   time.Hour * 24 * 7,
		ServerPort:      getEnv("SERVER_PORT", ":8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/bookkeeping?sslmode=disable"),
		AnthropicAPIKey: getEnv("ANTHROPIC_API_KEY", ""),
		GeminiAPIKey:    getEnv("GEMINI_API_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
