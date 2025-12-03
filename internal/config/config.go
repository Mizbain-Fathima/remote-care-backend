package config

import (
	"os"
	"time"
)

type Config struct {
	GRPCPort      string
	HTTPPort      string
	PostgresURL   string
	RedisAddr     string
	RabbitURL     string
	ServiceName   string
	Env           string
	ShutdownGrace time.Duration
}

func Load() *Config {
	return &Config{
		GRPCPort:      getEnv("GRPC_PORT", ":50051"),
		HTTPPort:      getEnv("HTTP_PORT", ":8080"),
		PostgresURL:   getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/voucher?sslmode=disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RabbitURL:     getEnv("RABBIT_URL", "amqp://guest:guest@localhost:5672/"),
		ServiceName:   getEnv("SERVICE_NAME", "voucher-payment-service"),
		Env:           getEnv("ENV", "dev"),
		ShutdownGrace: 10 * time.Second,
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
