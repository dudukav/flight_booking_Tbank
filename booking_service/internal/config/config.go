package config

import (
    "log"
    "os"
    "time"

    "github.com/joho/godotenv"
)

type Config struct {
    App struct {
        Env     string
        Version string
    }
    HTTP struct {
        Port           string
        ReadTimeout    time.Duration
        WriteTimeout   time.Duration
    }
    Database struct {
        DSN string
    }
    FlightService struct {
        Addr  string
        APIKey string
    }
}

func Load() *Config {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    cfg := &Config{}
 
    cfg.App.Env = getEnv("APP_ENV", "development")
    cfg.App.Version = getEnv("APP_VERSION", "1.0.0")

    cfg.HTTP.Port = getEnv("HTTP_PORT", "8080")
    cfg.HTTP.ReadTimeout = parseDuration(getEnv("HTTP_READ_TIMEOUT", "10s"))
    cfg.HTTP.WriteTimeout = parseDuration(getEnv("HTTP_WRITE_TIMEOUT", "30s"))

    cfg.Database.DSN = getEnv("DB_DSN", "postgres://user:pass@localhost/db?sslmode=disable")

    cfg.FlightService.Addr = getEnv("FLIGHT_SERVICE_ADDR", "localhost:8081")
    cfg.FlightService.APIKey = getEnv("FLIGHT_SERVICE_API_KEY", "dev-key")

    return cfg
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func parseDuration(s string) time.Duration {
    d, err := time.ParseDuration(s)
    if err != nil {
        return 30 * time.Second
    }
    return d
}
