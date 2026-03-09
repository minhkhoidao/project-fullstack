package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	DatabaseURL       string
	RedisURL          string
	KafkaBrokers      string
	JWTSecret         string
	JWTExpiry         time.Duration
	RefreshExpiry     time.Duration
	LogLevel          string
	ServiceName       string
	HTTPAddr          string
	ShutdownTimeout   time.Duration
}

func Load(serviceName, defaultAddr string) (*Config, error) {
	jwtExpiry, err := time.ParseDuration(getEnv("JWT_EXPIRY", "24h"))
	if err != nil {
		return nil, fmt.Errorf("parse JWT_EXPIRY: %w", err)
	}

	refreshExpiry, err := time.ParseDuration(getEnv("REFRESH_TOKEN_EXPIRY", "168h"))
	if err != nil {
		return nil, fmt.Errorf("parse REFRESH_TOKEN_EXPIRY: %w", err)
	}

	return &Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://fashion:fashion_secret@localhost:5432/fashion_ecommerce?sslmode=disable"),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379/0"),
		KafkaBrokers:    getEnv("KAFKA_BROKERS", "localhost:9092"),
		JWTSecret:       getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpiry:       jwtExpiry,
		RefreshExpiry:   refreshExpiry,
		LogLevel:        getEnv("LOG_LEVEL", "debug"),
		ServiceName:     serviceName,
		HTTPAddr:        getEnv("HTTP_ADDR", defaultAddr),
		ShutdownTimeout: 15 * time.Second,
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
