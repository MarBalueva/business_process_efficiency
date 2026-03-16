package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	JWTSecret  string
	JWTTTL     time.Duration
	EmbeddingEnabled bool
	EmbeddingURL     string
	EmbeddingTimeout time.Duration
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env файл не найден, используются переменные окружения системы")
	}

	embeddingTimeoutSec := getEnvAsInt("EMBEDDING_TIMEOUT_SEC", 120)
	if embeddingTimeoutSec <= 0 {
		embeddingTimeoutSec = 120
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "postgres"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		JWTSecret:  getEnv("JWT_SECRET", ""),
		JWTTTL:     time.Hour * 24,
		EmbeddingEnabled: getEnv("EMBEDDING_ENABLED", "false") == "true",
		EmbeddingURL:     getEnv("EMBEDDING_URL", "http://localhost:8001"),
		EmbeddingTimeout: time.Duration(embeddingTimeoutSec) * time.Second,
	}
}

func getEnv(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}
