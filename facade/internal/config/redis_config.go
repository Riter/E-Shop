package config

import (
	"log"
	"strconv"
)

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func LoadRedisConfig() *RedisConfig {
	db, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		log.Printf("Invalid REDIS_DB value, using default 0: %v", err)
		db = 0
	}

	return &RedisConfig{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       db,
	}
}
