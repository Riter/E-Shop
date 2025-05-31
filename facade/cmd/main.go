package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/Riter/E-Shop/internal/config"
	"github.com/Riter/E-Shop/internal/handlers"
	psq "github.com/Riter/E-Shop/internal/storage/postgres"
	"github.com/Riter/E-Shop/internal/storage/redis"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	_, err := rdb.Ping(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Successfully connected to Redis")

    cfg := config.LoadPostgresConfigFromEnv()
    dbClient, err := psq.NewPostgres(ctx, cfg)
    if err != nil {
        log.Fatalf("Ошибка подключения к БД: %v", err)
    }


	go func() {
		http.Handle("/metrics", promhttp.Handler())
			if err := http.ListenAndServe(":10666", nil); err != nil {
				slog.Error("failed to start metrics server", slog.Any("err", err))
			}
  	}()

	r := chi.NewRouter()
	handlers.GetProducts(ctx, dbClient, rdb)

	r.Post("/products", handlers.GetProducts(ctx, dbClient, rdb))

	// Запуск сервера
	log.Println("Listening on :8089")
	if err := http.ListenAndServe(":8089", r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
	log.Println("Server stopped")


}