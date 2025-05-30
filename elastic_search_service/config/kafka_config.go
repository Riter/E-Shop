package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type KafkaConfig struct {
	Brokers []string
	GroupID string
	Topic   string
}

func LoadKafkaConfig() *KafkaConfig {
	err := godotenv.Load("../environment/kafka.env")
	if err != nil {
		log.Fatalf("ошибка при загрузке kafka.env файла: %v", err)
	}

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 {
		log.Fatal("KAFKA_BROKERS не указан")
	}

	return &KafkaConfig{
		Brokers: brokers,
		GroupID: os.Getenv("KAFKA_GROUP_ID"),
		Topic:   os.Getenv("KAFKA_TOPIC"),
	}
}
