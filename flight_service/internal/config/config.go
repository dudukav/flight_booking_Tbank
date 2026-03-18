package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresDSN string
	RedisAddr   string
	APIKey      string
	GRPCPort    string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(".env not found, using system env")
	}

	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("POSTGRES_DSN is not set")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		return nil, fmt.Errorf("REDIS_ADDR is not set")
	}

	apiKey := os.Getenv("FLIGHT_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("FLIGHT_API_KEY is not set")
	}

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	return &Config{
		PostgresDSN: dsn,
		RedisAddr:   redisAddr,
		APIKey:      apiKey,
		GRPCPort:    port,
	}, nil
}