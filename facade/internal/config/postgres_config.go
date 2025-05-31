package config

import (
    "fmt"
    "os"
)

type PostgresConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    DBName   string
    SSLMode  string
}

func LoadPostgresConfigFromEnv() PostgresConfig {
    return PostgresConfig{
        Host:     os.Getenv("POSTGRES_HOST"),
        Port:     getEnvAsInt("POSTGRES_PORT", 5432),
        User:     os.Getenv("POSTGRES_USER"),
        Password: os.Getenv("POSTGRES_PASSWORD"),
        DBName:   os.Getenv("POSTGRES_DB"),
        SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
    }
}

func (cfg PostgresConfig) DSN() string {
    return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
}

// Вспомогательные функции
func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}

func getEnvAsInt(key string, defaultVal int) int {
    if valStr := os.Getenv(key); valStr != "" {
        var val int
        fmt.Sscanf(valStr, "%d", &val)
        return val
    }
    return defaultVal
}
