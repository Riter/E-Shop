package config

import (
	"os"
)

type PsqlConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
	SSLMode  string
}

func LoadPostgresConfig() *PsqlConfig {
	return &PsqlConfig{
		User:     getEnv("POSTGRES_USER", "postgres"),
		Password: getEnv("POSTGRES_PASSWORD", ""),
		Host:     getEnv("POSTGRES_HOST", "localhost"),
		Port:     getEnv("POSTGRES_PORT", "5432"),
		DBName:   getEnv("POSTGRES_DB", "postgres"),
		SSLMode:  getEnv("POSTGRES_SSL", "disable"),
	}
}

func getEnv(key string, defaultval string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return defaultval
}
