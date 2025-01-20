package config

import (
	"os"
	"strings"
)

type Config struct {
	// CORS Configuration
	CorsAllowedOrigins []string
	CorsAllowedMethods []string
	CorsAllowedHeaders []string

	// Redis Configuration
	RedisHost     string
	RedisPort     string
	RedisPassword string

	// File Path
	CsvPath string
}

func LoadConfig() *Config {
	return &Config{
		CorsAllowedOrigins: strings.Split(getEnvOrDefault("CORS_ALLOWED_ORIGINS", "*"), ","),
		CorsAllowedMethods: strings.Split(getEnvOrDefault("CORS_ALLOWED_METHODS", "GET,POST,DELETE"), ","),
		CorsAllowedHeaders: strings.Split(getEnvOrDefault("CORS_ALLOWED_HEADERS", "Accept,Authorization,Content-Type"), ","),

		RedisHost:     getEnvOrDefault("REDIS_HOST", "redis"),
		RedisPort:     getEnvOrDefault("REDIS_PORT", "6379"),
		RedisPassword: getEnvOrDefault("REDIS_PASSWORD", ""),

		CsvPath: getEnvOrDefault("CV_PATH", "SWIFT_CODES.csv"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
