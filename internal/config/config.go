package config

import (
	"log"
	"os"
)

// Config holds application configuration loaded from environment variables.
// For brevity only a subset of the variables is included.
type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	GoogleClientID string
	GoogleSecret   string
	GoogleRedirect string
	AWSRegion      string
	SQSQueueURL    string
	LogLevel       string
}

// Load reads environment variables into Config.
func Load() *Config {
	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://user:pass@localhost:5432/db"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		GoogleClientID: os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleSecret:   os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirect: os.Getenv("GOOGLE_REDIRECT_URL"),
		AWSRegion:      getEnv("AWS_REGION", "us-east-1"),
		SQSQueueURL:    os.Getenv("AWS_SQS_QUEUE_URL"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}
	if cfg.JWTSecret == "" {
		log.Println("warning: JWT_SECRET not set")
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
