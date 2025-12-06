package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	GRPCPort      string
	HTTPPort      string
	PostgresURL   string
	RedisURL      string
	RedisAddr     string
	RedisPassword string
	RedisTLS      bool
	RabbitURL     string
	ServiceName   string
	Env           string
	ShutdownGrace time.Duration
}

func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	val = strings.ToLower(val)
	return val == "1" || val == "true" || val == "yes"
}

func Load() *Config {
	return &Config{
		GRPCPort:      getEnv("GRPC_PORT", ":50051"),
		HTTPPort:      getEnv("HTTP_PORT", ":8080"),
		PostgresURL:   getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/voucher?sslmode=disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RabbitURL:     getEnv("RABBIT_URL", "amqp://guest:guest@localhost:5672/"),
		RedisURL:      getEnv("REDIS_URL", ""),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisTLS:      getEnvBool("REDIS_TLS", false),
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
