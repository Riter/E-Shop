package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBMaxConnLifeTime int

	DBMinioRootUser  string
	DBMinioRootPassw string
	DBMinioBucket    string
	DBMinioEndpoint  string
}

func LoadConfig() *Config {
	err := godotenv.Load("../environment/.env")
	if err != nil {
		log.Fatalf("ошибка при получении .env файла для баз данных: %v", err)
	}

	MaxOpenConns, err := strconv.Atoi(os.Getenv("POSTGRES_MAX_OPEN_CONNS"))
	if err != nil {
		panic("ошибка при считывании POSTGRES_MAX_OPEN_CONNS")
	}
	MaxIdleConns, err := strconv.Atoi(os.Getenv("POSTGRES_MAX_IDLE_CONNS"))
	if err != nil {
		panic("ошибка при считывании POSTGRES_MAX_IDLE_CONNS")
	}
	MaxConnLifeTime, err := strconv.Atoi(os.Getenv("POSTGRES_CONN_MAX_LIFETIME"))
	if err != nil {
		panic("ошибка при считывании POSTGRES_CONN_MAX_LIFETIME")
	}

	return &Config{
		DBHost:            os.Getenv("POSTGRES_HOST"),
		DBPort:            os.Getenv("POSTGRES_PORT"),
		DBUser:            os.Getenv("POSTGRES_USER"),
		DBPassword:        os.Getenv("POSTGRES_PASSWORD"),
		DBName:            os.Getenv("POSTGRES_NAME"),
		DBSSLMode:         os.Getenv("POSTGRES_SSLMODE"),
		DBMaxOpenConns:    MaxOpenConns,
		DBMaxIdleConns:    MaxIdleConns,
		DBMaxConnLifeTime: MaxConnLifeTime,

		DBMinioRootUser:  os.Getenv("MINIO_ROOT_USER"),
		DBMinioRootPassw: os.Getenv("MINIO_ROOT_PASSWORD"),
		DBMinioBucket:    os.Getenv("MINIO_BUCKET"),
		DBMinioEndpoint:  os.Getenv("MINIO_ENDPOINT"),
	}

}
