package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret        string
	JWTExpiration    time.Duration
	ServerPort       string
	DatabaseURL      string
	AnthropicAPIKey  string
	GeminiAPIKey     string
	TinkClientID     string
	TinkClientSecret string
	TinkCallbackURL  string
	TinkEncryptionKey string
	StripeSecretKey    string
	StripeWebhookSecret string
	FrontendURL        string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	}

	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" || jwtSecret == "your-secret-key-change-this-in-production" {
		log.Println("WARNING: JWT_SECRET is not set or using default. Set a strong secret in production!")
	}
	if jwtSecret == "" {
		jwtSecret = "dev-only-secret-not-for-production"
	}

	return &Config{
		JWTSecret:       jwtSecret,
		JWTExpiration:   time.Hour * 24 * 7,
		ServerPort:      getEnv("SERVER_PORT", ":8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/bookkeeping?sslmode=disable"),
		AnthropicAPIKey:   getEnv("ANTHROPIC_API_KEY", ""),
		GeminiAPIKey:      getEnv("GEMINI_API_KEY", ""),
		TinkClientID:       getEnv("TINK_CLIENT_ID", ""),
		TinkClientSecret:   getEnv("TINK_CLIENT_SECRET", ""),
		TinkCallbackURL:    getEnv("TINK_CALLBACK_URL", "http://localhost:3000/bank/callback"),
		TinkEncryptionKey:  getEnv("TINK_ENCRYPTION_KEY", ""),
		StripeSecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		FrontendURL:         getEnv("FRONTEND_URL", "http://localhost:3000"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
