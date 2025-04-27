package main

import (
	"context"
	"fmt"

	"github.com/Riter/E-Shop/internal/config"
	redisclient "github.com/Riter/E-Shop/internal/storage/redis"
	"github.com/redis/go-redis/v9"
)




func main() {
	// Загружаем конфигурацию из переменных окружения
	ctx := context.Background()
	cfg := config.LoadRedisConfig()
    rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
    client := redisclient.NewRedisClient(rdb)
	err := client.Ping(ctx)
	if err !=nil{
		panic(fmt.Sprintf("Can't connect to Redis %s", err.Error()))
	}

    // Установить значение
    if err := client.Set("mykey", "Hello from go-redis!"); err != nil {
        panic(err)
    }

    // Получить значение
    value, err := client.Get("mykey")
    if err != nil {
        panic(err)
    }

    fmt.Println("Got value:", value)
}