package config

import (
	"log"

	"github.com/joho/godotenv"
)

type PsqlCongfig struct {
	DBUser            string
	DBPassword        string
	DBName            string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBMaxConnLifeTime int
}

func LoadConfig() *PsqlCongfig {
	err := godotenv.Load("../environment/psql.env")
	if err != nil {
		log.Fatalf("ошибка при считывании файла окружения для сервиса комментариев: %v", err)
	}
}
