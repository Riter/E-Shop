package main

import (
	"context"
	"log"
	"time"

	"github.com/Riter/E-Shop/internal/config"
	"github.com/Riter/E-Shop/internal/storage/redis"
)




func main() {
	// Загружаем конфигурацию из переменных окружения
	redisCfg := config.LoadRedisConfig()

	// Создаем клиент Redis
	rdb := redis.NewRedisClient(redisCfg)
	defer rdb.Close()

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Successfully connected to Redis")

	// Пример работы с Redis
	err = rdb.Set(ctx, "key", "value", 10*time.Minute).Err()
	if err != nil {
		log.Printf("Failed to set key: %v", err)
	}

	val, err := rdb.Get(ctx, "key").Result()
	if err != nil {
		log.Printf("Failed to get key: %v", err)
	} else {
		log.Printf("Got value from Redis: %s", val)
	}
}