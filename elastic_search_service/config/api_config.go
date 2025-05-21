package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type APIConfig struct {
	APIPort uint
}

func LoadAPIConfig() *APIConfig {
	err := godotenv.Load("../environment/api.env")
	if err != nil {
		log.Fatalf("ошибка при загрузке api.env файла: %v", err)
	}

	portStr := os.Getenv("APP_PORT")
	port, err := strconv.ParseUint(portStr, 10, 32)
	if err != nil {
		log.Fatalf("неверный формат порта: %v", err)
	}

	return &APIConfig{
		APIPort: uint(port),
	}
}
