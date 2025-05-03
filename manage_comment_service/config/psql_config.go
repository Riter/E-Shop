package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type PsqlConfig struct {
	DBUser            string
	DBPassword        string
	DBName            string
	DBHost            string
	DBPort            string
	DBSSLMode         string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBMaxConnLifeTime int
}

func LoadConfig() *PsqlConfig {
	err := godotenv.Load("../environment/psql.env")
	if err != nil {
		log.Fatalf("ошибка при считывании .env файла: %v", err)
	}

	maxOpenConns, err := strconv.Atoi(os.Getenv("POSTGRES_MAX_OPEN_CONNS"))
	if err != nil {
		log.Fatalf("ошибка при парсинге POSTGRES_MAX_OPEN_CONNS: %v", err)
	}

	maxIdleConns, err := strconv.Atoi(os.Getenv("POSTGRES_MAX_IDLE_CONNS"))
	if err != nil {
		log.Fatalf("ошибка при парсинге POSTGRES_MAX_IDLE_CONNS: %v", err)
	}

	maxConnLifeTime, err := strconv.Atoi(os.Getenv("POSTGRES_CONN_MAX_LIFETIME"))
	if err != nil {
		log.Fatalf("ошибка при парсинге POSTGRES_CONN_MAX_LIFETIME: %v", err)
	}

	return &PsqlConfig{
		DBUser:            os.Getenv("POSTGRES_USER"),
		DBPassword:        os.Getenv("POSTGRES_PASSWORD"),
		DBName:            os.Getenv("POSTGRES_DB"),
		DBHost:            os.Getenv("POSTGRES_HOST"),
		DBPort:            os.Getenv("POSTGRES_PORT"),
		DBSSLMode:         os.Getenv("POSTGRES_SSLMODE"),
		DBMaxOpenConns:    maxOpenConns,
		DBMaxIdleConns:    maxIdleConns,
		DBMaxConnLifeTime: maxConnLifeTime,
	}
}
