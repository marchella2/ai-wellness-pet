package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	Port         string
	DatabaseURL  string
	GeminiAPIKey string
}

func LoadEnv() *Env {
	if err := godotenv.Load(); err != nil {
		log.Println("[config] .env file not found, falling back to system environment variables")
	}

	env := &Env{
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
	}

	if env.DatabaseURL == "" {
		log.Fatal("[config] DATABASE_URL is required")
	}

	return env
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
