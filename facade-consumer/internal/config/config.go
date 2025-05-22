package config

import (
    "log"
    "os"
    "strconv"
    "strings"
)

type Config struct {
    KafkaBrokers []string
    KafkaTopic   string
    KafkaGroupID string

    RedisAddr     string
    RedisPassword string
    RedisDB       int
}

func LoadConfig() *Config {
    redisDB := 0
    if val := os.Getenv("REDIS_DB"); val != "" {
        var err error
        redisDB, err = strconv.Atoi(val)
        if err != nil {
            log.Fatalf("Invalid REDIS_DB value: %v", err)
        }
    }

    return &Config{
        KafkaBrokers: strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
        KafkaTopic:   os.Getenv("KAFKA_TOPIC"),
        KafkaGroupID: os.Getenv("KAFKA_GROUP_ID"),

        RedisAddr:     os.Getenv("REDIS_ADDR"),
        RedisPassword: os.Getenv("REDIS_PASS"),
        RedisDB:       redisDB,
    }
}
